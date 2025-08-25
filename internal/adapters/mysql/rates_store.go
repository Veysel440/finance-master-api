package mysql

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/Veysel440/finance-master-api/internal/services"
)

type RatesStore struct{ DB *sql.DB }

func NewRatesStore(db *sql.DB) *RatesStore { return &RatesStore{DB: db} }

func (s *RatesStore) Load(base string) (*services.CacheRecord, error) {
	row := s.DB.QueryRowContext(context.Background(),
		`SELECT base, rate_date, rates, saved_at FROM rates_cache WHERE base = ?`, base)
	var b string
	var d time.Time
	var saved time.Time
	var js []byte
	if err := row.Scan(&b, &d, &js, &saved); err != nil {
		return nil, err
	}
	var m map[string]float64
	if err := json.Unmarshal(js, &m); err != nil {
		return nil, err
	}
	return &services.CacheRecord{
		Base: b, Date: d.Format("2006-01-02"), Rates: m, SavedAt: saved,
	}, nil
}

func (s *RatesStore) Save(rec *services.CacheRecord) error {
	js, _ := json.Marshal(rec.Rates)
	_, err := s.DB.ExecContext(context.Background(), `
		INSERT INTO rates_cache (base, rate_date, rates, saved_at)
		VALUES (?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE rate_date=VALUES(rate_date), rates=VALUES(rates), saved_at=VALUES(saved_at)
	`, rec.Base, rec.Date, js, time.Now())
	return err
}

var _ services.RatesStore = (*RatesStore)(nil)
