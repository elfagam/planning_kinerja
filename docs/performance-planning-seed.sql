-- Minimal seed data for docs/performance-planning-schema.sql
-- Idempotent insert pattern for repeated execution.

USE `e-plan-ai`;

-- =========================
-- Master Data
-- =========================
INSERT INTO unit_pengusul (kode, nama, keterangan, aktif)
SELECT 'UP-MED', 'Unit Pengusul Medis', 'Unit pengusul layanan medis', 1
WHERE NOT EXISTS (SELECT 1 FROM unit_pengusul WHERE kode = 'UP-MED');

INSERT INTO unit_pelaksana (kode, nama, keterangan, aktif)
SELECT 'UL-RAWAT', 'Unit Pelaksana Rawat Inap', 'Unit pelaksana rawat inap', 1
WHERE NOT EXISTS (SELECT 1 FROM unit_pelaksana WHERE kode = 'UL-RAWAT');

INSERT INTO users (unit_pengusul_id, unit_pelaksana_id, nama_lengkap, email, password_hash, role, aktif)
SELECT up.id, ul.id, 'Admin Perencanaan', 'admin.planning@rsud.local', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'ADMIN', 1
FROM unit_pengusul up
JOIN unit_pelaksana ul ON ul.kode = 'UL-RAWAT'
WHERE up.kode = 'UP-MED'
  AND NOT EXISTS (SELECT 1 FROM users WHERE email = 'admin.planning@rsud.local');

INSERT INTO users (unit_pengusul_id, unit_pelaksana_id, nama_lengkap, email, password_hash, role, aktif)
SELECT up.id, ul.id, 'Verifier Kinerja', 'verifier@rsud.local', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'VERIFIKATOR', 1
FROM unit_pengusul up
JOIN unit_pelaksana ul ON ul.kode = 'UL-RAWAT'
WHERE up.kode = 'UP-MED'
  AND NOT EXISTS (SELECT 1 FROM users WHERE email = 'verifier@rsud.local');

-- =========================
-- Hierarchy Data
-- =========================
INSERT INTO visi (kode, nama, deskripsi, tahun_mulai, tahun_selesai, aktif)
SELECT 'VISI-01', 'RSUD unggul dalam mutu layanan', 'Visi periode 2025-2029', 2025, 2029, 1
WHERE NOT EXISTS (SELECT 1 FROM visi WHERE kode = 'VISI-01');

INSERT INTO misi (visi_id, kode, nama, deskripsi)
SELECT v.id, 'MISI-01', 'Meningkatkan mutu pelayanan klinis', 'Fokus mutu berbasis indikator'
FROM visi v
WHERE v.kode = 'VISI-01'
  AND NOT EXISTS (SELECT 1 FROM misi WHERE kode = 'MISI-01');

INSERT INTO tujuan (misi_id, kode, nama, deskripsi)
SELECT m.id, 'TUJ-01', 'Meningkatkan kepuasan pasien', 'Outcome layanan pasien'
FROM misi m
WHERE m.kode = 'MISI-01'
  AND NOT EXISTS (SELECT 1 FROM tujuan WHERE kode = 'TUJ-01');

INSERT INTO indikator_tujuan (tujuan_id, kode, nama, formula, satuan, baseline)
SELECT t.id, 'IT-01', 'Indeks Kepuasan Pasien', '(total skor/total responden)', 'Skor', 80.00
FROM tujuan t
WHERE t.kode = 'TUJ-01'
  AND NOT EXISTS (SELECT 1 FROM indikator_tujuan WHERE kode = 'IT-01');

INSERT INTO sasaran (tujuan_id, kode, nama, deskripsi)
SELECT t.id, 'SAS-01', 'Kepatuhan standar pelayanan', 'Pemenuhan standar layanan minimal'
FROM tujuan t
WHERE t.kode = 'TUJ-01'
  AND NOT EXISTS (SELECT 1 FROM sasaran WHERE kode = 'SAS-01');

INSERT INTO indikator_sasaran (sasaran_id, kode, nama, formula, satuan, baseline)
SELECT s.id, 'IS-01', 'Persentase Kepatuhan SPM', '(indikator terpenuhi/total indikator)*100', 'Persen', 82.00
FROM sasaran s
WHERE s.kode = 'SAS-01'
  AND NOT EXISTS (SELECT 1 FROM indikator_sasaran WHERE kode = 'IS-01');

INSERT INTO program (sasaran_id, unit_pengusul_id, kode, nama, deskripsi)
SELECT s.id, up.id, 'PRG-01', 'Program Peningkatan Mutu', 'Program mutu layanan RSUD'
FROM sasaran s
JOIN unit_pengusul up ON up.kode = 'UP-MED'
WHERE s.kode = 'SAS-01'
  AND NOT EXISTS (SELECT 1 FROM program WHERE kode = 'PRG-01');

INSERT INTO indikator_program (program_id, kode, nama, formula, satuan, baseline)
SELECT p.id, 'IP-01', 'Capaian Program Mutu', '(kegiatan selesai/kegiatan rencana)*100', 'Persen', 75.00
FROM program p
WHERE p.kode = 'PRG-01'
  AND NOT EXISTS (SELECT 1 FROM indikator_program WHERE kode = 'IP-01');

