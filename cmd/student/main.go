package main

import "fmt"

// Student menyimpan informasi seorang mahasiswa.
type Student struct {
	ID       string
	Name     string
	Grade    float64
	IsActive bool
}

// GetInfo memakai value receiver karena hanya membaca data.
func (s Student) GetInfo() string {
	return fmt.Sprintf(
		"ID: %s | Name: %s | Grade: %.2f | Active: %t",
		s.ID,
		s.Name,
		s.Grade,
		s.IsActive,
	)
}

// UpdateGrade memakai pointer receiver karena mengubah nilai Grade.
func (s *Student) UpdateGrade(grade float64) {
	s.Grade = grade
}

// Activate memakai pointer receiver karena mengubah status mahasiswa.
func (s *Student) Activate() {
	s.IsActive = true
}

// Deactivate memakai pointer receiver karena mengubah status mahasiswa.
func (s *Student) Deactivate() {
	s.IsActive = false
}

func main() {
	student := Student{
		ID:       "STD001",
		Name:     "Muhammad Ganis Sayogyo",
		Grade:    85.5,
		IsActive: false,
	}

	fmt.Println("Data awal:")
	fmt.Println(student.GetInfo())

	student.Activate()
	fmt.Println("\nSetelah Activate:")
	fmt.Println(student.GetInfo())

	student.UpdateGrade(92.5)
	fmt.Println("\nSetelah UpdateGrade:")
	fmt.Println(student.GetInfo())

	student.Deactivate()
	fmt.Println("\nSetelah Deactivate:")
	fmt.Println(student.GetInfo())
}
