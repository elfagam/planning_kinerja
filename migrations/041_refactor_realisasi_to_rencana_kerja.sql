-- Migration to refactor realisasi_rencana_kerja and target_dan_realisasi to Rencana Kerja level

-- 1. Table realisasi_rencana_kerja
ALTER TABLE realisasi_rencana_kerja 
DROP FOREIGN KEY fk_id_realisasi_rk_indikator_rk;

ALTER TABLE realisasi_rencana_kerja
DROP INDEX uq_realisasi_rk_periode;

ALTER TABLE realisasi_rencana_kerja
ADD COLUMN rencana_kerja_id BIGINT UNSIGNED NOT NULL AFTER id;

ALTER TABLE realisasi_rencana_kerja
ADD CONSTRAINT fk_realisasi_rk_rk FOREIGN KEY (rencana_kerja_id) REFERENCES rencana_kerja(id) ON DELETE CASCADE;

ALTER TABLE realisasi_rencana_kerja
ADD UNIQUE KEY uq_realisasi_rk_periode (rencana_kerja_id, tahun, bulan, triwulan);

ALTER TABLE realisasi_rencana_kerja
DROP COLUMN indikator_rencana_kerja_id;


-- 2. Table target_dan_realisasi
ALTER TABLE target_dan_realisasi
DROP FOREIGN KEY fk_id_target_realisasi_indikator_rk;

ALTER TABLE target_dan_realisasi
DROP INDEX uq_target_realisasi_periode;

ALTER TABLE target_dan_realisasi
ADD COLUMN rencana_kerja_id BIGINT UNSIGNED NOT NULL AFTER id;

ALTER TABLE target_dan_realisasi
ADD CONSTRAINT fk_target_realisasi_rk FOREIGN KEY (rencana_kerja_id) REFERENCES rencana_kerja(id) ON DELETE CASCADE;

ALTER TABLE target_dan_realisasi
ADD UNIQUE KEY uq_target_realisasi_periode (rencana_kerja_id, tahun, triwulan);

ALTER TABLE target_dan_realisasi
DROP COLUMN indikator_rencana_kerja_id;
