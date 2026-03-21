-- Verification checklist for:
-- 023_add_indikator_kegiatan_fk_to_indikator_sub_kegiatan.sql
-- 024_enforce_indikator_sub_kegiatan_indikator_kegiatan_not_null.sql

-- A) Check unresolved rows before enforcing NOT NULL
SELECT COUNT(*) AS null_indikator_kegiatan_id
FROM indikator_sub_kegiatan
WHERE indikator_kegiatan_id IS NULL;

-- Expected before running 024: 0

-- B) Check orphan relation (should be 0)
SELECT COUNT(*) AS orphan_relations
FROM indikator_sub_kegiatan isk
LEFT JOIN indikator_kegiatan ik ON ik.id = isk.indikator_kegiatan_id
WHERE ik.id IS NULL;

-- C) Check column nullability status
SELECT COLUMN_NAME, IS_NULLABLE, COLUMN_TYPE
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA = DATABASE()
  AND TABLE_NAME = 'indikator_sub_kegiatan'
  AND COLUMN_NAME = 'indikator_kegiatan_id';

-- Expected after running 024: IS_NULLABLE = 'NO'

-- D) Check FK existence and delete/update rules
SELECT
  rc.CONSTRAINT_NAME,
  rc.UPDATE_RULE,
  rc.DELETE_RULE,
  kcu.TABLE_NAME,
  kcu.COLUMN_NAME,
  kcu.REFERENCED_TABLE_NAME,
  kcu.REFERENCED_COLUMN_NAME
FROM information_schema.REFERENTIAL_CONSTRAINTS rc
JOIN information_schema.KEY_COLUMN_USAGE kcu
  ON rc.CONSTRAINT_SCHEMA = kcu.CONSTRAINT_SCHEMA
 AND rc.CONSTRAINT_NAME = kcu.CONSTRAINT_NAME
WHERE rc.CONSTRAINT_SCHEMA = DATABASE()
  AND kcu.TABLE_NAME = 'indikator_sub_kegiatan'
  AND kcu.COLUMN_NAME = 'indikator_kegiatan_id';

-- Expected after 024:
-- CONSTRAINT_NAME = fk_indikator_sub_kegiatan_indikator_kegiatan
-- UPDATE_RULE = CASCADE
-- DELETE_RULE = RESTRICT

-- E) Check index existence
SELECT INDEX_NAME, SEQ_IN_INDEX, COLUMN_NAME
FROM information_schema.STATISTICS
WHERE TABLE_SCHEMA = DATABASE()
  AND TABLE_NAME = 'indikator_sub_kegiatan'
  AND INDEX_NAME = 'idx_indikator_sub_kegiatan_indikator_kegiatan';

-- F) Optional distribution sanity check (top mappings)
SELECT isk.indikator_kegiatan_id, COUNT(*) AS total_sub_indikator
FROM indikator_sub_kegiatan isk
GROUP BY isk.indikator_kegiatan_id
ORDER BY total_sub_indikator DESC, isk.indikator_kegiatan_id ASC
LIMIT 20;

-- G) Optional detail sample for manual spot-check
SELECT
  isk.id AS indikator_sub_kegiatan_id,
  isk.kode AS indikator_sub_kegiatan_kode,
  sk.id AS sub_kegiatan_id,
  sk.kegiatan_id,
  ik.id AS indikator_kegiatan_id,
  ik.kode AS indikator_kegiatan_kode,
  ik.nama AS indikator_kegiatan_nama
FROM indikator_sub_kegiatan isk
JOIN sub_kegiatan sk ON sk.id = isk.sub_kegiatan_id
JOIN indikator_kegiatan ik ON ik.id = isk.indikator_kegiatan_id
ORDER BY isk.id ASC
LIMIT 50;
