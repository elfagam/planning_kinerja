-- Requested Indonesian table set for RSUD performance planning.
-- Compatible with existing migrations 001-006.
-- Note: table `users` already exists from 001_init_schema.sql and is reused.

CREATE TABLE IF NOT EXISTS unit_pengusul (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    kode VARCHAR(30) NOT NULL,
    nama VARCHAR(150) NOT NULL,
    keterangan TEXT NULL,
    aktif TINYINT(1) NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uq_unit_pengusul_kode (kode)
);

CREATE TABLE IF NOT EXISTS unit_pelaksana (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    kode VARCHAR(30) NOT NULL,
    nama VARCHAR(150) NOT NULL,
    keterangan TEXT NULL,
    aktif TINYINT(1) NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uq_unit_pelaksana_kode (kode)
);

CREATE TABLE IF NOT EXISTS visi (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    kode VARCHAR(30) NOT NULL,
    nama VARCHAR(255) NOT NULL,
    deskripsi TEXT NULL,
    tahun_mulai SMALLINT NOT NULL,
    tahun_selesai SMALLINT NOT NULL,
    aktif TINYINT(1) NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uq_visi_kode (kode)
);

CREATE TABLE IF NOT EXISTS misi (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    visi_id BIGINT UNSIGNED NOT NULL,
    kode VARCHAR(30) NOT NULL,
    nama VARCHAR(255) NOT NULL,
    deskripsi TEXT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT fk_id_misi_visi FOREIGN KEY (visi_id) REFERENCES visi(id),
    UNIQUE KEY uq_misi_kode (kode),
    INDEX idx_misi_visi (visi_id)
);

CREATE TABLE IF NOT EXISTS tujuan (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    misi_id BIGINT UNSIGNED NOT NULL,
    kode VARCHAR(30) NOT NULL,
    nama VARCHAR(255) NOT NULL,
    deskripsi TEXT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT fk_id_tujuan_misi FOREIGN KEY (misi_id) REFERENCES misi(id),
    UNIQUE KEY uq_tujuan_kode (kode),
    INDEX idx_tujuan_misi (misi_id)
);

CREATE TABLE IF NOT EXISTS indikator_tujuan (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    tujuan_id BIGINT UNSIGNED NOT NULL,
    kode VARCHAR(40) NOT NULL,
    nama VARCHAR(255) NOT NULL,
    formula TEXT NULL,
    satuan VARCHAR(60) NULL,
    baseline DECIMAL(18,2) NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT fk_id_indikator_tujuan_tujuan FOREIGN KEY (tujuan_id) REFERENCES tujuan(id),
    UNIQUE KEY uq_indikator_tujuan_kode (kode),
    INDEX idx_indikator_tujuan_tujuan (tujuan_id)
);

CREATE TABLE IF NOT EXISTS sasaran (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    tujuan_id BIGINT UNSIGNED NOT NULL,
    kode VARCHAR(30) NOT NULL,
    nama VARCHAR(255) NOT NULL,
    deskripsi TEXT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT fk_id_sasaran_tujuan FOREIGN KEY (tujuan_id) REFERENCES tujuan(id),
    UNIQUE KEY uq_sasaran_kode (kode),
    INDEX idx_sasaran_tujuan (tujuan_id)
);

CREATE TABLE IF NOT EXISTS indikator_sasaran (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    sasaran_id BIGINT UNSIGNED NOT NULL,
    kode VARCHAR(40) NOT NULL,
    nama VARCHAR(255) NOT NULL,
    formula TEXT NULL,
    satuan VARCHAR(60) NULL,
    baseline DECIMAL(18,2) NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT fk_id_indikator_sasaran_sasaran FOREIGN KEY (sasaran_id) REFERENCES sasaran(id),
    UNIQUE KEY uq_indikator_sasaran_kode (kode),
    INDEX idx_indikator_sasaran_sasaran (sasaran_id)
);

CREATE TABLE IF NOT EXISTS program (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    sasaran_id BIGINT UNSIGNED NOT NULL,
    unit_pengusul_id BIGINT UNSIGNED NOT NULL,
    kode VARCHAR(40) NOT NULL,
    nama VARCHAR(255) NOT NULL,
    deskripsi TEXT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT fk_id_program_sasaran FOREIGN KEY (sasaran_id) REFERENCES sasaran(id),
    CONSTRAINT fk_id_program_unit_pengusul FOREIGN KEY (unit_pengusul_id) REFERENCES unit_pengusul(id),
    UNIQUE KEY uq_program_kode (kode),
    INDEX idx_program_sasaran (sasaran_id),
    INDEX idx_program_unit_pengusul (unit_pengusul_id)
);

