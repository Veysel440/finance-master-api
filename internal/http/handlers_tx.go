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

func write(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
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
		http.Error(w, err.Error(), 500)
		return
	}
	write(w, 200, map[string]any{"total": total, "data": items})
}

func (h *Handlers) TxCreate(w http.ResponseWriter, r *http.Request) {
	uid := UID(r)
	var t ports.Transaction
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		http.Error(w, "bad_request", 400)
		return
	}
	if err := h.Tx.Create(uid, &t); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	write(w, 201, t)
}

func (h *Handlers) TxUpdate(w http.ResponseWriter, r *http.Request) {
	uid := UID(r)
	var t ports.Transaction
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		http.Error(w, "bad_request", 400)
		return
	}
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	t.ID = id
	if err := h.Tx.Update(uid, &t); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	write(w, 200, t)
}

func (h *Handlers) TxDelete(w http.ResponseWriter, r *http.Request) {
	uid := UID(r)
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err := h.Tx.Delete(uid, id); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	write(w, 204, nil)
}

func (h *Handlers) TxSince(w http.ResponseWriter, r *http.Request) {
	uid := UID(r)
	s := r.URL.Query().Get("since")
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		http.Error(w, "bad_since", 400)
		return
	}
	items, err := h.Tx.Since(uid, t)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	write(w, 200, items)
}
func (h *Handlers) TxUpsertBatch(w http.ResponseWriter, r *http.Request) {
	uid := UID(r)
	var items []ports.Transaction
	if err := json.NewDecoder(r.Body).Decode(&items); err != nil {
		http.Error(w, "bad_request", 400)
		return
	}
	if err := h.Tx.UpsertBatch(uid, items); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	write(w, 200, map[string]string{"status": "ok"})
}

func (h *Handlers) Register(writer http.ResponseWriter, request *http.Request) {

}

func (h *Handlers) Login(writer http.ResponseWriter, request *http.Request) {

}

func (h *Handlers) TxSummary(w http.ResponseWriter, r *http.Request) {
	uid := UID(r)
	f := r.URL.Query().Get("from")
	t := r.URL.Query().Get("to")
	if f == "" || t == "" {
		http.Error(w, "from/to required (RFC3339)", 400)
		return
	}
	from, err := time.Parse(time.RFC3339, f)
	if err != nil {
		http.Error(w, "bad_from", 400)
		return
	}
	to, err2 := time.Parse(time.RFC3339, t)
	if err2 != nil {
		http.Error(w, "bad_to", 400)
		return
	}
	rows, err := h.Tx.Summary(uid, from, to)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	write(w, 200, rows)
}
