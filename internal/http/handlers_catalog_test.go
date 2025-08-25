package http

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWalletCreate_Validation(t *testing.T) {
	h := &CatalogHandlers{}
	body := `{"name":"","currency":"US"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/wallets", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.WalletCreate(w, req)

	if w.Code != 400 {
		t.Fatalf("want 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCategoryCreate_Validation(t *testing.T) {
	h := &CatalogHandlers{}
	body := `{"name":"","type":"foo"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/categories", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.CategoryCreate(w, req)

	if w.Code != 400 {
		t.Fatalf("want 400, got %d: %s", w.Code, w.Body.String())
	}
}
