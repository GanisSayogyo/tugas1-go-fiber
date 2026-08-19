package main

import "fmt"

func main() {
	// Lima variabel dengan tipe data berbeda
	nama := "Muhammad Ganis Sayogyo"
	umur := 20
	ipk := 3.31
	mahasiswaAktif := true
	keahlian := []string{"3D Modelling", "Backend Developing", "AI Engineering"}

	fmt.Println("=== Data Diri ===")
	fmt.Println("Nama Mahasiswa:", nama)
	fmt.Println("Umur:", umur)
	fmt.Println("IPK:", ipk)
	fmt.Println("Mahasiswa aktif:", mahasiswaAktif)
	fmt.Println("Keahlian:", keahlian)

	// Nama mahasiswa menjadi key dan nilainya menjadi value
	nilaiMahasiswa := map[string]float64{
		"Ganis": 90.3,
		"Farel": 88.9,
		"Dinda": 89.1,
	}

	// Menambahkan data baru ke map
	nilaiMahasiswa["Rehan"] = 87.0
	fmt.Println("\nData setelah Budi ditambahkan:", nilaiMahasiswa)

	// Membaca data dengan pengecekan keberadaan key
	nilaiGanis, ditemukan := nilaiMahasiswa["Ganis"]
	if ditemukan {
		fmt.Println("Nilai Ganis:", nilaiGanis)
	} else {
		fmt.Println("Data Ganis tidak ditemukan")
	}

	// Menghapus data dari map
	delete(nilaiMahasiswa, "Farel")
	fmt.Println("Data setelah Farel dihapus:", nilaiMahasiswa)

	// Menelusuri seluruh isi map
	fmt.Println("\n=== Nilai Mahasiswa Keseluruhan ===")
	for namaMahasiswa, nilai := range nilaiMahasiswa {
		fmt.Printf("%s: %.1f\n", namaMahasiswa, nilai)
	}
}