CREATE TABLE IF NOT EXISTS indikator_program (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    program_id BIGINT UNSIGNED NOT NULL,
    kode VARCHAR(40) NOT NULL,
    nama VARCHAR(255) NOT NULL,
    formula TEXT NULL,
    satuan VARCHAR(60) NULL,
    baseline DECIMAL(18,2) NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT fk_id_indikator_program_program FOREIGN KEY (program_id) REFERENCES program(id),
    UNIQUE KEY uq_indikator_program_kode (kode),
    INDEX idx_indikator_program_program (program_id)
);

CREATE TABLE IF NOT EXISTS kegiatan (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    program_id BIGINT UNSIGNED NOT NULL,
    unit_pelaksana_id BIGINT UNSIGNED NOT NULL,
    kode VARCHAR(40) NOT NULL,
    nama VARCHAR(255) NOT NULL,
    deskripsi TEXT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT fk_id_kegiatan_program FOREIGN KEY (program_id) REFERENCES program(id),
    CONSTRAINT fk_id_kegiatan_unit_pelaksana FOREIGN KEY (unit_pelaksana_id) REFERENCES unit_pelaksana(id),
    UNIQUE KEY uq_kegiatan_kode (kode),
    INDEX idx_kegiatan_program (program_id),
    INDEX idx_kegiatan_unit_pelaksana (unit_pelaksana_id)
);

CREATE TABLE IF NOT EXISTS indikator_kegiatan (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    kegiatan_id BIGINT UNSIGNED NOT NULL,
    kode VARCHAR(40) NOT NULL,
    nama VARCHAR(255) NOT NULL,
    formula TEXT NULL,
    satuan VARCHAR(60) NULL,
    baseline DECIMAL(18,2) NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT fk_id_indikator_kegiatan_kegiatan FOREIGN KEY (kegiatan_id) REFERENCES kegiatan(id),
    UNIQUE KEY uq_indikator_kegiatan_kode (kode),
    INDEX idx_indikator_kegiatan_kegiatan (kegiatan_id)
);

CREATE TABLE IF NOT EXISTS sub_kegiatan (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    kegiatan_id BIGINT UNSIGNED NOT NULL,
    kode VARCHAR(40) NOT NULL,
    nama VARCHAR(255) NOT NULL,
    deskripsi TEXT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT fk_id_sub_kegiatan_kegiatan FOREIGN KEY (kegiatan_id) REFERENCES kegiatan(id),
    UNIQUE KEY uq_sub_kegiatan_kode (kode),
    INDEX idx_sub_kegiatan_kegiatan (kegiatan_id)
);

CREATE TABLE IF NOT EXISTS indikator_sub_kegiatan (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    sub_kegiatan_id BIGINT UNSIGNED NOT NULL,
    kode VARCHAR(40) NOT NULL,
    nama VARCHAR(255) NOT NULL,
    formula TEXT NULL,
    satuan VARCHAR(60) NULL,
    baseline DECIMAL(18,2) NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT fk_id_indikator_sub_kegiatan_sub_kegiatan FOREIGN KEY (sub_kegiatan_id) REFERENCES sub_kegiatan(id),
    UNIQUE KEY uq_indikator_sub_kegiatan_kode (kode),
    INDEX idx_indikator_sub_kegiatan_sub_kegiatan (sub_kegiatan_id)
);

