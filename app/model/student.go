package model

import "time"

type Student struct {
	ID          int        `json:"id"`
	UserID      int        `json:"user_id"`
	NIM         string     `json:"nim"`
	Nama        string     `json:"nama"`
	Prodi       string     `json:"prodi"`
	Angkatan    int        `json:"angkatan"`
	IPKTerakhir float64    `json:"ipk_terakhir"`
	DeletedAt   *time.Time `json:"-"`
	CreatedAt   time.Time  `json:"created_at"`
}

type CreateStudentRequest struct {
	NIM         string   `json:"nim" validate:"required,len=12"`
	Nama        string   `json:"nama" validate:"required"`
	Email       string   `json:"email" validate:"required,email"`
	Prodi       string   `json:"prodi" validate:"required"`
	Angkatan    int      `json:"angkatan" validate:"required,angkatan"`
	IPKTerakhir *float64 `json:"ipk_terakhir,omitempty" validate:"omitempty,min=0,max=4"`
}

type ReplaceStudentRequest struct {
	Nama        string   `json:"nama" validate:"required"`
	Prodi       string   `json:"prodi" validate:"required"`
	Angkatan    int      `json:"angkatan" validate:"required"`
	IPKTerakhir *float64 `json:"ipk_terakhir,omitempty" validate:"omitempty,min=0,max=4"`
}

type StudentDetail struct {
	Student
	TotalSKS int      `json:"total_sks"`
	BatasSKS int      `json:"batas_sks"`
	Courses  []Course `json:"courses"`
}
