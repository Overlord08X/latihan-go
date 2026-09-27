package helper

import (
	"errors"
	"reflect"
	"strings"
	"unicode"

	"github.com/go-playground/validator/v10"
)

// validate dibuat SEKALI untuk seluruh aplikasi.
var validate = newValidator()

func newValidator() *validator.Validate {
	v := validator.New()

	// Pemetaan nama field JSON ke nama field struct
	v.RegisterTagNameFunc(func(field reflect.StructField) string {
		name := strings.SplitN(field.Tag.Get("json"), ",", 2)[0]
		if name == "" || name == "-" {
			return field.Name
		}
		return name
	})

	// Aturan kustom 1: nospace — tidak boleh mengandung spasi atau whitespace
	_ = v.RegisterValidation("nospace", func(fl validator.FieldLevel) bool {
		return !strings.ContainsAny(fl.Field().String(), " \t\n\r")
	})

	// Aturan kustom 2: username — hanya huruf, angka, titik, dan garis bawah
	_ = v.RegisterValidation("username", func(fl validator.FieldLevel) bool {
		for _, r := range fl.Field().String() {
			if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '.' && r != '_' {
				return false
			}
		}
		return true
	})

	// Aturan kustom 3: strongpassword
	// Perbaikan Bug Perilaku #5: Bandingkan dengan == "" (valid jika tidak ada pesan error)
	_ = v.RegisterValidation("strongpassword", func(fl validator.FieldLevel) bool {
		return passwordStrength(fl.Field().String()) == ""
	})

	// Aturan kustom 4: nim untuk entitas student (Tugas Mandiri D.2)
	// NIM mahasiswa harus 9 digit angka numerik
	_ = v.RegisterValidation("nim", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		if len(val) != 9 {
			return false
		}
		for _, r := range val {
			if !unicode.IsDigit(r) {
				return false
			}
		}
		return true
	})

	return v
}

// ValidateStruct menjalankan seluruh aturan pada tag struct dan
// mengembalikan peta nama field ke pesan berbahasa Indonesia.
// Mengembalikan nil berarti tidak ada pelanggaran.
func ValidateStruct(s any) map[string]string {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	var invalid *validator.InvalidValidationError
	if errors.As(err, &invalid) {
		return map[string]string{"_": "objek yang divalidasi tidak sah"}
	}

	var fieldErrors validator.ValidationErrors
	if !errors.As(err, &fieldErrors) {
		return map[string]string{"_": "validasi gagal"}
	}

	result := make(map[string]string, len(fieldErrors))
	for _, fe := range fieldErrors {
		if _, exists := result[fe.Field()]; !exists {
			result[fe.Field()] = messageFor(fe)
		}
	}
	return result
}

// messageFor menerjemahkan nama tag menjadi kalimat yang dapat dibaca pemakai.
func messageFor(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "wajib diisi"
	case "email":
		return "format email tidak valid"
	case "min":
		if fe.Kind() == reflect.String {
			return "minimal " + fe.Param() + " karakter"
		}
		return "nilai minimal " + fe.Param()
	case "max":
		if fe.Kind() == reflect.String {
			return "maksimal " + fe.Param() + " karakter"
		}
		return "nilai maksimal " + fe.Param()
	case "alphanum":
		return "hanya boleh berisi huruf dan angka"
	case "nospace":
		return "tidak boleh mengandung spasi"
	case "username":
		return "hanya boleh huruf, angka, titik, dan garis bawah"
	case "nim":
		return "NIM harus tepat 9 digit angka"
	case "strongpassword":
		if value, ok := fe.Value().(string); ok {
			msg := passwordStrength(value)
			if msg != "" {
				return msg
			}
		}
		return "password tidak memenuhi syarat"
	case "oneof":
		return "harus salah satu dari: " +
			strings.ReplaceAll(fe.Param(), " ", ", ")
	default:
		return "tidak memenuhi aturan " + fe.Tag()
	}
}

// passwordStrength memeriksa kriteria password dan mengembalikan string pesan kesalahan jika tidak valid.
// Mengembalikan "" jika password kuat dan sah.
func passwordStrength(p string) string {
	if len(p) < 8 {
		return "minimal 8 karakter"
	}

	// Daftar password yang terlalu umum
	common := map[string]bool{
		"password":    true,
		"password123": true,
		"12345678":    true,
		"qwerty123":   true,
		"admin123":    true,
		"rahasia123":  true,
	}
	if common[strings.ToLower(p)] {
		return "password terlalu umum"
	}

	hasLetter := false
	hasDigit := false
	for _, r := range p {
		if unicode.IsLetter(r) {
			hasLetter = true
		}
		if unicode.IsDigit(r) {
			hasDigit = true
		}
	}
	if !hasLetter || !hasDigit {
		return "harus memuat huruf dan angka"
	}

	return ""
}