CREATE TABLE IF NOT EXISTS rencana_kerja (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    indikator_sub_kegiatan_id BIGINT UNSIGNED NULL,
    kode VARCHAR(50) NOT NULL,
    nama VARCHAR(255) NOT NULL,
    tahun SMALLINT NOT NULL,
    triwulan TINYINT NULL,
    unit_pengusul_id BIGINT UNSIGNED NOT NULL,
    status ENUM('DRAFT', 'DIAJUKAN', 'DISETUJUI', 'DITOLAK') NOT NULL DEFAULT 'DRAFT',
    catatan TEXT NULL,
    dibuat_oleh BIGINT UNSIGNED NOT NULL,
    disetujui_oleh BIGINT UNSIGNED NULL,
    tanggal_persetujuan DATETIME NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT fk_id_rencana_kerja_unit_pengusul FOREIGN KEY (unit_pengusul_id) REFERENCES unit_pengusul(id),
    CONSTRAINT fk_id_rencana_kerja_dibuat_oleh FOREIGN KEY (dibuat_oleh) REFERENCES users(id),
    CONSTRAINT fk_id_rencana_kerja_disetujui_oleh FOREIGN KEY (disetujui_oleh) REFERENCES users(id),
    CONSTRAINT fk_id_rencana_kerja_indikator_sub_kegiatan FOREIGN KEY (indikator_sub_kegiatan_id) REFERENCES indikator_sub_kegiatan(id),
    UNIQUE KEY uq_rencana_kerja_kode (kode),
    INDEX idx_rencana_kerja_tahun (tahun),
    INDEX idx_rencana_kerja_status (status)
);

CREATE TABLE IF NOT EXISTS indikator_rencana_kerja (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    rencana_kerja_id BIGINT UNSIGNED NOT NULL,
    indikator_sub_kegiatan_id BIGINT UNSIGNED NULL,
    kode VARCHAR(50) NOT NULL,
    nama VARCHAR(255) NOT NULL,
    satuan VARCHAR(60) NULL,
    target_tahunan DECIMAL(18,2) NOT NULL DEFAULT 0,
    anggaran_tahunan DECIMAL(18,2) NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT fk_id_indikator_rk_rk FOREIGN KEY (rencana_kerja_id) REFERENCES rencana_kerja(id),
    CONSTRAINT fk_id_indikator_rk_indikator_sub_kegiatan FOREIGN KEY (indikator_sub_kegiatan_id) REFERENCES indikator_sub_kegiatan(id),
    UNIQUE KEY uq_indikator_rk_kode (kode),
    INDEX idx_indikator_rk_rk (rencana_kerja_id)
);

CREATE TABLE IF NOT EXISTS realisasi_rencana_kerja (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    indikator_rencana_kerja_id BIGINT UNSIGNED NOT NULL,
    tahun SMALLINT NOT NULL,
    bulan TINYINT NULL,
    triwulan TINYINT NULL,
    nilai_realisasi DECIMAL(18,2) NOT NULL DEFAULT 0,
    realisasi_anggaran DECIMAL(18,2) NOT NULL DEFAULT 0,
    keterangan TEXT NULL,
    diinput_oleh BIGINT UNSIGNED NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT fk_id_realisasi_rk_indikator_rk FOREIGN KEY (indikator_rencana_kerja_id) REFERENCES indikator_rencana_kerja(id),
    CONSTRAINT fk_id_realisasi_rk_diinput_oleh FOREIGN KEY (diinput_oleh) REFERENCES users(id),
    UNIQUE KEY uq_realisasi_rk_periode (indikator_rencana_kerja_id, tahun, bulan, triwulan),
    INDEX idx_realisasi_rk_tahun (tahun)
);

CREATE TABLE IF NOT EXISTS target_dan_realisasi (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    indikator_rencana_kerja_id BIGINT UNSIGNED NOT NULL,
    tahun SMALLINT NOT NULL,
    triwulan TINYINT NOT NULL,
    target_nilai DECIMAL(18,2) NOT NULL DEFAULT 0,
    realisasi_nilai DECIMAL(18,2) NOT NULL DEFAULT 0,
    capaian_persen DECIMAL(8,2) GENERATED ALWAYS AS (
        CASE
            WHEN target_nilai = 0 THEN 0
            ELSE (realisasi_nilai / target_nilai) * 100
        END
    ) STORED,
    status ENUM('ON_TRACK', 'WARNING', 'OFF_TRACK') NOT NULL DEFAULT 'ON_TRACK',
    diverifikasi_oleh BIGINT UNSIGNED NULL,
    tanggal_verifikasi DATETIME NULL,
    catatan TEXT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT fk_id_target_realisasi_indikator_rk FOREIGN KEY (indikator_rencana_kerja_id) REFERENCES indikator_rencana_kerja(id),
    CONSTRAINT fk_id_target_realisasi_verifikator FOREIGN KEY (diverifikasi_oleh) REFERENCES users(id),
    UNIQUE KEY uq_target_realisasi_periode (indikator_rencana_kerja_id, tahun, triwulan),
    INDEX idx_target_realisasi_status (status)
);
