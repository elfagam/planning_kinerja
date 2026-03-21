-- Migration: Add harga_satuan to indikator_rencana_kerja
-- Tambahkan kolom harga_satuan setelah target_tahunan

ALTER TABLE indikator_rencana_kerja
  ADD COLUMN harga_satuan DECIMAL(18,2) AFTER target_tahunan;
