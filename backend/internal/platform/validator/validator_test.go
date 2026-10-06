package validator

import (
	"errors"
	"testing"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/require"

	"indico/backend/internal/platform/httpx"
)

type validPayload struct {
	UserID string `json:"user_id" binding:"required,min=3,max=64,identifier"`
}

type invalidPayload struct {
	UserID string `json:"user_id" binding:"required,min=3,max=64,identifier"`
	Count  int    `json:"count" binding:"required,gt=0,max=1000"`
}

type everythingPayload struct {
	Required string `json:"required" binding:"required"`
	Min      string `json:"min" binding:"min=5"`
	Max      string `json:"max" binding:"max=3"`
	GT       int    `json:"gt" binding:"gt=0"`
	LTE      int    `json:"lte" binding:"lte=5"`
	Ident    string `json:"ident" binding:"identifier"`
	Fallback string `json:"fallback" binding:"number"`
}

func TestRegister(t *testing.T) {
	require.NoError(t, Register())
	defaultValidator, ok := binding.Validator.(*Default)
	require.True(t, ok)
	require.NotNil(t, defaultValidator.Engine())
}

func TestRegisterValidationSeamError(t *testing.T) {
	original := registerValidation
	t.Cleanup(func() { registerValidation = original })
	registerValidation = func(*validator.Validate, string, validator.Func) error {
		return errors.New("registrasi gagal")
	}

	require.Error(t, Register())

	_, err := newValidator()
	require.ErrorContains(t, err, "registrasi gagal")
}

func TestValidateStructBranches(t *testing.T) {
	v, err := newValidator()
	require.NoError(t, err)
	defaultValidator := &Default{validate: v}

	require.NoError(t, defaultValidator.ValidateStruct(nil))

	var nilPointer *validPayload
	require.NoError(t, defaultValidator.ValidateStruct(nilPointer))

	require.NoError(t, defaultValidator.ValidateStruct(&validPayload{UserID: "usr_9981"}))

	require.Error(t, defaultValidator.ValidateStruct(&invalidPayload{}))
	require.Error(t, defaultValidator.ValidateStruct(invalidPayload{}))

	require.NoError(t, defaultValidator.ValidateStruct(42))
	require.Same(t, v, defaultValidator.Engine())
}

func TestFieldNamesFollowJSONTags(t *testing.T) {
	type payload struct {
		NoTag   string `binding:"required"`
		Ignored string `json:"-" binding:"required"`
		Empty   string `json:"," binding:"required"`
	}
	v, err := newValidator()
	require.NoError(t, err)

	err = v.Struct(payload{})
	require.Error(t, err)

	fields := map[string]bool{}
	for _, detail := range Translate(err) {
		fields[detail.Field] = true
	}
	require.True(t, fields["NoTag"])
	require.True(t, fields["Ignored"])
	require.True(t, fields["Empty"])
}

func TestTranslateAllRules(t *testing.T) {
	v, err := newValidator()
	require.NoError(t, err)

	payload := everythingPayload{Max: "abcd", LTE: 10, Ident: "bad id!", Fallback: "abc"}
	err = v.Struct(payload)
	require.Error(t, err)

	byField := map[string]httpx.ErrorDetail{}
	for _, detail := range Translate(err) {
		byField[detail.Field] = detail
	}

	require.Equal(t, "required wajib diisi", byField["required"].Message)
	require.Equal(t, "min minimal 5 karakter", byField["min"].Message)
	require.Equal(t, "max maksimal 3 karakter", byField["max"].Message)
	require.Equal(t, "gt harus lebih dari 0", byField["gt"].Message)
	require.Equal(t, "lte maksimal 5", byField["lte"].Message)
	require.Contains(t, byField["ident"].Message, "hanya boleh berisi")
	require.Equal(t, "fallback tidak valid", byField["fallback"].Message)
}

func TestTranslateNonValidatorError(t *testing.T) {
	details := Translate(errors.New("json rusak"))
	require.Len(t, details, 1)
	require.Equal(t, "format request tidak valid", details[0].Message)
}

func TestValidateIdentifier(t *testing.T) {
	v, err := newValidator()
	require.NoError(t, err)

	type payload struct {
		Code string `json:"code" binding:"identifier"`
	}
	require.NoError(t, v.Struct(payload{Code: "usr_9981-A"}))
	require.Error(t, v.Struct(payload{Code: "spasi tidak boleh"}))
}
