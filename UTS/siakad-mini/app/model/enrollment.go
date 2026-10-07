package model

import "time"

type Enrollment struct {
	ID            int64     `json:"id"`
	StudentID     int64     `json:"student_id"`
	CourseID      int64     `json:"course_id"`
	TahunAkademik string    `json:"tahun_akademik"`
	CreatedAt     time.Time `json:"created_at"`
}
