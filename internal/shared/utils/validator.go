package utils

import (
	"net/mail"
	"strings"
	"unicode"

	"github.com/go-playground/validator/v10"
)

func ValidationErrors(err error) map[string]string {
	errors := map[string]string{}

	if err == nil {
		return errors
	}

	for _, err := range err.(validator.ValidationErrors) {
		field := err.Field()
		field = toSnakeCase(field)

		switch err.Tag() {
		case "required":
			errors[field] = field + " is required"
		case "email":
			errors[field] = "invalid email format"
		case "password":
			errors[field] = "password must be at least 8 characters long and include uppercase, lowercase, digit, and special character (@$!%*?&)"
		case "min":
			errors[field] = field + " must be at least " + err.Param() + " characters"
		case "max":
			errors[field] = field + " must be at most " + err.Param() + " characters"
		case "uuid":
			errors[field] = field + " must be a valid UUID"
		default:
			errors[field] = "invalid value for " + field
		}
	}

	return errors
}

func toSnakeCase(s string) string {
	var output []rune
	for i, r := range s {
		if i > 0 && unicode.IsUpper(r) {
			output = append(output, '_')
		}
		output = append(output, unicode.ToLower(r))
	}
	return string(output)
}

func IsStrongPassword(p string) bool {
	if len(p) < 8 {
		return false
	}

	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, c := range p {
		switch {
		case unicode.IsUpper(c):
			hasUpper = true
		case unicode.IsLower(c):
			hasLower = true
		case unicode.IsDigit(c):
			hasDigit = true
		case strings.ContainsRune("@$!%*?&", c):
			hasSpecial = true
		}
	}

	return hasUpper && hasLower && hasDigit && hasSpecial
}

func ValidateEmail(email string) error {
	_, err := mail.ParseAddress(email)
	return err
}

var Validate = validator.New()

func init() {
	Validate.RegisterValidation("password", func(fl validator.FieldLevel) bool {
		return IsStrongPassword(fl.Field().String())
	})
	Validate.RegisterValidation("email", func(fl validator.FieldLevel) bool {
		err := ValidateEmail(fl.Field().String())
		return err == nil
	})
}
