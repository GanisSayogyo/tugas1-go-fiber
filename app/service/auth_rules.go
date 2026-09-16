package service

import (
	"errors"
	"strings"
	"unicode"
)

func ValidatePasswordStrength(password string) bool {
	if len(password) < 8 {
		return false
	}

	hasLetter := false
	hasDigit := false

	for _, char := range password {
		if unicode.IsLetter(char) {
			hasLetter = true
		}

		if unicode.IsDigit(char) {
			hasDigit = true
		}
	}

	return hasLetter && hasDigit
}

func ValidateRegister(username, email, password string) error {
	if strings.TrimSpace(username) == "" {
		return errors.New("username wajib diisi")
	}

	if strings.TrimSpace(email) == "" {
		return errors.New("email wajib diisi")
	}

	if !ValidatePasswordStrength(password) {
		return errors.New("password minimal 8 karakter dan harus mengandung huruf dan angka")
	}

	return nil
}