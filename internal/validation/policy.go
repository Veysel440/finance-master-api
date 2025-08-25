package validation

import (
	"errors"
	"strings"
	"unicode"
)

var ErrWeakPassword = errors.New("weak_password")

var commonPasswords = map[string]struct{}{
	"123456": {}, "123456789": {}, "qwerty": {}, "password": {}, "111111": {},
	"123123": {}, "abc123": {}, "iloveyou": {}, "000000": {}, "password1": {},
	"1q2w3e4r": {}, "admin": {}, "welcome": {}, "monkey": {}, "dragon": {},
}

func ValidatePassword(pw, email string) error {
	if len(pw) < 10 || len(pw) > 256 {
		return ErrWeakPassword
	}
	var lower, upper, digit, sym bool
	for _, r := range pw {
		switch {
		case unicode.IsLower(r):
			lower = true
		case unicode.IsUpper(r):
			upper = true
		case unicode.IsDigit(r):
			digit = true
		case unicode.IsSpace(r):
			return ErrWeakPassword
		default:

			if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
				sym = true
			}
		}
	}
	classes := 0
	for _, ok := range []bool{lower, upper, digit, sym} {
		if ok {
			classes++
		}
	}
	if classes < 3 {
		return ErrWeakPassword
	}

	local := email
	if i := strings.IndexByte(email, '@'); i > 0 {
		local = email[:i]
	}
	if local != "" && strings.Contains(strings.ToLower(pw), strings.ToLower(local)) {
		return ErrWeakPassword
	}

	run := 1
	for i := 1; i < len(pw); i++ {
		if pw[i] == pw[i-1] {
			run++
			if run >= 4 {
				return ErrWeakPassword
			}
		} else {
			run = 1
		}
	}

	lc := strings.ToLower(pw)
	if _, bad := commonPasswords[lc]; bad {
		return ErrWeakPassword
	}

	if onlyDigits(lc) && isTrivialSequence(lc) {
		return ErrWeakPassword
	}

	return nil
}

func onlyDigits(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return len(s) > 0
}

func isTrivialSequence(s string) bool {
	if len(s) < 6 {
		return false
	}
	inc := 0
	for i := 1; i < len(s); i++ {
		if s[i] == s[i-1]+1 {
			inc++
			if inc >= 5 {
				return true
			}
		} else {
			inc = 0
		}
	}
	return false
}
