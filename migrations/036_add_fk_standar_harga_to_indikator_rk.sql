-- Migration: Add tb_standar_harga_id to indikator_rencana_kerja
ALTER TABLE indikator_rencana_kerja
ADD COLUMN tb_standar_harga_id BIGINT UNSIGNED NULL AFTER rencana_kerja_id,
ADD CONSTRAINT fk_indikator_rk_tb_standar_harga 
    FOREIGN KEY (tb_standar_harga_id) 
    REFERENCES tb_standar_harga(id) 
    ON DELETE SET NULL 
    ON UPDATE CASCADE;
