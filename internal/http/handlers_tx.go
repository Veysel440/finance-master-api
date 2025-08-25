package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/Veysel440/finance-master-api/internal/errs"
	imw "github.com/Veysel440/finance-master-api/internal/http/middleware"
	"github.com/Veysel440/finance-master-api/internal/ports"
	"github.com/Veysel440/finance-master-api/internal/services"
	"github.com/Veysel440/finance-master-api/internal/validation"
	"github.com/go-chi/chi/v5"
)

type Handlers struct {
	Auth *services.AuthService
	Tx   *services.TxService
}

type txIn struct {
	Type       string  `json:"type"       validate:"required,txtype"`
	Amount     float64 `json:"amount"     validate:"required,gt=0"`
	Currency   string  `json:"currency"   validate:"required,currency"`
	CategoryID int64   `json:"categoryId" validate:"required,gt=0"`
	WalletID   int64   `json:"walletId"   validate:"required,gt=0"`
	Note       *string `json:"note"       validate:"omitempty,max=1000"`
	OccurredAt string  `json:"occurredAt" validate:"required,datetime=2006-01-02T15:04:05Z07:00"`
}

func clampPage(r *http.Request) (int, int) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	size, _ := strconv.Atoi(r.URL.Query().Get("size"))
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	return page, size
}

func (h *Handlers) TxList(w http.ResponseWriter, r *http.Request) {
	uid := UID(r)
	page, size := clampPage(r)
	q := r.URL.Query().Get("q")
	f := r.URL.Query().Get("from")
	t := r.URL.Query().Get("to")

	if f != "" && t != "" {
		from, err := time.Parse(time.RFC3339, f)
		if err != nil {
			WriteAppError(w, errs.ValidationFailed("bad from"))
			return
		}
		to, err2 := time.Parse(time.RFC3339, t)
		if err2 != nil {
			WriteAppError(w, errs.ValidationFailed("bad to"))
			return
		}
		items, total, err := h.Tx.ListRange(uid, from, to, q, page, size)
		if err != nil {
			FromError(w, err)
			return
		}
		WriteJSON(w, 200, map[string]any{"total": total, "data": items})
		return
	}

	items, total, err := h.Tx.List(uid, page, size, q)
	if err != nil {
		FromError(w, err)
		return
	}
	WriteJSON(w, 200, map[string]any{"total": total, "data": items})
}

func (h *Handlers) TxCreate(w http.ResponseWriter, r *http.Request) {
	uid := UID(r)
	var in txIn
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		Fail(w, 400, "bad_request", "invalid json")
		return
	}
	if in.Type != "income" && in.Type != "expense" {
		WriteAppError(w, errs.ValidationFailed("bad type"))
		return
	}
	if in.Amount <= 0 {
		WriteAppError(w, errs.ValidationFailed("bad amount"))
		return
	}
	occ, err := time.Parse(time.RFC3339, in.OccurredAt)
	if err != nil {
		WriteAppError(w, errs.ValidationFailed("bad occurredAt"))
		return
	}
	occ = occ.UTC()

	t := ports.Transaction{
		WalletID: in.WalletID, CategoryID: in.CategoryID, Type: in.Type,
		Amount: in.Amount, Currency: in.Currency, Note: in.Note, OccurredAt: occ,
	}

	key := imw.FromContext(r)
	if key != "" {
		if err := h.Tx.CreateIdem(uid, key, &t); err != nil {
			FromError(w, err)
			return
		}
		WriteJSON(w, 201, t)
		return
	}

	if err := h.Tx.Create(uid, &t); err != nil {
		FromError(w, err)
		return
	}
	WriteJSON(w, 201, t)
}

func (h *Handlers) TxUpdate(w http.ResponseWriter, r *http.Request) {
	uid := UID(r)
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	var in txIn
	if err := DecodeStrict(r, &in); err != nil {
		Fail(w, 400, "bad_request", "invalid json")
		return
	}
	if err := validation.ValidateStruct(in); err != nil {
		WriteAppError(w, errs.ValidationFailed(validation.ValidationMessage(err)))
		return
	}
	occ, _ := time.Parse(time.RFC3339, in.OccurredAt)
	t := ports.Transaction{
		ID: id, WalletID: in.WalletID, CategoryID: in.CategoryID, Type: in.Type,
		Amount: in.Amount, Currency: in.Currency, Note: in.Note, OccurredAt: occ,
	}
	if err := h.Tx.Update(uid, &t); err != nil {
		FromError(w, err)
		return
	}
	WriteJSON(w, 200, t)
}

func (h *Handlers) TxDelete(w http.ResponseWriter, r *http.Request) {
	uid := UID(r)
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err := h.Tx.Delete(uid, id); err != nil {
		FromError(w, err)
		return
	}
	w.WriteHeader(204)
}

func (h *Handlers) TxSince(w http.ResponseWriter, r *http.Request) {
	uid := UID(r)
	s := r.URL.Query().Get("since")
	tm, err := time.Parse(time.RFC3339, s)
	if err != nil {
		WriteAppError(w, errs.ValidationFailed("bad since"))
		return
	}
	items, err := h.Tx.Since(uid, tm)
	if err != nil {
		FromError(w, err)
		return
	}
	WriteJSON(w, 200, items)
}

func (h *Handlers) TxUpsertBatch(w http.ResponseWriter, r *http.Request) {
	uid := UID(r)
	var items []ports.Transaction
	if err := json.NewDecoder(r.Body).Decode(&items); err != nil {
		Fail(w, 400, "bad_request", "invalid json")
		return
	}
	if len(items) > 500 {
		Fail(w, 413, "payload_too_large", "max 500 items")
		return
	}
	if err := h.Tx.UpsertBatch(uid, items); err != nil {
		FromError(w, err)
		return
	}
	WriteJSON(w, 200, map[string]string{"status": "ok"})
}

func (h *Handlers) TxGetOne(w http.ResponseWriter, r *http.Request) {
	uid := UID(r)
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	t, err := h.Tx.GetOne(uid, id)
	if err != nil {
		FromError(w, err)
		return
	}
	WriteJSON(w, 200, t)
}

func (h *Handlers) TxSummary(w http.ResponseWriter, r *http.Request) {
	uid := UID(r)
	f := r.URL.Query().Get("from")
	t := r.URL.Query().Get("to")
	from, err := time.Parse(time.RFC3339, f)
	if err != nil {
		WriteAppError(w, errs.ValidationFailed("bad from"))
		return
	}
	to, err2 := time.Parse(time.RFC3339, t)
	if err2 != nil {
		WriteAppError(w, errs.ValidationFailed("bad to"))
		return
	}
	rows, err := h.Tx.Summary(uid, from, to)
	if err != nil {
		FromError(w, err)
		return
	}
	WriteJSON(w, 200, rows)
}
