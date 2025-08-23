package services

import (
	"time"
)

type RatesFetcher interface {
	Latest(base string) (baseOut string, date time.Time, rates map[string]float64, err error)
}
type RatesStore interface {
	Load(base string) (*CacheRecord, error)
	Save(rec *CacheRecord) error
}
type CacheRecord struct {
	Base    string
	Date    string
	Rates   map[string]float64
	SavedAt time.Time
}

type RatesService struct {
	F        RatesFetcher
	Store    RatesStore
	TTL      time.Duration
	StaleTTL time.Duration

	cache map[string]cached
}
type cached struct {
	base  string
	date  time.Time
	rates map[string]float64
	at    time.Time
}

func (s *RatesService) Latest(base string) (string, time.Time, map[string]float64, error) {
	if base == "" {
		base = "TRY"
	}
	if s.cache == nil {
		s.cache = map[string]cached{}
	}
	if c, ok := s.cache[base]; ok && time.Since(c.at) < s.TTL {
		return c.base, c.date, c.rates, nil
	}

	b, d, r, err := s.F.Latest(base)
	if err == nil {
		s.cache[base] = cached{base: b, date: d, rates: r, at: time.Now()}
		if s.Store != nil {
			_ = s.Store.Save(&CacheRecord{Base: b, Date: d.Format("2006-01-02"), Rates: r, SavedAt: time.Now()})
		}
		return b, d, r, nil
	}

	if s.Store != nil {
		if rec, e := s.Store.Load(base); e == nil {
			dd, _ := time.Parse("2006-01-02", rec.Date)
			if time.Since(rec.SavedAt) <= s.StaleTTL {
				return rec.Base, dd, rec.Rates, nil
			}
		}
	}
	return "", time.Time{}, nil, err
}
