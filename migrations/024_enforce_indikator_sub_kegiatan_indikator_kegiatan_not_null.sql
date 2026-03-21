-- Enforce indikator_sub_kegiatan.indikator_kegiatan_id as NOT NULL
-- Prerequisite: migration 023 has run successfully and backfill is complete.

-- 1) Drop old FK (SET NULL) if present
SET @has_fk := (
  SELECT COUNT(*)
  FROM information_schema.REFERENTIAL_CONSTRAINTS
  WHERE CONSTRAINT_SCHEMA = DATABASE()
    AND TABLE_NAME = 'indikator_sub_kegiatan'
    AND CONSTRAINT_NAME = 'fk_indikator_sub_kegiatan_indikator_kegiatan'
);

SET @sql := IF(
  @has_fk = 1,
  'ALTER TABLE indikator_sub_kegiatan DROP FOREIGN KEY fk_indikator_sub_kegiatan_indikator_kegiatan',
  'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 2) Enforce NOT NULL (will fail if unresolved NULL values still exist)
ALTER TABLE indikator_sub_kegiatan
  MODIFY COLUMN indikator_kegiatan_id BIGINT UNSIGNED NOT NULL;

-- 3) Ensure index exists
SET @has_idx := (
  SELECT COUNT(*)
  FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'indikator_sub_kegiatan'
    AND INDEX_NAME = 'idx_indikator_sub_kegiatan_indikator_kegiatan'
);

SET @sql := IF(
  @has_idx = 0,
  'ALTER TABLE indikator_sub_kegiatan ADD INDEX idx_indikator_sub_kegiatan_indikator_kegiatan (indikator_kegiatan_id)',
  'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 4) Re-create FK compatible with NOT NULL
SET @has_fk_new := (
  SELECT COUNT(*)
  FROM information_schema.REFERENTIAL_CONSTRAINTS
  WHERE CONSTRAINT_SCHEMA = DATABASE()
    AND TABLE_NAME = 'indikator_sub_kegiatan'
    AND CONSTRAINT_NAME = 'fk_indikator_sub_kegiatan_indikator_kegiatan'
);

SET @sql := IF(
  @has_fk_new = 0,
  'ALTER TABLE indikator_sub_kegiatan ADD CONSTRAINT fk_indikator_sub_kegiatan_indikator_kegiatan FOREIGN KEY (indikator_kegiatan_id) REFERENCES indikator_kegiatan(id) ON UPDATE CASCADE ON DELETE RESTRICT',
  'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
