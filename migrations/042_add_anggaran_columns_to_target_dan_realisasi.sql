-- Migration to add financial (anggaran) columns to target_dan_realisasi table
-- These columns are required for the KPI summary and performance tracking in the dashboard

ALTER TABLE target_dan_realisasi
ADD COLUMN target_anggaran DECIMAL(18,2) NOT NULL DEFAULT 0 AFTER realisasi_nilai,
ADD COLUMN realisasi_anggaran DECIMAL(18,2) NOT NULL DEFAULT 0 AFTER target_anggaran,
ADD COLUMN capaian_anggaran DECIMAL(8,2) GENERATED ALWAYS AS (
    CASE
        WHEN target_anggaran = 0 THEN 0
        ELSE (realisasi_anggaran / target_anggaran) * 100
    END
) STORED AFTER realisasi_anggaran;
