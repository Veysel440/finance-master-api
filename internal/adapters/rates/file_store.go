package rates

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type CacheRecord struct {
	Base    string             `json:"base"`
	Date    string             `json:"date"`
	Rates   map[string]float64 `json:"rates"`
	SavedAt time.Time          `json:"savedAt"`
}

type FileStore struct{ Path string }

func (f *FileStore) Load(base string) (*CacheRecord, error) {
	b, err := os.ReadFile(f.Path)
	if err != nil {
		return nil, err
	}
	var rec CacheRecord
	if err := json.Unmarshal(b, &rec); err != nil {
		return nil, err
	}
	if rec.Base != base && base != "" {
		return nil, os.ErrNotExist
	}
	return &rec, nil
}

func (f *FileStore) Save(rec *CacheRecord) error {
	if err := os.MkdirAll(filepath.Dir(f.Path), 0o755); err != nil {
		return err
	}
	b, _ := json.Marshal(rec)
	return os.WriteFile(f.Path, b, 0o644)
}
