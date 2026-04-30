package validator

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"

	ae "skyfox/error"
)

func HandleStructValidationError(err error) interface{} {
	if err == nil {
		return nil
	}

	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) || len(validationErrors) == 0 {
		return ae.BadRequestError("invalid request", "malformed request payload", err)
	}

	var structErrors []string
	for _, fieldErr := range validationErrors {
		structErrors = append(structErrors, fieldError{fieldErr}.String())
		return ae.BadRequestError("validation failed", structErrors[0], fieldErr)
	}
	return nil
}

type fieldError struct {
	err validator.FieldError
}

func (e fieldError) String() string {
	var sb strings.Builder

	sb.WriteString("field '" + e.err.Field() + "'")
	sb.WriteString(", condition: " + validationErrorToText(e.err))

	if e.err.Value() != nil && e.err.Value() != "" {
		sb.WriteString(fmt.Sprintf(", provided: %v", e.err.Value()))
	}
	return sb.String()
}

func validationErrorToText(fieldErr validator.FieldError) string {
	switch fieldErr.ActualTag() {
	case "required":
		return fmt.Sprintf("%s is required", fieldErr.Field())
	case "max":
		return fmt.Sprintf("%s cannot be longer than %s", fieldErr.Field(), fieldErr.Param())
	case "min":
		return fmt.Sprintf("%s must be longer than %s", fieldErr.Field(), fieldErr.Param())
	case "email":
		return fmt.Sprintf("Invalid email format")
	case "len":
		return fmt.Sprintf("%s must be %s characters long", fieldErr.Field(), fieldErr.Param())
	case "gte":
		return fmt.Sprintf("%s must be greater than %s", fieldErr.Field(), fieldErr.Param())
	}
	return fmt.Sprintf("%s is not valid", fieldErr.Field())
}
