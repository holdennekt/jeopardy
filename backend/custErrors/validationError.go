package custErrors

import (
	"fmt"
	"reflect"

	"github.com/go-playground/validator/v10"
)

func messageForTag(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", fe.Field())
	case "min":
		switch fe.Type().Kind() {
		case reflect.String:
			return fmt.Sprintf("%s must be at least %s characters long", fe.StructNamespace(), fe.Param())
		case reflect.Int:
			return fmt.Sprintf("%s must be at least %s", fe.StructNamespace(), fe.Param())
		case reflect.Slice:
			return fmt.Sprintf("%s must have at least %s items", fe.StructNamespace(), fe.Param())
		}
	case "max":
		switch fe.Type().Kind() {
		case reflect.String:
			return fmt.Sprintf("%s must be at most %s characters long", fe.StructNamespace(), fe.Param())
		case reflect.Int:
			return fmt.Sprintf("%s must be at most %s", fe.StructNamespace(), fe.Param())
		case reflect.Slice:
			return fmt.Sprintf("%s must have at most %s items", fe.StructNamespace(), fe.Param())
		}
	case "url":
		return fmt.Sprintf("%s must be a URL", fe.StructNamespace())
	case "email":
		return fmt.Sprintf("%s must be a valid email", fe.StructNamespace())
	}
	return fe.Error()
}

func ParseValidationErrors(err error) []string {
	switch err := err.(type) {
	case validator.ValidationErrors:
		errs := make([]string, len(err))
		for i, fieldErr := range err {
			errs[i] = messageForTag(fieldErr)
		}
		return errs
	default:
		return []string{err.Error()}
	}
}
