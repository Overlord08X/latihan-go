package helper

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()

	// Register custom tag 'nim'
	_ = validate.RegisterValidation("nim", func(fl validator.FieldLevel) bool {
		return ValidateNIM(fl.Field().String())
	})

	// Register custom tag 'angkatan_valid'
	_ = validate.RegisterValidation("angkatan_valid", func(fl validator.FieldLevel) bool {
		val := int(fl.Field().Int())
		return ValidateAngkatan(val, time.Now().Year())
	})

	// Register custom tag 'tahun_akademik'
	_ = validate.RegisterValidation("tahun_akademik", func(fl validator.FieldLevel) bool {
		return ValidateTahunAkademik(fl.Field().String())
	})
}

func GetValidator() *validator.Validate {
	return validate
}

func FormatValidationErrors(err error) map[string][]string {
	res := make(map[string][]string)
	valErrs, ok := err.(validator.ValidationErrors)
	if !ok {
		res["general"] = []string{err.Error()}
		return res
	}

	for _, fe := range valErrs {
		field := strings.ToLower(fe.Field())
		var msg string
		switch fe.Tag() {
		case "required":
			msg = fmt.Sprintf("Field %s wajib diisi", field)
		case "email":
			msg = "Format email tidak valid"
		case "min":
			msg = fmt.Sprintf("Field %s minimal %s karakter", field, fe.Param())
		case "max":
			msg = fmt.Sprintf("Field %s maksimal %s karakter", field, fe.Param())
		case "nim":
			msg = "NIM harus tepat 12 digit angka"
		case "angkatan_valid":
			msg = fmt.Sprintf("Angkatan harus 4 digit dan tidak melebihi tahun %d", time.Now().Year())
		case "tahun_akademik":
			msg = "Format tahun akademik harus YYYY/YYYY-(Ganjil|Genap), misal: 2026/2027-Ganjil"
		case "gte":
			msg = fmt.Sprintf("Field %s minimal %s", field, fe.Param())
		case "lte":
			msg = fmt.Sprintf("Field %s maksimal %s", field, fe.Param())
		default:
			msg = fmt.Sprintf("Field %s tidak valid (%s)", field, fe.Tag())
		}
		res[field] = append(res[field], msg)
	}
	return res
}
