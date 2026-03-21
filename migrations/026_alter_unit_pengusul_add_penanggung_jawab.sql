-- Add penanggungjawab fields to unit_pengusul after nama

SET @col1_exists := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'unit_pengusul'
    AND COLUMN_NAME = 'nama_penanggungjawab'
);

SET @ddl1 := IF(
  @col1_exists = 0,
  'ALTER TABLE unit_pengusul ADD COLUMN nama_penanggungjawab VARCHAR(150) NULL AFTER nama',
  'SELECT "column nama_penanggungjawab already exists on unit_pengusul"'
);

PREPARE stmt1 FROM @ddl1;
EXECUTE stmt1;
DEALLOCATE PREPARE stmt1;

-- ---

SET @col2_exists := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'unit_pengusul'
    AND COLUMN_NAME = 'nip_penanggungjawab'
);

SET @ddl2 := IF(
  @col2_exists = 0,
  'ALTER TABLE unit_pengusul ADD COLUMN nip_penanggungjawab VARCHAR(30) NULL AFTER nama_penanggungjawab',
  'SELECT "column nip_penanggungjawab already exists on unit_pengusul"'
);

PREPARE stmt2 FROM @ddl2;
EXECUTE stmt2;
DEALLOCATE PREPARE stmt2;

-- ---

SET @col3_exists := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'unit_pengusul'
    AND COLUMN_NAME = 'jabatan_penanggungjawab'
);

SET @ddl3 := IF(
  @col3_exists = 0,
  'ALTER TABLE unit_pengusul ADD COLUMN jabatan_penanggungjawab VARCHAR(100) NULL AFTER nip_penanggungjawab',
  'SELECT "column jabatan_penanggungjawab already exists on unit_pengusul"'
);

PREPARE stmt3 FROM @ddl3;
EXECUTE stmt3;
DEALLOCATE PREPARE stmt3;
