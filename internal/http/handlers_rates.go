package http

import (
	"net/http"

	"github.com/Veysel440/finance-master-api/internal/services"
	"github.com/Veysel440/finance-master-api/internal/validation"
)

type RatesHandlers struct{ S *services.RatesService }

func (h *RatesHandlers) Latest(w http.ResponseWriter, r *http.Request) {
	base := r.URL.Query().Get("base")
	type rq struct {
		Base string `validate:"omitempty,currency"`
	}
	q := rq{Base: base}
	if err := validation.ValidateStruct(q); err != nil {
		Fail(w, 400, "validation_failed", validation.ValidationMessage(err))
		return
	}

	b, d, rates, err := h.S.Latest(base)
	if err != nil {
		Fail(w, 502, "rates_unavailable", "rates service unavailable")
		return
	}
	WriteJSON(w, 200, map[string]any{
		"base": b, "date": d.Format("2006-01-02"), "rates": rates,
	})
}
