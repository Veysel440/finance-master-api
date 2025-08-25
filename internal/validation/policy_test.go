package validation

import "testing"

func TestValidatePassword_OK(t *testing.T) {
	if err := ValidatePassword("A1strong!pass", "v@e.com"); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestValidatePassword_TooShort(t *testing.T) {
	if err := ValidatePassword("A1!short", "v@e.com"); err == nil {
		t.Fatal("expected weak_password")
	}
}

func TestValidatePassword_Classes(t *testing.T) {
	if err := ValidatePassword("alllowercaseonly", "v@e.com"); err == nil {
		t.Fatal("expected weak_password")
	}
}

func TestValidatePassword_Repeats(t *testing.T) {
	if err := ValidatePassword("AAAAbbbbb1111", "v@e.com"); err == nil {
		t.Fatal("expected weak_password")
	}
}

func TestValidatePassword_CommonList(t *testing.T) {
	if err := ValidatePassword("password", "v@e.com"); err == nil {
		t.Fatal("expected weak_password")
	}
}

func TestValidatePassword_EmailLocalLeak(t *testing.T) {
	if err := ValidatePassword("Veysel2024!", "Veysel@e.com"); err == nil {
		t.Fatal("expected weak_password")
	}
}

func TestValidatePassword_TrivialSequence(t *testing.T) {
	if err := ValidatePassword("1234567A!", "v@e.com"); err == nil {
		t.Fatal("expected weak_password")
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		pw, email string
		ok        bool
	}{
		{"short1A!", "a@b.c", false},
		{"noDigitsAAAA!", "user@test.com", false},
		{"alllowercase123!", "user@test.com", true},
		{"UPPERlower123", "user@test.com", true},
		{"userUSER123!", "user@test.com", false},
		{"aaaaBBBB!!!!", "x@y.z", false},
		{"Abcdefghi1!", "x@y.z", true},
	}
	for _, tt := range tests {
		err := ValidatePassword(tt.pw, tt.email)
		if tt.ok && err != nil {
			t.Errorf("expected ok for %q, got %v", tt.pw, err)
		}
		if !tt.ok && err == nil {
			t.Errorf("expected error for %q", tt.pw)
		}
	}
}
