-- Seed default informasi records for text switcher.
-- Safe to run multiple times by guarding with NOT EXISTS checks.

INSERT INTO informasi (
    informasi,
    tahun,
    pilihan_route_halaman_tujuan,
    tanggal_pembuatan,
    tanggal_ubah
)
SELECT
    'Dashboard: Fokus pada indikator kinerja prioritas tahun berjalan.',
    YEAR(CURRENT_DATE),
    '/dashboard',
    NOW() - INTERVAL 2 MINUTE,
    NOW() - INTERVAL 2 MINUTE
WHERE NOT EXISTS (
    SELECT 1 FROM informasi
    WHERE informasi = 'Dashboard: Fokus pada indikator kinerja prioritas tahun berjalan.'
      AND pilihan_route_halaman_tujuan = '/dashboard'
);

INSERT INTO informasi (
    informasi,
    tahun,
    pilihan_route_halaman_tujuan,
    tanggal_pembuatan,
    tanggal_ubah
)
SELECT
    'Rencana Kerja: Pastikan keterkaitan sub kegiatan dan pagu sudah valid.',
    YEAR(CURRENT_DATE),
    '/rencana-kerja',
    NOW() - INTERVAL 1 MINUTE,
    NOW() - INTERVAL 1 MINUTE
WHERE NOT EXISTS (
    SELECT 1 FROM informasi
    WHERE informasi = 'Rencana Kerja: Pastikan keterkaitan sub kegiatan dan pagu sudah valid.'
      AND pilihan_route_halaman_tujuan = '/rencana-kerja'
);

INSERT INTO informasi (
    informasi,
    tahun,
    pilihan_route_halaman_tujuan,
    tanggal_pembuatan,
    tanggal_ubah
)
SELECT
    'Target dan Evaluasi: Review capaian triwulan terakhir sebelum finalisasi.',
    YEAR(CURRENT_DATE),
    '/target-evaluasi',
    NOW(),
    NOW()
WHERE NOT EXISTS (
    SELECT 1 FROM informasi
    WHERE informasi = 'Target dan Evaluasi: Review capaian triwulan terakhir sebelum finalisasi.'
      AND pilihan_route_halaman_tujuan = '/target-evaluasi'
);
