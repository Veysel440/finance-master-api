package http_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	api "github.com/Veysel440/finance-master-api/internal/http"
)

func Test_RequestID_HeaderPresent(t *testing.T) {
	r := api.Router(&api.API{})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	r.ServeHTTP(w, req)
	if w.Header().Get("X-Request-Id") == "" {
		t.Fatalf("X-Request-Id missing")
	}
}
