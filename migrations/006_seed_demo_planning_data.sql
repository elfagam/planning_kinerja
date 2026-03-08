-- Seed demo planning hierarchy and performance data
-- Safe to run multiple times (idempotent pattern).

-- ------------------------------
-- Strategic hierarchy
-- ------------------------------
INSERT INTO visis (code, name, description, start_year, end_year, is_active)
SELECT 'VISI-2025-2029',
       'Menjadi RSUD unggul dengan layanan berkualitas dan berfokus pada keselamatan pasien',
       'Visi jangka menengah RSUD untuk periode 2025-2029',
       2025,
       2029,
       1
WHERE NOT EXISTS (SELECT 1 FROM visis WHERE code = 'VISI-2025-2029');

INSERT INTO misis (visi_id, code, name, description)
SELECT v.id,
       'MISI-01',
       'Meningkatkan mutu pelayanan klinis dan keselamatan pasien',
       'Fokus pada clinical governance, patient safety, dan continuous improvement'
FROM visis v
WHERE v.code = 'VISI-2025-2029'
  AND NOT EXISTS (SELECT 1 FROM misis WHERE code = 'MISI-01');

INSERT INTO tujuans (misi_id, code, name, description)
SELECT m.id,
       'TUJ-01',
       'Meningkatkan kepuasan dan outcome layanan pasien',
       'Outcome layanan menjadi target utama dalam perencanaan kinerja'
FROM misis m
WHERE m.code = 'MISI-01'
  AND NOT EXISTS (SELECT 1 FROM tujuans WHERE code = 'TUJ-01');

INSERT INTO indikator_tujuans (tujuan_id, code, name, formula, unit, baseline)
SELECT t.id,
       'ITJ-01',
       'Indeks Kepuasan Pasien',
       '(Jumlah skor survei / jumlah responden)',
       'Skor',
       78.50
FROM tujuans t
WHERE t.code = 'TUJ-01'
  AND NOT EXISTS (SELECT 1 FROM indikator_tujuans WHERE code = 'ITJ-01');

INSERT INTO sasarans (tujuan_id, code, name, description)
SELECT t.id,
       'SAS-01',
       'Peningkatan kepatuhan standar pelayanan minimal',
       'Memastikan unit layanan mematuhi standar mutu RSUD'
FROM tujuans t
WHERE t.code = 'TUJ-01'
  AND NOT EXISTS (SELECT 1 FROM sasarans WHERE code = 'SAS-01');

INSERT INTO indikator_sasarans (sasaran_id, code, name, formula, unit, baseline)
SELECT s.id,
       'ISS-01',
       'Persentase Kepatuhan SPM',
       '(Jumlah indikator SPM terpenuhi / total indikator SPM) * 100',
       'Persen',
       82.00
FROM sasarans s
WHERE s.code = 'SAS-01'
  AND NOT EXISTS (SELECT 1 FROM indikator_sasarans WHERE code = 'ISS-01');

INSERT INTO programs (sasaran_id, unit_id, code, name, description)
SELECT s.id,
       u.id,
       'PRG-01',
       'Program Peningkatan Mutu Pelayanan',
       'Program untuk perbaikan mutu klinis dan non-klinis'
FROM sasarans s
JOIN units u ON u.code = 'UNIT-MED'
WHERE s.code = 'SAS-01'
  AND NOT EXISTS (SELECT 1 FROM programs WHERE code = 'PRG-01');

INSERT INTO indikator_programs (program_id, code, name, formula, unit, baseline)
SELECT p.id,
       'IPR-01',
       'Persentase Implementasi Rencana Mutu Program',
       '(Jumlah rencana mutu terlaksana / total rencana mutu) * 100',
       'Persen',
       75.00
FROM programs p
WHERE p.code = 'PRG-01'
  AND NOT EXISTS (SELECT 1 FROM indikator_programs WHERE code = 'IPR-01');

INSERT INTO kegiatans (program_id, code, name, description)
SELECT p.id,
       'KEG-01',
       'Pelatihan Keselamatan Pasien',
       'Pelatihan berkala untuk tenaga kesehatan terkait patient safety'
FROM programs p
WHERE p.code = 'PRG-01'
  AND NOT EXISTS (SELECT 1 FROM kegiatans WHERE code = 'KEG-01');

