package service

import "testing"

func TestValidatePasswordStrength(t *testing.T) {
	tests := []struct {
		name     string
		password string
		want     bool
	}{
		{
			name:     "valid password",
			password: "Password123",
			want:     true,
		},
		{
			name:     "too short",
			password: "Pass1",
			want:     false,
		},
		{
			name:     "without digit",
			password: "Password",
			want:     false,
		},
		{
			name:     "without letter",
			password: "12345678",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidatePasswordStrength(tt.password)

			if got != tt.want {
				t.Errorf(
					"ValidatePasswordStrength(%q) = %v, want %v",
					tt.password,
					got,
					tt.want,
				)
			}
		})
	}
}