package rates

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type HTTPClient struct {
	BaseURL string
	Client  *http.Client
}

type latestResp struct {
	Base  string             `json:"base"`
	Date  string             `json:"date"`
	Rates map[string]float64 `json:"rates"`
}

func (h *HTTPClient) Latest(base string) (string, time.Time, map[string]float64, error) {
	u := fmt.Sprintf("%s?base=%s", h.BaseURL, base)
	resp, err := h.Client.Get(u)
	if err != nil {
		return "", time.Time{}, nil, err
	}
	defer resp.Body.Close()
	var lr latestResp
	if err := json.NewDecoder(resp.Body).Decode(&lr); err != nil {
		return "", time.Time{}, nil, err
	}
	d, _ := time.Parse("2006-01-02", lr.Date)
	return lr.Base, d, lr.Rates, nil
}
