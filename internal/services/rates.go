package services

import (
	"sync"
	"time"
)

type RatesFetcher interface {
	Latest(base string) (baseOut string, date time.Time, rates map[string]float64, err error)
}

type RatesService struct {
	F     RatesFetcher
	TTL   time.Duration
	mu    sync.Mutex
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
	s.mu.Lock()
	if s.cache == nil {
		s.cache = map[string]cached{}
	}
	if c, ok := s.cache[base]; ok && time.Since(c.at) < s.TTL {
		s.mu.Unlock()
		return c.base, c.date, c.rates, nil
	}
	s.mu.Unlock()

	b, d, r, err := s.F.Latest(base)
	if err != nil {
		return "", time.Time{}, nil, err
	}

	s.mu.Lock()
	s.cache[base] = cached{base: b, date: d, rates: r, at: time.Now()}
	s.mu.Unlock()
	return b, d, r, nil
}
