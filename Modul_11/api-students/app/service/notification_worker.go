package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"api-students/app/model"
	"api-students/app/repository"
	"api-students/helper"

	"github.com/hibiken/asynq"
)

// NotificationWorker mengerjakan task "notification:send".
//
// Bagi worker, Asynq adalah framework-nya — sama seperti Fiber bagi API.
// Karena itu method di sini menerima *asynq.Task, seperti method service
// lain menerima *fiber.Ctx.
type NotificationWorker struct {
	users         repository.UserRepository
	notifications repository.NotificationRepository
	tx            *repository.TxManager
	mailer        helper.Mailer
	logger        *slog.Logger
}

func NewNotificationWorker(
	users repository.UserRepository,
	notifications repository.NotificationRepository,
	tx *repository.TxManager,
	mailer helper.Mailer,
	logger *slog.Logger,
) *NotificationWorker {
	return &NotificationWorker{
		users:         users,
		notifications: notifications,
		tx:            tx,
		mailer:        mailer,
		logger:        logger,
	}
}

// HandleSend mengembalikan error untuk meminta Asynq melakukan retry nanti,
// atau error yang membungkus asynq.SkipRetry untuk menyatakan "jangan di-retry,
// task ini tidak mungkin berhasil".
//
// Perbaikan Bug #2: Parameter task bertipe pointer (*asynq.Task), bukan value (asynq.Task),
// agar sesuai dengan kontrak asynq.HandlerFunc / asynq.ServeMux.
func (w *NotificationWorker) HandleSend(ctx context.Context, t *asynq.Task) error {
	var job model.NotificationJob
	if err := json.Unmarshal(t.Payload(), &job); err != nil {
		return fmt.Errorf("%w: payload task rusak: %v", asynq.SkipRetry, err)
	}

	// Dukungan Tugas Mandiri D.2: Penanganan event course.quota_changed untuk banyak mahasiswa
	if job.EventType == model.EventCourseQuotaChanged {
		return w.handleQuotaChanged(ctx, job)
	}

	var event model.EnrollmentEvent
	if err := json.Unmarshal(job.Payload, &event); err != nil {
		return fmt.Errorf("%w: payload event %d rusak: %v", asynq.SkipRetry, job.OutboxID, err)
	}

	user, err := w.users.FindByID(ctx, event.UserID)
	if errors.Is(err, repository.ErrNotFound) {
		return fmt.Errorf("%w: user %d tidak ditemukan", asynq.SkipRetry, event.UserID)
	}
	if err != nil {
		return err // database sedang bermasalah: layak di-retry
	}

	msg, err := BuildMessage(job.EventType, event, *user)
	if err != nil {
		return fmt.Errorf("%w: %v", asynq.SkipRetry, err)
	}

	alreadySent := false
	err = w.tx.WithTxOnce(ctx, func(tx repository.DBTX) error {
		// Claim dan pengiriman berada di dalam SATU transaction:
		// - pengiriman gagal -> claim ikut dibatalkan -> attempt berikutnya masih dapat mengirim;
		// - pengiriman sukses -> claim di-commit -> attempt berikutnya melihat claim ini dan tidak mengirim lagi.
		// Perbaikan Bug #8: Panggil w.notifications.WithTx(tx), BUKAN w.notifications langsung tanpa WithTx!
		claimed, err := w.notifications.WithTx(tx).Claim(ctx, model.Notification{
			OutboxID: job.OutboxID,
			UserID:   user.ID,
			Channel:  "email",
			Subject:  msg.Subject,
			Body:     msg.Body,
		})
		if err != nil {
			return err
		}
		if !claimed {
			alreadySent = true
			return nil
		}
		return w.mailer.Send(ctx, msg)
	})

	// Perbaikan Bug #9: Bungkus asynq.SkipRetry dengan %w, BUKAN %v!
	// fmt.Errorf("%w: %v", asynq.SkipRetry, err) agar errors.Is(returnedErr, asynq.SkipRetry) bernilai true.
	if errors.Is(err, helper.ErrPermanentDelivery) {
		return fmt.Errorf("%w: %v", asynq.SkipRetry, err)
	}
	if err != nil {
		return err
	}

	retry, _ := asynq.GetRetryCount(ctx)
	w.logger.Info("notification_processed",
		slog.Int64("outbox_id", job.OutboxID),
		slog.String("event", job.EventType),
		slog.Int("retry", retry),
		slog.Bool("already_sent", alreadySent))
	return nil
}

// handleQuotaChanged menangani pengiriman email ke seluruh peserta kelas (Tugas Mandiri D.2).
// Bila pengiriman ke mahasiswa ke-N gagal, retry tidak akan mengirim ulang ke mahasiswa 1..(N-1).
func (w *NotificationWorker) handleQuotaChanged(ctx context.Context, job model.NotificationJob) error {
	var event model.CourseQuotaChangedEvent
	if err := json.Unmarshal(job.Payload, &event); err != nil {
		return fmt.Errorf("%w: payload quota_changed %d rusak: %v", asynq.SkipRetry, job.OutboxID, err)
	}

	retry, _ := asynq.GetRetryCount(ctx)
	anySent := false

	for _, userID := range event.StudentUserIDs {
		user, err := w.users.FindByID(ctx, userID)
		if errors.Is(err, repository.ErrNotFound) {
			continue
		}
		if err != nil {
			return err
		}

		msg, err := BuildQuotaChangedMessage(event, *user)
		if err != nil {
			return fmt.Errorf("%w: %v", asynq.SkipRetry, err)
		}

		alreadySent := false
		err = w.tx.WithTxOnce(ctx, func(tx repository.DBTX) error {
			claimed, err := w.notifications.WithTx(tx).Claim(ctx, model.Notification{
				OutboxID: job.OutboxID,
				UserID:   user.ID,
				Channel:  "email",
				Subject:  msg.Subject,
				Body:     msg.Body,
			})
			if err != nil {
				return err
			}
			if !claimed {
				alreadySent = true
				return nil
			}
			return w.mailer.Send(ctx, msg)
		})

		if errors.Is(err, helper.ErrPermanentDelivery) {
			w.logger.Warn("task_failed",
				slog.String("type", model.TaskSendNotification),
				slog.String("attempt", fmt.Sprintf("%d/6", retry+1)),
				slog.String("error", err.Error()))
			continue
		}
		if err != nil {
			return err // Server email sibuk atau error sementara: kembalikan error agar Asynq retry
		}

		if !alreadySent {
			anySent = true
		}

		w.logger.Info("notification_processed",
			slog.Int64("outbox_id", job.OutboxID),
			slog.String("event", job.EventType),
			slog.Int("user_id", user.ID),
			slog.Int("retry", retry),
			slog.Bool("already_sent", alreadySent))
	}

	_ = anySent
	return nil
}
