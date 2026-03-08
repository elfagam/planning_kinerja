-- Add optional link from rencana_kerja to indikator_sub_kegiatan
-- Equivalent request:
-- ALTER TABLE rencana_kerja ADD indikator_sub_kegiatan_id BIGINT UNSIGNED NULL;

SET @col_exists := (
	SELECT COUNT(1)
	FROM INFORMATION_SCHEMA.COLUMNS
	WHERE TABLE_SCHEMA = DATABASE()
	  AND TABLE_NAME = 'rencana_kerja'
	  AND COLUMN_NAME = 'indikator_sub_kegiatan_id'
);

SET @ddl := IF(
	@col_exists = 0,
	'ALTER TABLE rencana_kerja ADD COLUMN indikator_sub_kegiatan_id BIGINT UNSIGNED NULL',
	'SELECT 1'
);

PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
