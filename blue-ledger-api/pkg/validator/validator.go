package validator

import (
	"net/http"
	"reflect"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

// CustomValidator wraps go-playground/validator for use with Echo.
type CustomValidator struct {
	v *validator.Validate
}

// New creates a new CustomValidator instance.
func New() *CustomValidator {
	v := validator.New()

	// Register custom tag name function so validator uses JSON field names in error messages.
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		return fld.Tag.Get("json")
	})

	return &CustomValidator{v: v}
}

// Validate implements echo.Validator.
func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.v.Struct(i); err != nil {
		// Return a 422 Unprocessable Entity with the validation errors.
		return echo.NewHTTPError(http.StatusUnprocessableEntity, err.Error())
	}
	return nil
}
