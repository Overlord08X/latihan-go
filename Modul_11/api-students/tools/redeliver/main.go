package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"time"

	"api-students/app/model"
	"api-students/config"
	"api-students/database"

	"github.com/hibiken/asynq"
)

func main() {
	outboxID := flag.Int64("outbox", 0, "ID outbox yang akan dikirim ulang")
	copies := flag.Int("copies", 5, "Jumlah salinan task yang akan di-enqueue")
	flag.Parse()

	if *outboxID <= 0 {
		log.Fatalf("flag -outbox wajib diisi dengan ID yang valid (misal: -outbox 12)")
	}
	if *copies <= 0 {
		log.Fatalf("flag -copies minimal 1")
	}

	config.LoadEnv()
	cfg := config.Load()

	pool, err := database.Connect(cfg.DBDSN)
	if err != nil {
		log.Fatalf("gagal koneksi ke database: %v", err)
	}
	defer pool.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var id int64
	var eventType string
	var payload []byte
	err = pool.QueryRow(ctx, "SELECT id, event_type, payload FROM outbox WHERE id = $1", *outboxID).Scan(&id, &eventType, &payload)
	if err != nil {
		log.Fatalf("outbox id %d tidak ditemukan di database: %v", *outboxID, err)
	}

	job := model.NotificationJob{
		OutboxID:  id,
		EventType: eventType,
		Payload:   payload,
	}

	data, err := json.Marshal(job)
	if err != nil {
		log.Fatalf("gagal serialize job: %v", err)
	}

	client := asynq.NewClient(config.RedisOpt())
	defer client.Close()

	for i := 0; i < *copies; i++ {
		task := asynq.NewTask(model.TaskSendNotification, data)
		info, err := client.Enqueue(task,
			asynq.Queue(model.QueueNotifications),
			asynq.MaxRetry(5),
			asynq.Timeout(30*time.Second),
		)
		if err != nil {
			log.Fatalf("gagal enqueue task ke-%d: %v", i+1, err)
		}
		fmt.Printf("[%d/%d] Berhasil enqueue task: ID=%s Queue=%s OutboxID=%d Event=%s\n",
			i+1, *copies, info.ID, info.Queue, id, eventType)
	}

	fmt.Printf("Selesai memasukkan %d salinan task untuk outbox ID %d ke Redis queue '%s'.\n",
		*copies, id, model.QueueNotifications)
}
