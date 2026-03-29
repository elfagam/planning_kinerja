-- 037_alter_qna_notifications_add_fk_cascade.sql
-- Menambahkan Foreign Key Cascade agar notifikasi terhapus otomatis saat pertanyaan dihapus

-- Bersihkan data yatim (orphaned data) jika ada sebelumnya (opsional tapi aman)
DELETE FROM `qna_notifications` WHERE `question_id` NOT IN (SELECT `id` FROM `questions`);

-- Tambahkan constraint FK dengan ON DELETE CASCADE
ALTER TABLE `qna_notifications`
ADD CONSTRAINT `fk_notifications_question_id`
FOREIGN KEY (`question_id`) REFERENCES `questions`(`id`)
ON DELETE CASCADE;
