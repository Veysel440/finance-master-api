package validation

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

var (
	V            *validator.Validate
	rxCurrency   = regexp.MustCompile(`^[A-Z]{3}$`)
	allowedTypes = map[string]struct{}{"income": {}, "expense": {}}
)

func init() {
	V = validator.New(validator.WithRequiredStructEnabled())
	_ = V.RegisterValidation("currency", func(fl validator.FieldLevel) bool {
		s, _ := fl.Field().Interface().(string)
		return s == "" || rxCurrency.MatchString(s)
	})
	_ = V.RegisterValidation("txtype", func(fl validator.FieldLevel) bool {
		s := strings.ToLower(fl.Field().String())
		_, ok := allowedTypes[s]
		return ok
	})
}

func ValidateStruct(s any) error { return V.Struct(s) }

func ValidationMessage(err error) string {
	if err == nil {
		return ""
	}
	ves, ok := err.(validator.ValidationErrors)
	if !ok {
		return err.Error()
	}
	parts := make([]string, 0, len(ves))
	for _, fe := range ves {
		switch fe.Tag() {
		case "required":
			parts = append(parts, fmt.Sprintf("%s is required", fe.Field()))
		case "email":
			parts = append(parts, "invalid email")
		case "currency":
			parts = append(parts, "invalid currency")
		case "txtype":
			parts = append(parts, "type must be income|expense")
		case "max":
			parts = append(parts, fmt.Sprintf("%s too long", fe.Field()))
		case "min":
			parts = append(parts, fmt.Sprintf("%s too short", fe.Field()))
		default:
			parts = append(parts, fmt.Sprintf("%s invalid", fe.Field()))
		}
	}
	return strings.Join(parts, "; ")
}
