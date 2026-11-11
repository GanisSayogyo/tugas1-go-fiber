package service

import "testing"

func TestWithinCreditLimit(t *testing.T) {
	tests := []struct {
		name   string
		taken  int
		adding int
		want   bool
	}{
		{"masih di bawah batas", 18, 3, true},
		{"tepat 24 SKS", 21, 3, true},
		{"melewati batas", 21, 6, false},
		{"sudah mencapai batas", 24, 3, false},
		{"belum mengambil SKS", 0, 6, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WithinCreditLimit(tt.taken, tt.adding)

			if got != tt.want {
				t.Errorf(
					"WithinCreditLimit(%d, %d) = %v, want %v",
					tt.taken,
					tt.adding,
					got,
					tt.want,
				)
			}
		})
	}
}

func TestCreditsAfterSwap(t *testing.T) {
	tests := []struct {
		name    string
		taken   int
		leaving int
		joining int
		want    int
	}{
		{"pindah kelas SKS sama", 18, 3, 3, 18},
		{"pindah ke kelas lebih besar", 18, 3, 6, 21},
		{"pindah ke kelas lebih kecil", 18, 6, 3, 15},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CreditsAfterSwap(
				tt.taken,
				tt.leaving,
				tt.joining,
			)

			if got != tt.want {
				t.Errorf(
					"CreditsAfterSwap(%d, %d, %d) = %d, want %d",
					tt.taken,
					tt.leaving,
					tt.joining,
					got,
					tt.want,
				)
			}
		})
	}
}
