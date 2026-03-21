-- Migration to add sasaran_id to indikator_program table

ALTER TABLE indikator_program
ADD COLUMN sasaran_id BIGINT UNSIGNED NOT NULL AFTER id;
