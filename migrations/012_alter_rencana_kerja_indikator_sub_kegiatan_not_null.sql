-- Enforce rencana_kerja.indikator_sub_kegiatan_id as NOT NULL.
-- Strategy:
-- 1) Backfill NULL values from indikator_rencana_kerja when possible.
-- 2) Alter to NOT NULL only when no NULL values remain.

-- Step 1: Backfill from child table mapping (if available)
UPDATE rencana_kerja rk
JOIN (
    SELECT
        rencana_kerja_id,
        MIN(indikator_sub_kegiatan_id) AS indikator_sub_kegiatan_id
    FROM indikator_rencana_kerja
    WHERE indikator_sub_kegiatan_id IS NOT NULL
    GROUP BY rencana_kerja_id
) src ON src.rencana_kerja_id = rk.id
SET rk.indikator_sub_kegiatan_id = src.indikator_sub_kegiatan_id
WHERE rk.indikator_sub_kegiatan_id IS NULL;

-- Step 2: Enforce NOT NULL only if it is now safe
SET @rk_col_exists := (
    SELECT COUNT(1)
    FROM INFORMATION_SCHEMA.COLUMNS
    WHERE TABLE_SCHEMA = DATABASE()
      AND TABLE_NAME = 'rencana_kerja'
      AND COLUMN_NAME = 'indikator_sub_kegiatan_id'
);

SET @rk_has_null := (
    SELECT COUNT(1)
    FROM rencana_kerja
    WHERE indikator_sub_kegiatan_id IS NULL
);

SET @ddl := IF(
    @rk_col_exists = 1 AND @rk_has_null = 0,
    'ALTER TABLE rencana_kerja MODIFY COLUMN indikator_sub_kegiatan_id BIGINT UNSIGNED NOT NULL',
    'SELECT 1'
);

PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
