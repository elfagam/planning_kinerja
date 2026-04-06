-- Migration: 043_fix_standar_harga_id_auto_increment.sql
-- Goal: Add AUTO_INCREMENT to tb_standar_harga(id) and ensure FK compatibility

-- 1. Drop the foreign key constraint temporarily to allow column modification
ALTER TABLE indikator_rencana_kerja DROP FOREIGN KEY fk_indikator_rk_tb_standar_harga;

-- 2. Modify tb_standar_harga(id) to BIGINT UNSIGNED AUTO_INCREMENT
-- This also matches GORM's uint64 pattern for primary keys
ALTER TABLE tb_standar_harga MODIFY id BIGINT UNSIGNED AUTO_INCREMENT;

-- 3. Modify referencing column in indikator_rencana_kerja to BIGINT UNSIGNED
-- Important for type compatibility between referencing and referenced columns
ALTER TABLE indikator_rencana_kerja MODIFY tb_standar_harga_id BIGINT UNSIGNED;

-- 4. Re-add the foreign key constraint
ALTER TABLE indikator_rencana_kerja 
ADD CONSTRAINT fk_indikator_rk_tb_standar_harga 
FOREIGN KEY (tb_standar_harga_id) 
REFERENCES tb_standar_harga(id);
