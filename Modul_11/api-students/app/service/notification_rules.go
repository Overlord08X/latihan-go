package service

import (
	"errors"
	"fmt"
	"strconv"

	"api-students/app/model"
)

var errUnknownEvent = errors.New("jenis event tidak dikenal")

// BuildMessage menyusun pesan dari sebuah event. Fungsi murni: tidak
// menyentuh database, Redis, maupun email, sehingga mudah diuji.
func BuildMessage(eventType string, e model.EnrollmentEvent, user model.User) (model.Message, error) {
	course := e.CourseCode + " " + e.CourseName
	var subject, body string

	switch eventType {
	case model.EventEnrollmentCreated:
		subject = "KRS: berhasil terdaftar di " + course
		body = fmt.Sprintf("Halo %s, Anda tercatat sebagai peserta kelas %s.", user.Username, course)
	case model.EventEnrollmentDropped:
		subject = "KRS: keluar dari " + course
		body = fmt.Sprintf("Halo %s, Anda sudah tidak terdaftar di kelas %s.", user.Username, course)
	case model.EventEnrollmentSwapped:
		from := e.FromCourseCode + " " + e.FromCourseName
		subject = "KRS: pindah dari " + from + " ke " + course
		body = fmt.Sprintf("Halo %s, Anda telah pindah dari kelas %s ke kelas %s.", user.Username, from, course)
	default:
		return model.Message{}, fmt.Errorf("%w: %q", errUnknownEvent, eventType)
	}

	return model.Message{To: user.Email, Subject: subject, Body: body}, nil
}

// BuildQuotaChangedMessage menyusun pesan untuk event perubahan kuota kelas (Tugas Mandiri D.2).
func BuildQuotaChangedMessage(e model.CourseQuotaChangedEvent, user model.User) (model.Message, error) {
	subject := "KRS: kuota " + e.CourseCode + " " + e.CourseName + " berubah menjadi " + strconv.Itoa(e.Quota)
	body := fmt.Sprintf("Halo %s, kuota kelas %s %s telah diubah menjadi %d.", user.Username, e.CourseCode, e.CourseName, e.Quota)
	return model.Message{To: user.Email, Subject: subject, Body: body}, nil
}
