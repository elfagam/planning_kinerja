-- Create informasi table for rotating topic text and route navigation.
-- This migration is idempotent via CREATE TABLE IF NOT EXISTS.

CREATE TABLE IF NOT EXISTS informasi (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    informasi TEXT NOT NULL,
    tahun INT NOT NULL,
    pilihan_route_halaman_tujuan VARCHAR(120) NOT NULL,
    tanggal_pembuatan DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    tanggal_ubah DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_informasi_tahun (tahun),
    INDEX idx_informasi_tanggal_pembuatan (tanggal_pembuatan)
);
