-- Add direct relationship: indikator_sub_kegiatan -> indikator_kegiatan
-- Safe rollout:
-- 1) add nullable column after id
-- 2) add index
-- 3) backfill best-effort from sub_kegiatan.kegiatan_id
-- 4) add FK with ON UPDATE CASCADE, ON DELETE SET NULL

-- 1) Add column if not exists
SET @has_col := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'indikator_sub_kegiatan'
    AND COLUMN_NAME = 'indikator_kegiatan_id'
);

SET @sql := IF(
  @has_col = 0,
  'ALTER TABLE indikator_sub_kegiatan ADD COLUMN indikator_kegiatan_id BIGINT UNSIGNED NULL AFTER id',
  'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 2) Add index if not exists
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

-- 3) Backfill indikator_kegiatan_id where possible
-- Uses the earliest indikator_kegiatan per kegiatan as fallback mapping.
UPDATE indikator_sub_kegiatan isk
JOIN sub_kegiatan sk ON sk.id = isk.sub_kegiatan_id
JOIN (
  SELECT kegiatan_id, MIN(id) AS indikator_kegiatan_id
  FROM indikator_kegiatan
  GROUP BY kegiatan_id
) ik_map ON ik_map.kegiatan_id = sk.kegiatan_id
SET isk.indikator_kegiatan_id = ik_map.indikator_kegiatan_id
WHERE isk.indikator_kegiatan_id IS NULL;

-- 4) Add foreign key if not exists
SET @has_fk := (
  SELECT COUNT(*)
  FROM information_schema.REFERENTIAL_CONSTRAINTS
  WHERE CONSTRAINT_SCHEMA = DATABASE()
    AND TABLE_NAME = 'indikator_sub_kegiatan'
    AND CONSTRAINT_NAME = 'fk_indikator_sub_kegiatan_indikator_kegiatan'
);

SET @sql := IF(
  @has_fk = 0,
  'ALTER TABLE indikator_sub_kegiatan ADD CONSTRAINT fk_indikator_sub_kegiatan_indikator_kegiatan FOREIGN KEY (indikator_kegiatan_id) REFERENCES indikator_kegiatan(id) ON UPDATE CASCADE ON DELETE SET NULL',
  'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;