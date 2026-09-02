package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// LoadEnv memuat variabel dari file .env.
// Jika .env tidak ditemukan, aplikasi tetap mencoba
// menggunakan environment system.
func LoadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Println("peringatan: berkas .env tidak ditemukan, memakai environment system")
	}
}

// GetEnv mengambil nilai environment.
// Jika kosong, gunakan nilai fallback.
func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}

	return fallback
}

// GetEnvInt mengambil environment sebagai integer.
// Jika tidak ada atau bukan angka, gunakan fallback.
func GetEnvInt(key string, fallback int) int {
	value := GetEnv(key, "")

	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		log.Printf(
			"peringatan: %s bukan angka (%q), memakai bawaan %d",
			key,
			value,
			fallback,
		)

		return fallback
	}

	return parsed
}