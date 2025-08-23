package http

import (
	"net/http"

	"github.com/Veysel440/finance-master-api/internal/services"
)

type RatesHandlers struct{ S *services.RatesService }

func (h *RatesHandlers) Latest(w http.ResponseWriter, r *http.Request) {
	base := r.URL.Query().Get("base")
	b, d, rates, err := h.S.Latest(base)
	if err != nil {
		Fail(w, 502, "rates_unavailable", "rates unavailable")
		return
	}
	WriteJSON(w, 200, map[string]any{
		"base": b, "date": d.Format("2006-01-02"), "rates": rates,
	})
}
