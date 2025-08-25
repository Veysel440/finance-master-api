package http

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/Veysel440/finance-master-api/internal/errs"
	"github.com/Veysel440/finance-master-api/internal/ports"
	"github.com/Veysel440/finance-master-api/internal/services"
	"github.com/Veysel440/finance-master-api/internal/validation"
	"github.com/go-chi/chi/v5"
)

type CatalogHandlers struct {
	Wallet *services.WalletService
	Cat    *services.CategoryService
}

/* Wallets */
type walletReq struct {
	Name     string `json:"name"     validate:"required,min=1,max=100"`
	Currency string `json:"currency" validate:"required,currency"`
}

func (h *CatalogHandlers) WalletList(w http.ResponseWriter, r *http.Request) {
	rows, err := h.Wallet.List(UID(r))
	if err != nil {
		FromError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

func (h *CatalogHandlers) WalletCreate(w http.ResponseWriter, r *http.Request) {
	var in walletReq
	if err := DecodeStrict(r, &in); err != nil {
		Fail(w, 400, "bad_request", "invalid json")
		return
	}
	if err := validation.ValidateStruct(in); err != nil {
		WriteAppError(w, errs.ValidationFailed(validation.ValidationMessage(err)))
		return
	}
	wal := ports.Wallet{Name: in.Name, Currency: in.Currency}
	if err := h.Wallet.Create(UID(r), &wal); err != nil {
		FromError(w, err)
		return
	}
	WriteJSON(w, http.StatusCreated, wal)
}

func (h *CatalogHandlers) WalletUpdate(w http.ResponseWriter, r *http.Request) {
	var in walletReq
	if err := DecodeStrict(r, &in); err != nil {
		Fail(w, 400, "bad_request", "invalid json")
		return
	}
	if err := validation.ValidateStruct(in); err != nil {
		WriteAppError(w, errs.ValidationFailed(validation.ValidationMessage(err)))
		return
	}
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	wal := ports.Wallet{ID: id, Name: in.Name, Currency: in.Currency}
	if err := h.Wallet.Update(UID(r), &wal); err != nil {
		FromError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, wal)
}

func (h *CatalogHandlers) WalletDelete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err := h.Wallet.Delete(UID(r), id); err != nil {
		if err == errs.HasTransactions {
			WriteAppError(w, errs.HasTransactions)
			return
		}
		FromError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

/* Categories */
type catReq struct {
	Name string `json:"name" validate:"required,min=1,max=100"`
	Type string `json:"type" validate:"required,txtype"`
}

func (h *CatalogHandlers) CategoryList(w http.ResponseWriter, r *http.Request) {
	typ := strings.TrimSpace(r.URL.Query().Get("type"))
	rows, err := h.Cat.List(UID(r), typ)
	if err != nil {
		FromError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

func (h *CatalogHandlers) CategoryCreate(w http.ResponseWriter, r *http.Request) {
	var in catReq
	if err := DecodeStrict(r, &in); err != nil {
		Fail(w, 400, "bad_request", "invalid json")
		return
	}
	if err := validation.ValidateStruct(in); err != nil {
		WriteAppError(w, errs.ValidationFailed(validation.ValidationMessage(err)))
		return
	}
	c := ports.Category{Name: in.Name, Type: in.Type}
	if err := h.Cat.Create(UID(r), &c); err != nil {
		FromError(w, err)
		return
	}
	WriteJSON(w, http.StatusCreated, c)
}

func (h *CatalogHandlers) CategoryUpdate(w http.ResponseWriter, r *http.Request) {
	var in catReq
	if err := DecodeStrict(r, &in); err != nil {
		Fail(w, 400, "bad_request", "invalid json")
		return
	}
	if err := validation.ValidateStruct(in); err != nil {
		WriteAppError(w, errs.ValidationFailed(validation.ValidationMessage(err)))
		return
	}
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	c := ports.Category{ID: id, Name: in.Name, Type: in.Type}
	if err := h.Cat.Update(UID(r), &c); err != nil {
		FromError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, c)
}

func (h *CatalogHandlers) CategoryDelete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err := h.Cat.Delete(UID(r), id); err != nil {
		if err == errs.HasTransactions {
			WriteAppError(w, errs.HasTransactions)
			return
		}
		FromError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
