package validator

import (
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"

	"indico/backend/internal/platform/httpx"
)

// Translate converts any binding error into field-level details with Bahasa
// Indonesia messages. Non-validator errors (malformed JSON, wrong types) get a
// single generic detail.
func Translate(err error) []httpx.ErrorDetail {
	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		return []httpx.ErrorDetail{{Message: "format request tidak valid"}}
	}

	details := make([]httpx.ErrorDetail, 0, len(validationErrors))
	for _, fieldError := range validationErrors {
		details = append(details, httpx.ErrorDetail{
			Field:   fieldError.Field(),
			Rule:    fieldError.Tag(),
			Message: messageFor(fieldError),
		})
	}
	return details
}

func messageFor(fieldError validator.FieldError) string {
	switch fieldError.Tag() {
	case "required":
		return fmt.Sprintf("%s wajib diisi", fieldError.Field())
	case "min":
		return fmt.Sprintf("%s minimal %s karakter", fieldError.Field(), fieldError.Param())
	case "max":
		return fmt.Sprintf("%s maksimal %s karakter", fieldError.Field(), fieldError.Param())
	case "gt":
		return fmt.Sprintf("%s harus lebih dari %s", fieldError.Field(), fieldError.Param())
	case "lte":
		return fmt.Sprintf("%s maksimal %s", fieldError.Field(), fieldError.Param())
	case "identifier":
		return fmt.Sprintf("%s hanya boleh berisi huruf, angka, tanda hubung, dan garis bawah", fieldError.Field())
	default:
		return fmt.Sprintf("%s tidak valid", fieldError.Field())
	}
}
