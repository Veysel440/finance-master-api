package validation

import "testing"

type txIn struct {
	Type     string `validate:"required,txtype"`
	Currency string `validate:"required,currency"`
}

func TestTags(t *testing.T) {
	ok := txIn{Type: "income", Currency: "USD"}
	if err := ValidateStruct(ok); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	bad := txIn{Type: "foo", Currency: "USDT"}
	if err := ValidateStruct(bad); err == nil {
		t.Fatalf("expected validation error")
	}
}
