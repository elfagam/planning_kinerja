# Arsitektur Sistem E-Planning RSUD

## 1. Arsitektur Makro

Sistem menggunakan pola **Modular Monolith** agar cepat diimplementasikan, mudah di-maintain, dan siap diekstrak menjadi microservices bila beban meningkat.

Komponen utama:

- `web` (Bootstrap UI): antarmuka admin/perencana/verifikator/pimpinan.
- `api` (Gin): REST API versioned (`/api/v1`).
- `application`: orkestrasi use case per domain.
- `domain`: business rules dan model inti.
- `infrastructure`: MySQL repository, logging, audit, export.

Alur data:

1. User mengakses halaman Bootstrap.
2. UI mengirim request ke Gin API.
3. API memanggil service/use case per module.
4. Use case berinteraksi dengan repository MySQL.
5. Hasil dikembalikan dalam JSON terstandar.

## 2. Layering Clean Architecture

Struktur per module:

- `domain`: entitas, value object, aturan bisnis.
- `usecase`: alur bisnis (create, validate, approve, report).
- `repository`: implementasi persistence MySQL.
- `delivery/http`: handler Gin + DTO + route.

Shared components:

- `internal/shared/middleware`: recovery, auth, audit middleware.
- `internal/shared/response`: standar response sukses/gagal.
- `internal/shared/database`: connection pool dan transaction helper.

## 3. Penjelasan Module

1. **Visi**

- Pernyataan arah strategis RSUD jangka panjang.
- Menjadi akar struktur perencanaan.

2. **Misi**

- Penjabaran tindakan strategis dari Visi.
- Satu Visi dapat memiliki banyak Misi.

3. **Tujuan**

- Outcome strategis dari Misi.
- Digunakan sebagai poros penyusunan Sasaran.

4. **Indikator Tujuan**

- KPI untuk mengukur ketercapaian Tujuan.
- Menyimpan formula, satuan, baseline, target.

5. **Sasaran**

- Target antara yang lebih operasional dari Tujuan.
- Menjadi referensi pembentukan Program.

6. **Indikator Sasaran**

- KPI spesifik untuk Sasaran.
- Digunakan dalam evaluasi periodik.

7. **Program**

- Paket kerja lintas aktivitas untuk mencapai Sasaran.
- Umumnya terikat unit kerja pemilik program.

8. **Indikator Program**

- KPI outcome/output level program.

9. **Kegiatan**

- Breakdown Program ke aktivitas terencana.

10. **Indikator Kegiatan**

- KPI level aktivitas (volume, kualitas, waktu, biaya).

11. **Sub Kegiatan**

- Detail eksekusi Kegiatan paling granular.
- Sering menjadi unit realisasi anggaran/operasional.

12. **Indikator Sub Kegiatan**

- KPI paling detail untuk monitoring lapangan.

13. **Renja**

- Rencana kerja tahunan yang mengikat Program/Kegiatan/Sub Kegiatan dengan tahun anggaran.

14. **Indikator Kinerja**

- Konsolidasi indikator lintas level untuk pelaporan organisasi.

15. **Target dan Realisasi**

- Mencatat target awal/revisi serta realisasi periodik.
- Menghasilkan capaian persen, deviasi, dan status kinerja.

## 4. Keamanan dan Governance

- RBAC: `admin`, `perencana`, `verifikator`, `pimpinan`, `auditor`.
- Audit trail untuk perubahan target/realisasi.
- Soft delete (`deleted_at`) dan active flag.
- Periode lock untuk mencegah perubahan pasca pengesahan.
- Transaction boundary di use case kritikal (approval, revisi target).

## 5. Roadmap Implementasi

1. Finalisasi DDL dan migrasi awal.
2. Implement auth + RBAC + user management.
3. Implement master planning hierarchy (Visi sampai Sub Kegiatan).
4. Implement Renja tahunan dan approval flow.
5. Implement monitoring target/realisasi + dashboard.
6. Tambahkan reporting PDF/Excel.
