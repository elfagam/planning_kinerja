-- 037_alter_qna_notifications_add_fk_cascade.sql
-- Menambahkan Foreign Key Cascade agar notifikasi terhapus otomatis saat pertanyaan dihapus

-- Bersihkan data yatim (orphaned data) jika ada sebelumnya (opsional tapi aman)
DELETE FROM `qna_notifications` WHERE `question_id` NOT IN (SELECT `id` FROM `questions`);

-- Tambahkan constraint FK dengan ON DELETE CASCADE
-- Kita drop dulu yang lama (kita tahu ini ada karena error Duplicate key sebelumnya)
-- Jika ini gagal di environment lain, abaikan saja atau buat idempotent.
ALTER TABLE `qna_notifications` DROP CONSTRAINT `fk_notifications_question_id`;

ALTER TABLE `qna_notifications`
ADD CONSTRAINT `fk_notifications_question_id`
FOREIGN KEY (`question_id`) REFERENCES `questions`(`id`)
ON DELETE CASCADE;