INSERT INTO indikator_kegiatans (kegiatan_id, code, name, formula, unit, baseline)
SELECT k.id,
       'IKG-01',
       'Cakupan Tenaga Kesehatan Terlatih',
       '(Jumlah tenaga terlatih / total tenaga sasaran) * 100',
       'Persen',
       70.00
FROM kegiatans k
WHERE k.code = 'KEG-01'
  AND NOT EXISTS (SELECT 1 FROM indikator_kegiatans WHERE code = 'IKG-01');

INSERT INTO sub_kegiatans (kegiatan_id, code, name, description)
SELECT k.id,
       'SUB-01',
       'Workshop Insiden Keselamatan Pasien',
       'Simulasi pelaporan dan analisis insiden keselamatan pasien'
FROM kegiatans k
WHERE k.code = 'KEG-01'
  AND NOT EXISTS (SELECT 1 FROM sub_kegiatans WHERE code = 'SUB-01');

INSERT INTO indikator_sub_kegiatans (sub_kegiatan_id, code, name, formula, unit, baseline)
SELECT sk.id,
       'ISK-01',
       'Jumlah Workshop Terselenggara',
       'Jumlah workshop terlaksana dalam periode berjalan',
       'Kegiatan',
       2.00
FROM sub_kegiatans sk
WHERE sk.code = 'SUB-01'
  AND NOT EXISTS (SELECT 1 FROM indikator_sub_kegiatans WHERE code = 'ISK-01');

-- ------------------------------
-- Renja and budget item
-- ------------------------------
INSERT INTO renjas (period_id, unit_id, code, name, status, notes)
SELECT p.id,
       u.id,
       'RENJA-UNIT-MED-2026Q1',
       'Renja Unit Medis Triwulan I 2026',
       'DRAFT',
       'Renja awal untuk demonstrasi sistem e-Planning RSUD'
FROM periods p
JOIN units u ON u.code = 'UNIT-MED'
WHERE p.year = 2026 AND p.quarter = 1 AND p.month IS NULL
  AND NOT EXISTS (SELECT 1 FROM renjas WHERE code = 'RENJA-UNIT-MED-2026Q1');

INSERT INTO renja_items (renja_id, program_id, kegiatan_id, sub_kegiatan_id, budget)
SELECT r.id,
       pr.id,
       k.id,
       sk.id,
       150000000.00
FROM renjas r
JOIN programs pr ON pr.code = 'PRG-01'
JOIN kegiatans k ON k.code = 'KEG-01'
JOIN sub_kegiatans sk ON sk.code = 'SUB-01'
WHERE r.code = 'RENJA-UNIT-MED-2026Q1'
  AND NOT EXISTS (
      SELECT 1 FROM renja_items ri
      WHERE ri.renja_id = r.id
        AND ri.program_id = pr.id
        AND ri.kegiatan_id = k.id
        AND ri.sub_kegiatan_id = sk.id
  );

-- ------------------------------
-- KPI and target realization
-- ------------------------------
INSERT INTO indikator_kinerjas (code, name, reference_type, reference_id, formula, unit)
SELECT 'IKIN-01',
       'Capaian Kepuasan Pasien Triwulanan',
       'INDIKATOR_TUJUAN',
       it.id,
       '(Nilai survei kepuasan pasien per triwulan)',
       'Skor'
FROM indikator_tujuans it
WHERE it.code = 'ITJ-01'
  AND NOT EXISTS (SELECT 1 FROM indikator_kinerjas WHERE code = 'IKIN-01');

INSERT INTO target_realisasis (
    indikator_kinerja_id,
    period_id,
    target_value,
    realisasi_value,
    deviation_value,
    capaian_percent,
    status,
    verification_status,
    verified_by,
    verified_at,
    notes
)
SELECT ik.id,
       p.id,
       82.00,
       80.50,
       -1.50,
       98.17,
       'WARNING',
       'DRAFT',
       NULL,
       NULL,
       'Realisasi awal triwulan I, perlu tindak lanjut pada unit prioritas'
FROM indikator_kinerjas ik
JOIN periods p ON p.year = 2026 AND p.quarter = 1 AND p.month IS NULL
WHERE ik.code = 'IKIN-01'
  AND NOT EXISTS (
      SELECT 1 FROM target_realisasis tr
      WHERE tr.indikator_kinerja_id = ik.id AND tr.period_id = p.id
  );
