package utils

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/go-playground/validator/v10"
)

var v = validator.New()

func ParseAndValidate(c *fiber.Ctx, out interface{}) bool {
	if err := c.BodyParser(out); err != nil {
		_ = Fail(c, fiber.StatusBadRequest, "payload tidak valid", nil)
		return false
	}
	if errs := Validate(out); errs != nil {
		_ = Fail(c, fiber.StatusUnprocessableEntity, "validasi gagal", errs)
		return false
	}
	return true
}

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func Validate(payload interface{}) []FieldError {
	err := v.Struct(payload)
	if err == nil {
		return nil
	}

	var fieldErrors []FieldError
	for _, e := range err.(validator.ValidationErrors) {
		fieldErrors = append(fieldErrors, FieldError{
			Field:   strings.ToLower(e.Field()),
			Message: buildMessage(e),
		})
	}
	return fieldErrors
}

func buildMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return fmt.Sprintf("%s wajib diisi", e.Field())
	case "email":
		return "format email tidak valid"
	case "min":
		return fmt.Sprintf("%s minimal %s karakter/nilai", e.Field(), e.Param())
	case "max":
		return fmt.Sprintf("%s maksimal %s karakter/nilai", e.Field(), e.Param())
	case "len":
		return fmt.Sprintf("%s harus %s karakter", e.Field(), e.Param())
	case "eqfield":
		return fmt.Sprintf("%s harus sama dengan %s", e.Field(), e.Param())
	default:
		return fmt.Sprintf("%s tidak valid (%s)", e.Field(), e.Tag())
	}
}
