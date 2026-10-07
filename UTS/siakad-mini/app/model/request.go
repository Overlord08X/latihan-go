package model

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type CreateStudentRequest struct {
	NIM         string   `json:"nim" validate:"required,nim"`
	Nama        string   `json:"nama" validate:"required,min=3"`
	Email       string   `json:"email" validate:"required,email"`
	Prodi       string   `json:"prodi" validate:"required"`
	Angkatan    int      `json:"angkatan" validate:"required,angkatan_valid"`
	IPKTerakhir *float64 `json:"ipk_terakhir,omitempty" validate:"omitempty,gte=0,lte=4"`
}

type UpdateStudentRequest struct {
	Nama        string   `json:"nama" validate:"required,min=3"`
	Prodi       string   `json:"prodi" validate:"required"`
	Angkatan    int      `json:"angkatan" validate:"required,angkatan_valid"`
	IPKTerakhir *float64 `json:"ipk_terakhir,omitempty" validate:"omitempty,gte=0,lte=4"`
}

type CreateEnrollmentRequest struct {
	CourseID      int64  `json:"course_id" validate:"required,gt=0"`
	TahunAkademik string `json:"tahun_akademik" validate:"required,tahun_akademik"`
}

type StudentFilterQuery struct {
	Page     int    `query:"page"`
	PerPage  int    `query:"per_page"`
	Prodi    string `query:"prodi"`
	Angkatan int    `query:"angkatan"`
	Search   string `query:"search"`
	Sort     string `query:"sort"`
}

type CourseFilterQuery struct {
	Semester  int    `query:"semester"`
	Search    string `query:"search"`
	Available bool   `query:"available"`
}
