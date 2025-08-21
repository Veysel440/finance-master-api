package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/Veysel440/finance-master-api/internal/ports"
	"github.com/Veysel440/finance-master-api/internal/services"
	"github.com/go-chi/chi/v5"
)

type Handlers struct {
	Auth *services.AuthService
	Tx   *services.TxService
}

type txIn struct {
	Type       string  `json:"type"`
	Amount     float64 `json:"amount"`
	Currency   string  `json:"currency"`
	CategoryID int64   `json:"categoryId"`
	WalletID   int64   `json:"walletId"`
	Note       *string `json:"note"`
	OccurredAt string  `json:"occurredAt"`
}

func (h *Handlers) TxList(w http.ResponseWriter, r *http.Request) {
	uid := UID(r)
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	q := r.URL.Query().Get("q")
	items, total, err := h.Tx.List(uid, page, 20, q)
	if err != nil {
		errJSON(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{"total": total, "data": items})
}

func (h *Handlers) TxCreate(w http.ResponseWriter, r *http.Request) {
	uid := UID(r)
	var in txIn
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		errJSON(w, 400, "bad_request")
		return
	}
	if in.Type != "income" && in.Type != "expense" {
		errJSON(w, 400, "bad_type")
		return
	}
	if in.Amount <= 0 {
		errJSON(w, 400, "bad_amount")
		return
	}
	occ, err := time.Parse(time.RFC3339, in.OccurredAt)
	if err != nil {
		errJSON(w, 400, "bad_occurredAt")
		return
	}
	t := ports.Transaction{
		WalletID: in.WalletID, CategoryID: in.CategoryID, Type: in.Type,
		Amount: in.Amount, Currency: in.Currency, Note: in.Note, OccurredAt: occ,
	}
	if err := h.Tx.Create(uid, &t); err != nil {
		errJSON(w, 500, err.Error())
		return
	}
	writeJSON(w, 201, t)
}

func (h *Handlers) TxUpdate(w http.ResponseWriter, r *http.Request) {
	uid := UID(r)
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	var in txIn
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		errJSON(w, 400, "bad_request")
		return
	}
	occ, err := time.Parse(time.RFC3339, in.OccurredAt)
	if err != nil {
		errJSON(w, 400, "bad_occurredAt")
		return
	}
	t := ports.Transaction{
		ID: id, WalletID: in.WalletID, CategoryID: in.CategoryID, Type: in.Type,
		Amount: in.Amount, Currency: in.Currency, Note: in.Note, OccurredAt: occ,
	}
	if err := h.Tx.Update(uid, &t); err != nil {
		errJSON(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, t)
}

func (h *Handlers) TxDelete(w http.ResponseWriter, r *http.Request) {
	uid := UID(r)
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err := h.Tx.Delete(uid, id); err != nil {
		errJSON(w, 500, err.Error())
		return
	}
	w.WriteHeader(204)
}

func (h *Handlers) TxSince(w http.ResponseWriter, r *http.Request) {
	uid := UID(r)
	s := r.URL.Query().Get("since")
	tm, err := time.Parse(time.RFC3339, s)
	if err != nil {
		errJSON(w, 400, "bad_since")
		return
	}
	items, err := h.Tx.Since(uid, tm)
	if err != nil {
		errJSON(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, items)
}

func (h *Handlers) TxUpsertBatch(w http.ResponseWriter, r *http.Request) {
	uid := UID(r)
	var items []ports.Transaction
	if err := json.NewDecoder(r.Body).Decode(&items); err != nil {
		errJSON(w, 400, "bad_request")
		return
	}
	if err := h.Tx.UpsertBatch(uid, items); err != nil {
		errJSON(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"status": "ok"})
}

func (h *Handlers) TxSummary(w http.ResponseWriter, r *http.Request) {
	uid := UID(r)
	f := r.URL.Query().Get("from")
	t := r.URL.Query().Get("to")
	from, err := time.Parse(time.RFC3339, f)
	if err != nil {
		errJSON(w, 400, "bad_from")
		return
	}
	to, err2 := time.Parse(time.RFC3339, t)
	if err2 != nil {
		errJSON(w, 400, "bad_to")
		return
	}
	rows, err := h.Tx.Summary(uid, from, to)
	if err != nil {
		errJSON(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, rows)
}
