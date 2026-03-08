ALTER TABLE indikator_sub_kegiatan
  ADD COLUMN anggaran_tahun_sebelumnya DECIMAL(18,2) NOT NULL DEFAULT 0 AFTER baseline,
  ADD COLUMN anggaran_tahun_ini DECIMAL(18,2) NOT NULL DEFAULT 0 AFTER anggaran_tahun_sebelumnya;
