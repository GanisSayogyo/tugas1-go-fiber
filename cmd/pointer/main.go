package main

import "fmt"

// swap menerima alamat dua integer dan menukar nilai aslinya
func swap(a, b *int) {
	*a, *b = *b, *a
}

// updateSlice menerima alamat slice dan menambahkan item baru
func updateSlice(s *[]string, newItem string) {
	*s = append(*s, newItem)
}

// changeByValue hanya mengubah salinan nilai
func changeByValue(value int) {
	value = 100
	fmt.Println("Di dalam changeByValue:", value)
}

// changeByPointer mengubah nilai asli melalui alamatnya
func changeByPointer(value *int) {
	*value = 100
	fmt.Println("Di dalam changeByPointer:", *value)
}

func main() {
	fmt.Println("=== Fungsi swap ===")
	angkaPertama := 10
	angkaKedua := 20

	fmt.Println("Sebelum swap:", angkaPertama, angkaKedua)
	swap(&angkaPertama, &angkaKedua)
	fmt.Println("Setelah swap:", angkaPertama, angkaKedua)

	fmt.Println("\n=== Fungsi updateSlice ===")
	teknologi := []string{"Go", "Fiber"}

	fmt.Println("Sebelum updateSlice:", teknologi)
	updateSlice(&teknologi, "PostgreSQL")
	fmt.Println("Setelah updateSlice:", teknologi)

	fmt.Println("\n=== Pass by Value dan Pointer ===")
	nilai := 42

	fmt.Println("Nilai awal:", nilai)
	changeByValue(nilai)
	fmt.Println("Setelah changeByValue:", nilai)

	changeByPointer(&nilai)
	fmt.Println("Setelah changeByPointer:", nilai)
}