INSERT INTO kegiatan (program_id, unit_pelaksana_id, kode, nama, deskripsi)
SELECT p.id, ul.id, 'KEG-01', 'Pelatihan Keselamatan Pasien', 'Pelatihan rutin patient safety'
FROM program p
JOIN unit_pelaksana ul ON ul.kode = 'UL-RAWAT'
WHERE p.kode = 'PRG-01'
  AND NOT EXISTS (SELECT 1 FROM kegiatan WHERE kode = 'KEG-01');

INSERT INTO indikator_kegiatan (kegiatan_id, kode, nama, formula, satuan, baseline)
SELECT k.id, 'IK-01', 'Cakupan SDM Terlatih', '(sdm terlatih/total sasaran)*100', 'Persen', 70.00
FROM kegiatan k
WHERE k.kode = 'KEG-01'
  AND NOT EXISTS (SELECT 1 FROM indikator_kegiatan WHERE kode = 'IK-01');

INSERT INTO sub_kegiatan (kegiatan_id, kode, nama, deskripsi)
SELECT k.id, 'SUB-01', 'Workshop Analisis Insiden', 'Analisis akar masalah insiden'
FROM kegiatan k
WHERE k.kode = 'KEG-01'
  AND NOT EXISTS (SELECT 1 FROM sub_kegiatan WHERE kode = 'SUB-01');

INSERT INTO indikator_sub_kegiatan (sub_kegiatan_id, kode, nama, formula, satuan, baseline)
SELECT sk.id, 'ISK-01', 'Jumlah Workshop Terlaksana', 'Jumlah workshop per tahun', 'Kegiatan', 2.00
FROM sub_kegiatan sk
WHERE sk.kode = 'SUB-01'
  AND NOT EXISTS (SELECT 1 FROM indikator_sub_kegiatan WHERE kode = 'ISK-01');

-- =========================
-- Work Plan and Realization
-- =========================
INSERT INTO rencana_kerja (
  indikator_sub_kegiatan_id,
  kode,
  nama,
  tahun,
  triwulan,
  unit_pengusul_id,
  status,
  catatan,
  dibuat_oleh,
  disetujui_oleh,
  tanggal_persetujuan
)
SELECT isk.id,
       'RK-2026-T1',
       'Rencana Kerja Triwulan I 2026',
       2026,
       1,
       up.id,
       'DISETUJUI',
       'Data seed untuk pengujian awal',
       u_admin.id,
       u_admin.id,
       CURRENT_TIMESTAMP
FROM indikator_sub_kegiatan isk
JOIN unit_pengusul up ON up.kode = 'UP-MED'
JOIN users u_admin ON u_admin.email = 'admin.planning@rsud.local'
WHERE isk.kode = 'ISK-01'
  AND NOT EXISTS (SELECT 1 FROM rencana_kerja WHERE kode = 'RK-2026-T1');

INSERT INTO indikator_rencana_kerja (
  rencana_kerja_id,
  kode,
  nama,
  satuan,
  target_tahunan,
  anggaran_tahunan
)
SELECT rk.id,
       'IRK-01',
       'Target Workshop Keselamatan Pasien',
       'Kegiatan',
       4.00,
       180000000.00
FROM rencana_kerja rk
WHERE rk.kode = 'RK-2026-T1'
  AND NOT EXISTS (SELECT 1 FROM indikator_rencana_kerja WHERE kode = 'IRK-01');

INSERT INTO realisasi_rencana_kerja (
  indikator_rencana_kerja_id,
  tahun,
  bulan,
  triwulan,
  nilai_realisasi,
  realisasi_anggaran,
  keterangan,
  diinput_oleh
)
SELECT irk.id,
       2026,
       NULL,
       1,
       1.00,
       40000000.00,
       'Realisasi triwulan I',
       u_admin.id
FROM indikator_rencana_kerja irk
JOIN users u_admin ON u_admin.email = 'admin.planning@rsud.local'
WHERE irk.kode = 'IRK-01'
  AND NOT EXISTS (
    SELECT 1
    FROM realisasi_rencana_kerja rrk
    WHERE rrk.indikator_rencana_kerja_id = irk.id
      AND rrk.tahun = 2026
      AND rrk.triwulan = 1
      AND rrk.bulan IS NULL
  );

INSERT INTO target_dan_realisasi (
  indikator_rencana_kerja_id,
  tahun,
  triwulan,
  target_nilai,
  realisasi_nilai,
  status,
  diverifikasi_oleh,
  tanggal_verifikasi,
  catatan
)
SELECT irk.id,
       2026,
       1,
       1.00,
       1.00,
       'ON_TRACK',
       u_ver.id,
       CURRENT_TIMESTAMP,
       'Target dan realisasi sesuai pada triwulan I'
FROM indikator_rencana_kerja irk
JOIN users u_ver ON u_ver.email = 'verifier@rsud.local'
WHERE irk.kode = 'IRK-01'
  AND NOT EXISTS (
    SELECT 1
    FROM target_dan_realisasi tdr
    WHERE tdr.indikator_rencana_kerja_id = irk.id
      AND tdr.tahun = 2026
      AND tdr.triwulan = 1
  );
