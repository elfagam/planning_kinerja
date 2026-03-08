-- Drop indikator_sub_kegiatan_id from indikator_rencana_kerja safely.
-- This migration is idempotent.

SET @tbl_exists := (
    SELECT COUNT(1)
    FROM INFORMATION_SCHEMA.TABLES
    WHERE TABLE_SCHEMA = DATABASE()
      AND TABLE_NAME = 'indikator_rencana_kerja'
);

SET @fk_exists := (
    SELECT COUNT(1)
    FROM INFORMATION_SCHEMA.TABLE_CONSTRAINTS
    WHERE TABLE_SCHEMA = DATABASE()
      AND TABLE_NAME = 'indikator_rencana_kerja'
      AND CONSTRAINT_TYPE = 'FOREIGN KEY'
      AND CONSTRAINT_NAME = 'fk_id_indikator_rk_indikator_sub_kegiatan'
);

SET @ddl := IF(
    @tbl_exists = 1 AND @fk_exists = 1,
    'ALTER TABLE indikator_rencana_kerja DROP FOREIGN KEY fk_id_indikator_rk_indikator_sub_kegiatan',
    'SELECT 1'
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
    SELECT COUNT(1)
    FROM INFORMATION_SCHEMA.STATISTICS
    WHERE TABLE_SCHEMA = DATABASE()
      AND TABLE_NAME = 'indikator_rencana_kerja'
      AND INDEX_NAME = 'idx_indikator_rk_indikator_sub_kegiatan'
);

SET @ddl := IF(
    @tbl_exists = 1 AND @idx_exists = 1,
    'ALTER TABLE indikator_rencana_kerja DROP INDEX idx_indikator_rk_indikator_sub_kegiatan',
    'SELECT 1'
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (
    SELECT COUNT(1)
    FROM INFORMATION_SCHEMA.COLUMNS
    WHERE TABLE_SCHEMA = DATABASE()
      AND TABLE_NAME = 'indikator_rencana_kerja'
      AND COLUMN_NAME = 'indikator_sub_kegiatan_id'
);

SET @ddl := IF(
    @tbl_exists = 1 AND @col_exists = 1,
    'ALTER TABLE indikator_rencana_kerja DROP COLUMN indikator_sub_kegiatan_id',
    'SELECT 1'
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
