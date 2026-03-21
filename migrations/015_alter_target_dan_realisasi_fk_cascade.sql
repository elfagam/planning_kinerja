-- Migration to update foreign key constraint for target_dan_realisasi
-- Adds ON DELETE CASCADE to fk_id_target_realisasi_indikator_rk

ALTER TABLE target_dan_realisasi
DROP FOREIGN KEY fk_id_target_realisasi_indikator_rk;

ALTER TABLE target_dan_realisasi
ADD CONSTRAINT fk_id_target_realisasi_indikator_rk
FOREIGN KEY (indikator_rencana_kerja_id)
REFERENCES indikator_rencana_kerja(id)
ON DELETE CASCADE;
