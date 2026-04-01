-- Migration: Add indexes to tb_standar_harga for better search performance
-- Target Table: tb_standar_harga

-- Index for id_rekening (exact or prefix match frequently used)
CREATE INDEX idx_standar_harga_rekening ON tb_standar_harga(id_rekening);

-- Index for jenis_standar (categorical filter)
CREATE INDEX idx_standar_harga_jenis ON tb_standar_harga(jenis_standar);

-- Prefix indexes for TEXT columns used in searches
-- Using 255 character prefix to keep index size manageable while covering most search criteria
CREATE INDEX idx_standar_harga_uraian ON tb_standar_harga(uraian_barang(255));
CREATE INDEX idx_standar_harga_spesifikasi ON tb_standar_harga(spesifikasi(255));
