package validation

import (
	"regexp"
	"time"
	"unicode"

	"github.com/go-playground/validator/v10"
)

var v = validator.New()

// currency: 3 harf ISO
var reCurrency = regexp.MustCompile(`^[A-Z]{3}$`)

func noCtrl(fl validator.FieldLevel) bool {
	s, _ := fl.Field().Interface().(string)
	for _, r := range s {
		if unicode.IsControl(r) && r != '\n' && r != '\r' && r != '\t' {
			return false
		}
	}
	return true
}
func txType(fl validator.FieldLevel) bool {
	s, _ := fl.Field().Interface().(string)
	return s == "income" || s == "expense"
}
func isCurrency(fl validator.FieldLevel) bool {
	s, _ := fl.Field().Interface().(string)
	return reCurrency.MatchString(s)
}
func iso8601(fl validator.FieldLevel) bool {
	s, _ := fl.Field().Interface().(string)
	_, err := time.Parse(time.RFC3339, s)
	return err == nil
}

func init() {
	_ = v.RegisterValidation("noctrl", noCtrl)
	_ = v.RegisterValidation("txtype", txType)
	_ = v.RegisterValidation("currency", isCurrency)
	_ = v.RegisterValidation("iso8601", iso8601)
}

func ValidateStruct(s any) error { return v.Struct(s) }

func ValidationMessage(err error) string {
	if err == nil {
		return ""
	}
	if ve, ok := err.(validator.ValidationErrors); ok && len(ve) > 0 {
		f := ve[0]
		return f.Field() + ":" + f.Tag()
	}
	return err.Error()
}
