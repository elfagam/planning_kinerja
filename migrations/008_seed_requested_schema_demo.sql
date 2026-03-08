-- Seed demo data for requested Indonesian schema tables (migration 007).
-- Safe to run multiple times (idempotent pattern).

-- ------------------------------
-- Unit pengusul & pelaksana
-- ------------------------------
INSERT INTO unit_pengusul (kode, nama, keterangan, aktif)
SELECT 'UP-MED', 'Unit Pengusul Medis', 'Unit pengusul program pelayanan medis', 1
WHERE NOT EXISTS (SELECT 1 FROM unit_pengusul WHERE kode = 'UP-MED');

INSERT INTO unit_pengusul (kode, nama, keterangan, aktif)
SELECT 'UP-NUR', 'Unit Pengusul Keperawatan', 'Unit pengusul program keperawatan', 1
WHERE NOT EXISTS (SELECT 1 FROM unit_pengusul WHERE kode = 'UP-NUR');

INSERT INTO unit_pelaksana (kode, nama, keterangan, aktif)
SELECT 'UL-RAWAT', 'Unit Pelaksana Rawat Inap', 'Pelaksana kegiatan rawat inap', 1
WHERE NOT EXISTS (SELECT 1 FROM unit_pelaksana WHERE kode = 'UL-RAWAT');

INSERT INTO unit_pelaksana (kode, nama, keterangan, aktif)
SELECT 'UL-IGD', 'Unit Pelaksana IGD', 'Pelaksana kegiatan IGD', 1
WHERE NOT EXISTS (SELECT 1 FROM unit_pelaksana WHERE kode = 'UL-IGD');

-- ------------------------------
-- Hierarki perencanaan
-- ------------------------------
INSERT INTO visi (kode, nama, deskripsi, tahun_mulai, tahun_selesai, aktif)
SELECT 'VISI-RSUD-01',
       'RSUD unggul dalam mutu layanan dan keselamatan pasien',
       'Visi strategis RSUD untuk periode perencanaan menengah',
       2025,
       2029,
       1
WHERE NOT EXISTS (SELECT 1 FROM visi WHERE kode = 'VISI-RSUD-01');

INSERT INTO misi (visi_id, kode, nama, deskripsi)
SELECT v.id,
       'MISI-RSUD-01',
       'Meningkatkan mutu pelayanan klinis',
       'Peningkatan mutu klinis berbasis indikator kinerja'
FROM visi v
WHERE v.kode = 'VISI-RSUD-01'
  AND NOT EXISTS (SELECT 1 FROM misi WHERE kode = 'MISI-RSUD-01');

INSERT INTO tujuan (misi_id, kode, nama, deskripsi)
SELECT m.id,
       'TUJUAN-RSUD-01',
       'Meningkatkan kepuasan pasien',
       'Kepuasan pasien sebagai tolok ukur kualitas layanan'
FROM misi m
WHERE m.kode = 'MISI-RSUD-01'
  AND NOT EXISTS (SELECT 1 FROM tujuan WHERE kode = 'TUJUAN-RSUD-01');

INSERT INTO indikator_tujuan (tujuan_id, kode, nama, formula, satuan, baseline)
SELECT t.id,
       'IT-01',
       'Indeks Kepuasan Pasien',
       '(Total skor survei / total responden)',
       'Skor',
       80.00
FROM tujuan t
WHERE t.kode = 'TUJUAN-RSUD-01'
  AND NOT EXISTS (SELECT 1 FROM indikator_tujuan WHERE kode = 'IT-01');

INSERT INTO sasaran (tujuan_id, kode, nama, deskripsi)
SELECT t.id,
       'SAS-RSUD-01',
       'Peningkatan kepatuhan standar pelayanan',
       'Meningkatkan kepatuhan terhadap standar layanan minimal'
FROM tujuan t
WHERE t.kode = 'TUJUAN-RSUD-01'
  AND NOT EXISTS (SELECT 1 FROM sasaran WHERE kode = 'SAS-RSUD-01');

INSERT INTO indikator_sasaran (sasaran_id, kode, nama, formula, satuan, baseline)
SELECT s.id,
       'IS-01',
       'Persentase Kepatuhan SPM',
       '(Jumlah indikator terpenuhi / total indikator) * 100',
       'Persen',
       82.00
FROM sasaran s
WHERE s.kode = 'SAS-RSUD-01'
  AND NOT EXISTS (SELECT 1 FROM indikator_sasaran WHERE kode = 'IS-01');

INSERT INTO program (sasaran_id, unit_pengusul_id, kode, nama, deskripsi)
SELECT s.id,
       up.id,
       'PRG-RSUD-01',
       'Program Peningkatan Mutu Pelayanan',
       'Program utama peningkatan mutu pelayanan rumah sakit'
FROM sasaran s
JOIN unit_pengusul up ON up.kode = 'UP-MED'
WHERE s.kode = 'SAS-RSUD-01'
  AND NOT EXISTS (SELECT 1 FROM program WHERE kode = 'PRG-RSUD-01');

