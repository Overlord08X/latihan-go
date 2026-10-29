-- Migration 014: Mengganti idempotency key notifications dari UNIQUE(outbox_id) menjadi UNIQUE(outbox_id, user_id)
-- untuk mendukung event broadcast course.quota_changed yang menghasilkan banyak notifikasi dari satu baris outbox (Tugas Mandiri D.2).

ALTER TABLE notifications DROP CONSTRAINT IF EXISTS notifications_outbox_unique;
ALTER TABLE notifications DROP CONSTRAINT IF EXISTS notifications_outbox_user_unique;
ALTER TABLE notifications ADD CONSTRAINT notifications_outbox_user_unique UNIQUE (outbox_id, user_id);
