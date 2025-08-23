package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Veysel440/finance-master-api/internal/errs"
	"github.com/Veysel440/finance-master-api/internal/ports"
	"github.com/Veysel440/finance-master-api/internal/services"
	"github.com/go-chi/chi/v5"
)

type CatalogHandlers struct {
	Wallet *services.WalletService
	Cat    *services.CategoryService
}

/* Wallets */
type walletReq struct {
	Name     string `json:"name"`
	Currency string `json:"currency"`
}

func (h *CatalogHandlers) WalletList(w http.ResponseWriter, r *http.Request) {
	rows, err := h.Wallet.List(UID(r))
	if err != nil {
		FromError(w, err)
		return
	}
	WriteJSON(w, 200, rows)
}

func (h *CatalogHandlers) WalletCreate(w http.ResponseWriter, r *http.Request) {
	var in walletReq
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		Fail(w, 400, "bad_request", "invalid json")
		return
	}
	wal := ports.Wallet{Name: in.Name, Currency: in.Currency}
	if err := h.Wallet.Create(UID(r), &wal); err != nil {
		// basit doğrulama hataları 400
		WriteAppError(w, errs.ValidationFailed(err.Error()))
		return
	}
	WriteJSON(w, 201, wal)
}

func (h *CatalogHandlers) WalletUpdate(w http.ResponseWriter, r *http.Request) {
	var in walletReq
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		Fail(w, 400, "bad_request", "invalid json")
		return
	}
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	wal := ports.Wallet{ID: id, Name: in.Name, Currency: in.Currency}
	if err := h.Wallet.Update(UID(r), &wal); err != nil {
		WriteAppError(w, errs.ValidationFailed(err.Error()))
		return
	}
	WriteJSON(w, 200, wal)
}

func (h *CatalogHandlers) WalletDelete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err := h.Wallet.Delete(UID(r), id); err != nil {
		FromError(w, err)
		return
	}
	w.WriteHeader(204)
}

/* Categories */
type catReq struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

func (h *CatalogHandlers) CategoryList(w http.ResponseWriter, r *http.Request) {
	typ := r.URL.Query().Get("type")
	rows, err := h.Cat.List(UID(r), typ)
	if err != nil {
		FromError(w, err)
		return
	}
	WriteJSON(w, 200, rows)
}

func (h *CatalogHandlers) CategoryCreate(w http.ResponseWriter, r *http.Request) {
	var in catReq
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		Fail(w, 400, "bad_request", "invalid json")
		return
	}
	c := ports.Category{Name: in.Name, Type: in.Type}
	if err := h.Cat.Create(UID(r), &c); err != nil {
		WriteAppError(w, errs.ValidationFailed(err.Error()))
		return
	}
	WriteJSON(w, 201, c)
}

func (h *CatalogHandlers) CategoryUpdate(w http.ResponseWriter, r *http.Request) {
	var in catReq
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		Fail(w, 400, "bad_request", "invalid json")
		return
	}
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	c := ports.Category{ID: id, Name: in.Name, Type: in.Type}
	if err := h.Cat.Update(UID(r), &c); err != nil {
		WriteAppError(w, errs.ValidationFailed(err.Error()))
		return
	}
	WriteJSON(w, 200, c)
}

func (h *CatalogHandlers) CategoryDelete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err := h.Cat.Delete(UID(r), id); err != nil {
		FromError(w, err)
		return
	}
	w.WriteHeader(204)
}