INSERT INTO indikator_program (program_id, kode, nama, formula, satuan, baseline)
SELECT p.id,
       'IP-01',
       'Capaian Implementasi Program',
       '(Kegiatan selesai / kegiatan direncanakan) * 100',
       'Persen',
       75.00
FROM program p
WHERE p.kode = 'PRG-RSUD-01'
  AND NOT EXISTS (SELECT 1 FROM indikator_program WHERE kode = 'IP-01');

INSERT INTO kegiatan (program_id, unit_pelaksana_id, kode, nama, deskripsi)
SELECT p.id,
       ul.id,
       'KEG-RSUD-01',
       'Pelatihan Keselamatan Pasien',
       'Pelatihan rutin tentang patient safety'
FROM program p
JOIN unit_pelaksana ul ON ul.kode = 'UL-RAWAT'
WHERE p.kode = 'PRG-RSUD-01'
  AND NOT EXISTS (SELECT 1 FROM kegiatan WHERE kode = 'KEG-RSUD-01');

INSERT INTO indikator_kegiatan (kegiatan_id, kode, nama, formula, satuan, baseline)
SELECT k.id,
       'IK-01',
       'Cakupan SDM Terlatih',
       '(SDM terlatih / total SDM sasaran) * 100',
       'Persen',
       70.00
FROM kegiatan k
WHERE k.kode = 'KEG-RSUD-01'
  AND NOT EXISTS (SELECT 1 FROM indikator_kegiatan WHERE kode = 'IK-01');

INSERT INTO sub_kegiatan (kegiatan_id, kode, nama, deskripsi)
SELECT k.id,
       'SUB-RSUD-01',
       'Workshop Analisis Insiden',
       'Workshop analisis akar masalah insiden keselamatan pasien'
FROM kegiatan k
WHERE k.kode = 'KEG-RSUD-01'
  AND NOT EXISTS (SELECT 1 FROM sub_kegiatan WHERE kode = 'SUB-RSUD-01');

INSERT INTO indikator_sub_kegiatan (sub_kegiatan_id, kode, nama, formula, satuan, baseline)
SELECT sk.id,
       'ISK-01',
       'Jumlah Workshop Terlaksana',
       'Jumlah workshop yang terlaksana dalam tahun berjalan',
       'Kegiatan',
       2.00
FROM sub_kegiatan sk
WHERE sk.kode = 'SUB-RSUD-01'
  AND NOT EXISTS (SELECT 1 FROM indikator_sub_kegiatan WHERE kode = 'ISK-01');

-- ------------------------------
-- Rencana kerja + indikator + realisasi
-- ------------------------------
INSERT INTO rencana_kerja (
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
SELECT 'RK-2026-UPMED-T1',
       'Rencana Kerja Unit Pengusul Medis Triwulan I 2026',
       2026,
       1,
       up.id,
       'DIAJUKAN',
       'Rencana kerja demonstrasi untuk modul Indonesian schema',
       u_creator.id,
       u_approver.id,
       CURRENT_TIMESTAMP
FROM unit_pengusul up
JOIN users u_creator ON u_creator.email = 'planner.med@rsudcontoh.go.id'
LEFT JOIN users u_approver ON u_approver.email = 'approver@rsudcontoh.go.id'
WHERE up.kode = 'UP-MED'
  AND NOT EXISTS (SELECT 1 FROM rencana_kerja WHERE kode = 'RK-2026-UPMED-T1');

INSERT INTO indikator_rencana_kerja (
  rencana_kerja_id,
  indikator_sub_kegiatan_id,
  kode,
  nama,
  satuan,
  target_tahunan,
  anggaran_tahunan
)
SELECT rk.id,
       isk.id,
       'IRK-01',
       'Target Pelaksanaan Workshop Keselamatan Pasien',
       'Kegiatan',
       4.00,
       180000000.00
FROM rencana_kerja rk
JOIN indikator_sub_kegiatan isk ON isk.kode = 'ISK-01'
WHERE rk.kode = 'RK-2026-UPMED-T1'
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
       'Realisasi triwulan I sesuai progres awal kegiatan',
       u.id
FROM indikator_rencana_kerja irk
JOIN users u ON u.email = 'planner.med@rsudcontoh.go.id'
WHERE irk.kode = 'IRK-01'
  AND NOT EXISTS (
      SELECT 1 FROM realisasi_rencana_kerja rrk
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
       u.id,
       CURRENT_TIMESTAMP,
       'Target dan realisasi triwulan I tercapai'
FROM indikator_rencana_kerja irk
JOIN users u ON u.email = 'verifier@rsudcontoh.go.id'
WHERE irk.kode = 'IRK-01'
  AND NOT EXISTS (
      SELECT 1 FROM target_dan_realisasi tdr
      WHERE tdr.indikator_rencana_kerja_id = irk.id
        AND tdr.tahun = 2026
        AND tdr.triwulan = 1
  );
