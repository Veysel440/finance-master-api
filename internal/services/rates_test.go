package services

import (
	"errors"
	"testing"
	"time"
)

type fakeFetcher struct {
	base string
	date time.Time
	r    map[string]float64
	err  error
}

func (f *fakeFetcher) Latest(base string) (string, time.Time, map[string]float64, error) {
	if f.err != nil {
		return "", time.Time{}, nil, f.err
	}
	return f.base, f.date, f.r, nil
}

type fakeStore struct {
	rec *CacheRecord
}

func (s *fakeStore) Load(base string) (*CacheRecord, error) { return s.rec, nil }
func (s *fakeStore) Save(rec *CacheRecord) error            { s.rec = rec; return nil }

func TestRates_CacheHit(t *testing.T) {
	f := &fakeFetcher{base: "USD", date: time.Now(), r: map[string]float64{"TRY": 33.3}}
	s := &RatesService{F: f, Store: &fakeStore{}, TTL: time.Hour, StaleTTL: 24 * time.Hour}

	b1, _, _, _ := s.Latest("usd")
	f.err = errors.New("down")
	b2, _, _, _ := s.Latest("usd")

	if b1 != "USD" || b2 != "USD" {
		t.Fatalf("cache not used")
	}
}

func TestRates_Fallback_Store(t *testing.T) {
	rec := &CacheRecord{Base: "EUR", Date: time.Now().Add(-24 * time.Hour).Format("2006-01-02"), Rates: map[string]float64{"TRY": 35.0}, SavedAt: time.Now()}
	f := &fakeFetcher{err: errors.New("down")}
	s := &RatesService{F: f, Store: &fakeStore{rec: rec}, TTL: time.Minute, StaleTTL: 48 * time.Hour}

	b, _, r, _ := s.Latest("eur")
	if b != "EUR" || r["TRY"] != 35.0 {
		t.Fatalf("fallback failed")
	}
}
