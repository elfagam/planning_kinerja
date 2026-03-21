CREATE TABLE IF NOT EXISTS pagu_sub_kegiatan (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    sub_kegiatan_id BIGINT UNSIGNED NOT NULL,
    pagu_tahun_sebelumnya DECIMAL(18,2) NOT NULL DEFAULT 0,
    pagu_tahun_ini DECIMAL(18,2) NOT NULL DEFAULT 0,
    CONSTRAINT fk_pagu_sub_kegiatan_sub_kegiatan FOREIGN KEY (sub_kegiatan_id) REFERENCES sub_kegiatan(id),
    INDEX idx_pagu_sub_kegiatan_sub_kegiatan (sub_kegiatan_id)
);
