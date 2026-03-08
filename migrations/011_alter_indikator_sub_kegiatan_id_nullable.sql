-- Ensure indikator_sub_kegiatan_id is nullable in existing environments.
-- This migration is idempotent and safe to re-run.

-- rencana_kerja.indikator_sub_kegiatan_id -> NULL
SET @rk_nullable := (
    SELECT IS_NULLABLE
    FROM INFORMATION_SCHEMA.COLUMNS
    WHERE TABLE_SCHEMA = DATABASE()
      AND TABLE_NAME = 'rencana_kerja'
      AND COLUMN_NAME = 'indikator_sub_kegiatan_id'
    LIMIT 1
);

SET @ddl := IF(
    @rk_nullable = 'NO',
    'ALTER TABLE rencana_kerja MODIFY COLUMN indikator_sub_kegiatan_id BIGINT UNSIGNED NULL',
    'SELECT 1'
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- indikator_rencana_kerja.indikator_sub_kegiatan_id -> NULL
SET @irk_nullable := (
    SELECT IS_NULLABLE
    FROM INFORMATION_SCHEMA.COLUMNS
    WHERE TABLE_SCHEMA = DATABASE()
      AND TABLE_NAME = 'indikator_rencana_kerja'
      AND COLUMN_NAME = 'indikator_sub_kegiatan_id'
    LIMIT 1
);

SET @ddl := IF(
    @irk_nullable = 'NO',
    'ALTER TABLE indikator_rencana_kerja MODIFY COLUMN indikator_sub_kegiatan_id BIGINT UNSIGNED NULL',
    'SELECT 1'
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
