// Package validator wires go-playground/validator as Gin's struct validator
// and registers the custom rules used by the API contracts.
package validator

import (
	"reflect"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// identifierPattern allows letters, digits, dash and underscore so ids such as
// usr_9981, item_4021 and res_883291 are accepted, while rejecting spaces and
// symbols.
var identifierPattern = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

// registerValidation is a seam so tests can force the registration failure
// path.
var registerValidation = func(v *validator.Validate, tag string, fn validator.Func) error {
	return v.RegisterValidation(tag, fn)
}

// Default implements gin's StructValidator interface.
type Default struct {
	validate *validator.Validate
}

// ValidateStruct implements gin's StructValidator interface.
func (d *Default) ValidateStruct(obj any) error {
	if obj == nil {
		return nil
	}
	switch reflect.TypeOf(obj).Kind() {
	case reflect.Ptr:
		value := reflect.ValueOf(obj)
		if value.IsNil() {
			return nil
		}
		return d.validate.Struct(value.Elem().Interface())
	case reflect.Struct:
		return d.validate.Struct(obj)
	default:
		return nil
	}
}

// Engine implements gin's StructValidator interface.
func (d *Default) Engine() any { return d.validate }

// Register installs the shared validator into Gin and registers custom rules.
// Field names in validation errors follow the JSON tags.
func Register() error {
	v, err := newValidator()
	if err != nil {
		return err
	}
	binding.Validator = &Default{validate: v}
	return nil
}

// newValidator builds a go-playground validator configured for gin's
// "binding" tags with JSON field names and the custom rules.
func newValidator() (*validator.Validate, error) {
	v := validator.New()
	v.SetTagName("binding")
	v.RegisterTagNameFunc(func(field reflect.StructField) string {
		// Match gin's behavior: prefer the json tag, then form, then the Go
		// field name.
		name := strings.SplitN(field.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			name = ""
		}
		if name == "" {
			name = strings.SplitN(field.Tag.Get("form"), ",", 2)[0]
		}
		if name == "" || name == "-" {
			return field.Name
		}
		return name
	})
	if err := registerValidation(v, "identifier", validateIdentifier); err != nil {
		return nil, err
	}
	return v, nil
}

func validateIdentifier(level validator.FieldLevel) bool {
	return identifierPattern.MatchString(level.Field().String())
}
