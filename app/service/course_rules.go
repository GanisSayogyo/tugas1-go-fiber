package service

// MaxCreditsPerSemester adalah batas SKS per semester.
const MaxCreditsPerSemester = 24

// WithinCreditLimit memeriksa apakah penambahan SKS masih diperbolehkan.
func WithinCreditLimit(taken, adding int) bool {
	return taken+adding <= MaxCreditsPerSemester
}

// CreditsAfterSwap menghitung total SKS setelah pindah kelas.
func CreditsAfterSwap(taken, leaving, joining int) int {
	return taken - leaving + joining
}
