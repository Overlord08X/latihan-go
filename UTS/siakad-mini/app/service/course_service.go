package service

import (
	"context"

	"siakad-mini/app/model"
	"siakad-mini/app/repository"
	"siakad-mini/helper"
)

type CourseService struct {
	courseRepo *repository.CourseRepository
}

func NewCourseService(courseRepo *repository.CourseRepository) *CourseService {
	return &CourseService{courseRepo: courseRepo}
}

func (s *CourseService) List(ctx context.Context, f model.CourseFilterQuery) ([]model.Course, error) {
	courses, err := s.courseRepo.List(ctx, f)
	if err != nil {
		return nil, helper.Internal(err, "Gagal mengambil daftar mata kuliah")
	}
	if courses == nil {
		courses = []model.Course{}
	}
	return courses, nil
}
