-- Reconcile migration for environments that previously ran an older/broken
-- variant of 007_requested_schema_tables_id.sql.
-- This migration is idempotent and safe to re-run.

-- ------------------------------
-- rencana_kerja reconciliation
-- ------------------------------
SET @rk_table_exists := (
    SELECT COUNT(1)
    FROM INFORMATION_SCHEMA.TABLES
    WHERE TABLE_SCHEMA = DATABASE()
      AND TABLE_NAME = 'rencana_kerja'
);

SET @rk_col_exists := (
    SELECT COUNT(1)
    FROM INFORMATION_SCHEMA.COLUMNS
    WHERE TABLE_SCHEMA = DATABASE()
      AND TABLE_NAME = 'rencana_kerja'
      AND COLUMN_NAME = 'indikator_sub_kegiatan_id'
);

SET @ddl := IF(
    @rk_table_exists = 1 AND @rk_col_exists = 0,
    'ALTER TABLE rencana_kerja ADD COLUMN indikator_sub_kegiatan_id BIGINT UNSIGNED NULL',
    'SELECT 1'
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @rk_idx_exists := (
    SELECT COUNT(1)
    FROM INFORMATION_SCHEMA.STATISTICS
    WHERE TABLE_SCHEMA = DATABASE()
      AND TABLE_NAME = 'rencana_kerja'
      AND INDEX_NAME = 'idx_rencana_kerja_indikator_sub_kegiatan'
);

SET @ddl := IF(
    @rk_table_exists = 1 AND @rk_idx_exists = 0,
    'ALTER TABLE rencana_kerja ADD INDEX idx_rencana_kerja_indikator_sub_kegiatan (indikator_sub_kegiatan_id)',
    'SELECT 1'
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @rk_fk_exists := (
    SELECT COUNT(1)
    FROM INFORMATION_SCHEMA.TABLE_CONSTRAINTS
    WHERE TABLE_SCHEMA = DATABASE()
      AND TABLE_NAME = 'rencana_kerja'
      AND CONSTRAINT_TYPE = 'FOREIGN KEY'
      AND CONSTRAINT_NAME = 'fk_id_rencana_kerja_indikator_sub_kegiatan'
);

SET @ddl := IF(
    @rk_table_exists = 1 AND @rk_fk_exists = 0,
    'ALTER TABLE rencana_kerja ADD CONSTRAINT fk_id_rencana_kerja_indikator_sub_kegiatan FOREIGN KEY (indikator_sub_kegiatan_id) REFERENCES indikator_sub_kegiatan(id)',
    'SELECT 1'
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- ------------------------------
-- indikator_rencana_kerja reconciliation
-- ------------------------------
SET @irk_table_exists := (
    SELECT COUNT(1)
    FROM INFORMATION_SCHEMA.TABLES
    WHERE TABLE_SCHEMA = DATABASE()
      AND TABLE_NAME = 'indikator_rencana_kerja'
);

SET @irk_rk_col_exists := (
    SELECT COUNT(1)
    FROM INFORMATION_SCHEMA.COLUMNS
    WHERE TABLE_SCHEMA = DATABASE()
      AND TABLE_NAME = 'indikator_rencana_kerja'
      AND COLUMN_NAME = 'rencana_kerja_id'
);

SET @ddl := IF(
    @irk_table_exists = 1 AND @irk_rk_col_exists = 0,
    'ALTER TABLE indikator_rencana_kerja ADD COLUMN rencana_kerja_id BIGINT UNSIGNED NULL',
    'SELECT 1'
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @irk_isk_col_exists := (
    SELECT COUNT(1)
    FROM INFORMATION_SCHEMA.COLUMNS
    WHERE TABLE_SCHEMA = DATABASE()
      AND TABLE_NAME = 'indikator_rencana_kerja'
      AND COLUMN_NAME = 'indikator_sub_kegiatan_id'
);

SET @ddl := IF(
    @irk_table_exists = 1 AND @irk_isk_col_exists = 0,
    'ALTER TABLE indikator_rencana_kerja ADD COLUMN indikator_sub_kegiatan_id BIGINT UNSIGNED NULL',
    'SELECT 1'
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @irk_rk_idx_exists := (
    SELECT COUNT(1)
    FROM INFORMATION_SCHEMA.STATISTICS
    WHERE TABLE_SCHEMA = DATABASE()
      AND TABLE_NAME = 'indikator_rencana_kerja'
      AND INDEX_NAME = 'idx_indikator_rk_rk'
);

SET @ddl := IF(
    @irk_table_exists = 1 AND @irk_rk_idx_exists = 0,
    'ALTER TABLE indikator_rencana_kerja ADD INDEX idx_indikator_rk_rk (rencana_kerja_id)',
    'SELECT 1'
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @irk_isk_idx_exists := (
    SELECT COUNT(1)
    FROM INFORMATION_SCHEMA.STATISTICS
    WHERE TABLE_SCHEMA = DATABASE()
      AND TABLE_NAME = 'indikator_rencana_kerja'
      AND INDEX_NAME = 'idx_indikator_rk_indikator_sub_kegiatan'
);

SET @ddl := IF(
    @irk_table_exists = 1 AND @irk_isk_idx_exists = 0,
    'ALTER TABLE indikator_rencana_kerja ADD INDEX idx_indikator_rk_indikator_sub_kegiatan (indikator_sub_kegiatan_id)',
    'SELECT 1'
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @irk_fk_rk_exists := (
    SELECT COUNT(1)
    FROM INFORMATION_SCHEMA.TABLE_CONSTRAINTS
    WHERE TABLE_SCHEMA = DATABASE()
      AND TABLE_NAME = 'indikator_rencana_kerja'
      AND CONSTRAINT_TYPE = 'FOREIGN KEY'
      AND CONSTRAINT_NAME = 'fk_id_indikator_rk_rk'
);

SET @ddl := IF(
    @irk_table_exists = 1 AND @irk_fk_rk_exists = 0,
    'ALTER TABLE indikator_rencana_kerja ADD CONSTRAINT fk_id_indikator_rk_rk FOREIGN KEY (rencana_kerja_id) REFERENCES rencana_kerja(id)',
    'SELECT 1'
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @irk_fk_isk_exists := (
    SELECT COUNT(1)
    FROM INFORMATION_SCHEMA.TABLE_CONSTRAINTS
    WHERE TABLE_SCHEMA = DATABASE()
      AND TABLE_NAME = 'indikator_rencana_kerja'
      AND CONSTRAINT_TYPE = 'FOREIGN KEY'
      AND CONSTRAINT_NAME = 'fk_id_indikator_rk_indikator_sub_kegiatan'
);

SET @ddl := IF(
    @irk_table_exists = 1 AND @irk_fk_isk_exists = 0,
    'ALTER TABLE indikator_rencana_kerja ADD CONSTRAINT fk_id_indikator_rk_indikator_sub_kegiatan FOREIGN KEY (indikator_sub_kegiatan_id) REFERENCES indikator_sub_kegiatan(id)',
    'SELECT 1'
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
