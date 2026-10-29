package model

import "time"

// Review adalah bentuk ulasan yang dikenal seluruh aplikasi.
type Review struct {
	ID           string    `json:"id"`
	CourseID     int       `json:"course_id"`
	UserID       int       `json:"user_id"`
	Username     string    `json:"username"`
	Rating       int       `json:"rating"`
	Comment      string    `json:"comment"`
	Tags         []string  `json:"tags"`
	Replies      []Reply   `json:"replies"`
	HelpfulCount int       `json:"helpful_count"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Reply disimpan DI DALAM dokumen Review (embedding), bukan di collection terpisah.
type Reply struct {
	ID        string    `json:"id"`
	UserID    int       `json:"user_id"`
	Username  string    `json:"username"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateReviewRequest struct {
	Rating  int      `json:"rating" validate:"required,min=1,max=5"`
	Comment string   `json:"comment" validate:"required,min=10,max=1000"`
	Tags    []string `json:"tags" validate:"max=5,dive,min=2,max=30"`
}

// UpdateReviewRequest memakai pointer seperti PATCH pada Modul 2 dan Modul 7.
type UpdateReviewRequest struct {
	Rating  *int      `json:"rating,omitempty" validate:"omitnil,min=1,max=5"`
	Comment *string   `json:"comment,omitempty" validate:"omitnil,min=10,max=1000"`
	Tags    *[]string `json:"tags,omitempty" validate:"omitnil,max=5,dive,min=2,max=30"`
}

type ReplyRequest struct {
	Body string `json:"body" validate:"required,min=2,max=500"`
}

// ReviewQuery adalah parameter daftar ulasan satu kelas.
type ReviewQuery struct {
	CourseID int
	Tag      string
	Before   string // id ulasan terakhir yang sudah dilihat client
	Limit    int
}

type ReviewSummary struct {
	CourseID      int            `json:"course_id"`
	Count         int            `json:"count"`
	AverageRating float64        `json:"average_rating"`
	Distribution  map[string]int `json:"distribution"`
	TopTags       []TagCount     `json:"top_tags"`
}

type TagCount struct {
	Tag   string `json:"tag"`
	Total int    `json:"total"`
}
