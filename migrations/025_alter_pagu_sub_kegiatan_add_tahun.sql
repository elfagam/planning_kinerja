SET @column_exists := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'pagu_sub_kegiatan'
    AND COLUMN_NAME = 'tahun'
);

SET @ddl := IF(
  @column_exists = 0,
  'ALTER TABLE pagu_sub_kegiatan ADD COLUMN tahun YEAR NULL AFTER sub_kegiatan_id',
  'SELECT "column tahun already exists on pagu_sub_kegiatan"'
);

PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
