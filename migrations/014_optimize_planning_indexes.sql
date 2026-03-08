-- Optimize indexes for common planning and reporting queries.
-- Idempotent migration using INFORMATION_SCHEMA checks.

-- program(unit_pengusul_id, sasaran_id)
SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'program' AND INDEX_NAME = 'idx_program_unit_sasaran'
);
SET @ddl := IF(@idx_exists = 0, 'ALTER TABLE program ADD INDEX idx_program_unit_sasaran (unit_pengusul_id, sasaran_id)', 'SELECT 1');
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- kegiatan(unit_pelaksana_id, program_id)
SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'kegiatan' AND INDEX_NAME = 'idx_kegiatan_unit_program'
);
SET @ddl := IF(@idx_exists = 0, 'ALTER TABLE kegiatan ADD INDEX idx_kegiatan_unit_program (unit_pelaksana_id, program_id)', 'SELECT 1');
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- rencana_kerja(unit_pengusul_id, tahun, triwulan)
SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'rencana_kerja' AND INDEX_NAME = 'idx_rencana_kerja_unit_periode'
);
SET @ddl := IF(@idx_exists = 0, 'ALTER TABLE rencana_kerja ADD INDEX idx_rencana_kerja_unit_periode (unit_pengusul_id, tahun, triwulan)', 'SELECT 1');
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- rencana_kerja(tahun, triwulan, status)
SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'rencana_kerja' AND INDEX_NAME = 'idx_rencana_kerja_periode_status'
);
SET @ddl := IF(@idx_exists = 0, 'ALTER TABLE rencana_kerja ADD INDEX idx_rencana_kerja_periode_status (tahun, triwulan, status)', 'SELECT 1');
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- indikator_rencana_kerja(rencana_kerja_id, kode)
SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'indikator_rencana_kerja' AND INDEX_NAME = 'idx_indikator_rk_rk_kode'
);
SET @ddl := IF(@idx_exists = 0, 'ALTER TABLE indikator_rencana_kerja ADD INDEX idx_indikator_rk_rk_kode (rencana_kerja_id, kode)', 'SELECT 1');
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- realisasi_rencana_kerja(tahun, triwulan, bulan)
SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'realisasi_rencana_kerja' AND INDEX_NAME = 'idx_realisasi_rk_periode'
);
SET @ddl := IF(@idx_exists = 0, 'ALTER TABLE realisasi_rencana_kerja ADD INDEX idx_realisasi_rk_periode (tahun, triwulan, bulan)', 'SELECT 1');
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- realisasi_rencana_kerja(diinput_oleh, tahun, triwulan)
SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'realisasi_rencana_kerja' AND INDEX_NAME = 'idx_realisasi_rk_input_user'
);
SET @ddl := IF(@idx_exists = 0, 'ALTER TABLE realisasi_rencana_kerja ADD INDEX idx_realisasi_rk_input_user (diinput_oleh, tahun, triwulan)', 'SELECT 1');
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- target_dan_realisasi(tahun, triwulan, status)
SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'target_dan_realisasi' AND INDEX_NAME = 'idx_target_realisasi_periode_status'
);
SET @ddl := IF(@idx_exists = 0, 'ALTER TABLE target_dan_realisasi ADD INDEX idx_target_realisasi_periode_status (tahun, triwulan, status)', 'SELECT 1');
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- target_dan_realisasi(diverifikasi_oleh, tahun, triwulan)
SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'target_dan_realisasi' AND INDEX_NAME = 'idx_target_realisasi_verifikator_periode'
);
SET @ddl := IF(@idx_exists = 0, 'ALTER TABLE target_dan_realisasi ADD INDEX idx_target_realisasi_verifikator_periode (diverifikasi_oleh, tahun, triwulan)', 'SELECT 1');
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;
