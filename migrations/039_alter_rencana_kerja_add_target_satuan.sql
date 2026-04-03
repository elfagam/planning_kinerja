-- Migration: Add target and satuan to rencana_kerja
-- Target Table: rencana_kerja

ALTER TABLE rencana_kerja
ADD COLUMN target DECIMAL(18,2) NOT NULL DEFAULT 0 AFTER triwulan,
ADD COLUMN satuan VARCHAR(60) NULL AFTER target;
