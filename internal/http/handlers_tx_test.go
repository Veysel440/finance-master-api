package http

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestTxCreate_BadType(t *testing.T) {
	h := &Handlers{}
	iso := time.Now().UTC().Format(time.RFC3339)
	body := `{"type":"foo","amount":10,"currency":"USD","walletId":1,"categoryId":1,"occurredAt":"` + iso + `"}`

	req := httptest.NewRequest(http.MethodPost, "/v1/transactions", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	h.TxCreate(w, req)

	if w.Code != 400 {
		t.Fatalf("want 400, got %d: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "validation") {
		t.Fatalf("expected validation message, got %s", w.Body.String())
	}
}

func TestTxUpsertBatch_TooLarge(t *testing.T) {
	h := &Handlers{}
	var sb strings.Builder
	sb.WriteString("[")
	for i := 0; i < 501; i++ {
		if i > 0 {
			sb.WriteByte(',')
		}
		sb.WriteString(`{}`)
	}
	sb.WriteString("]")

	req := httptest.NewRequest(http.MethodPost, "/v1/sync/transactions", bytes.NewBufferString(sb.String()))
	w := httptest.NewRecorder()
	h.TxUpsertBatch(w, req)

	if w.Code != 413 {
		t.Fatalf("want 413, got %d: %s", w.Code, w.Body.String())
	}
}
