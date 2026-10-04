package helper

import (
	"errors"
	"reflect"
	"regexp"
	"strings"
	"time"
	"unicode"
	"github.com/go-playground/validator/v10"
)

var tahunAkademikRegex = regexp.MustCompile(`^\d{4}/\d{4}-(Ganjil|Genap)$`)

// validate dibuat SEKALI untuk seluruh aplikasi.
var validate = newValidator()

func newValidator() *validator.Validate {
	v := validator.New()
	// Tanpa ini, pesan error menyebut nama field Go, padahal client mengirim nama JSON.
	v.RegisterTagNameFunc(func(field reflect.StructField) string {
		name := strings.SplitN(field.Tag.Get("json"), ",", 2)[0]
		if name == "" || name == "-" {
			return field.Name
		}
		return name
	})

	_ = v.RegisterValidation("nospace", func(fl validator.FieldLevel) bool {
		return !strings.ContainsAny(fl.Field().String(), " \t\n\r")
	})
	_ = v.RegisterValidation("username", func(fl validator.FieldLevel) bool {
		for _, r := range fl.Field().String() {
			if !unicode.IsLetter(r) && !unicode.IsDigit(r) &&
				r != '.' && r != '_' {
				return false
			}
		}
		return true
	})
	_ = v.RegisterValidation("strongpassword", func(fl validator.FieldLevel) bool {
		return passwordStrength(fl.Field().String()) == ""
	})
	_ = v.RegisterValidation("nim", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		if len(val) != 12 {
			return false
		}
		for _, r := range val {
			if !unicode.IsDigit(r) {
				return false
			}
		}
		return true
	})
	_ = v.RegisterValidation("angkatan", func(fl validator.FieldLevel) bool {
		val := int(fl.Field().Int())
		currentYear := time.Now().Year()
		return val >= 1000 && val <= 9999 && val <= currentYear
	})
	_ = v.RegisterValidation("tahun_akademik", func(fl validator.FieldLevel) bool {
		return tahunAkademikRegex.MatchString(fl.Field().String())
	})
	return v
}

// ValidateStruct menjalankan seluruh aturan pada tag struct.
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
	case "strongpassword":
		if value, ok := fe.Value().(string); ok {
			return passwordStrength(value)
		}
		return "password tidak memenuhi syarat"
	case "nim":
		return "NIM harus berupa 12 digit angka"
	case "len":
		return "harus tepat " + fe.Param() + " karakter"
	case "angkatan":
		return "angkatan harus 4 digit dan tidak melebihi tahun berjalan"
	case "tahun_akademik":
		return "format tahun akademik tidak valid (contoh: 2026/2027-Ganjil)"
	case "oneof":
		return "harus salah satu dari: " +
			strings.ReplaceAll(fe.Param(), " ", ", ")
	default:
		return "tidak memenuhi aturan " + fe.Tag()
	}
}

// passwordStrength memeriksa tingkat keamanan password.
func passwordStrength(pwd string) string {
	if len(pwd) < 8 {
		return "minimal 8 karakter"
	}
	var hasLetter, hasDigit bool
	for _, r := range pwd {
		switch {
		case unicode.IsLetter(r):
			hasLetter = true
		case unicode.IsDigit(r):
			hasDigit = true
		}
	}
	if !hasLetter || !hasDigit {
		return "harus memuat huruf dan angka"
	}
	weak := map[string]bool{
		"password1": true, "12345678": true, "qwerty123": true,
		"admin123": true, "password123": true,
	}
	if weak[strings.ToLower(pwd)] {
		return "password terlalu umum"
	}
	return ""
}
