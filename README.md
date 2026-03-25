# e-plan-ai (E-Planning RSUD)

[![Go CI](https://github.com/elfagam/planning_kinerja/actions/workflows/go-ci.yml/badge.svg)](https://github.com/elfagam/planning_kinerja/actions/workflows/go-ci.yml)

Blueprint awal aplikasi e-Planning RSUD berbasis Golang + Gin + MySQL + Bootstrap.

## Quick Start

```bash
cp .env.example .env
go run ./cmd/api
```

Lalu buka `http://192.168.20.1:7002/ui`.

## Kredensial Demo (Development)

Setelah menjalankan migrasi terbaru, gunakan akun berikut:

- Username/Email: `superadmin@rsudcontoh.go.id`
- Password: `Admin123!`

Akun demo lain yang tersedia:

- `planner.med@rsudcontoh.go.id`
- `reviewer.nur@rsudcontoh.go.id`
- `approver@rsudcontoh.go.id`
- `verifier@rsudcontoh.go.id`

Semua akun demo di atas menggunakan password yang sama: `Admin123!`.

## Struktur Folder

```text
cmd/api                     # entrypoint aplikasi HTTP (Gin)
cmd/migrate                 # entrypoint CLI migrasi
configs                     # template konfigurasi aplikasi
docs                        # arsitektur, endpoint API, struktur proyek
internal/app                # composition root / app wiring
internal/bootstrap          # inisialisasi router dan modul
internal/config             # konfigurasi env
internal/modules            # business modules
internal/shared             # middleware, response, db helper
migrations                  # SQL migration
pkg                         # reusable package umum
scripts                     # script utilitas developer
tests                       # integration test
web/templates               # halaman Bootstrap
web/assets                  # static assets
```

## Menjalankan Aplikasi

## Fitur Export CSV Indikator Kinerja

Ekspor laporan Indikator Kinerja ke format CSV dengan layout terstruktur (top info, tabel, tanda tangan) melalui endpoint berikut:

- **Endpoint:**

  `/api/v1/renja/export/indikator-csv?rencana_kerja_id={ID}&unit_pengusul_id={ID}`

- **Metode:** `GET` (wajib login/JWT)

- **Query Parameter:**
  - `rencana_kerja_id` (wajib)
  - `unit_pengusul_id` (wajib)

- **Contoh cURL:**

```bash
curl -X GET "http://localhost:8080/api/v1/renja/export/indikator-csv?rencana_kerja_id=1&unit_pengusul_id=1" \
  -H "Authorization: Bearer <TOKEN-ANDA>" -o indikator_kinerja.csv
```

- **Layout CSV:**
  - Top Info: Program, Kegiatan, Sub Kegiatan, Unit Pengusul, Tahun
  - Tabel: Kode Rencana, Nama Rencana, ID Rekening, ID Standar Harga, Kode Indikator, **Nama Indikator**, Satuan, Harga Satuan, Target, Anggaran
  - Bawah: Tanda tangan penanggung jawab

- **Akses dari UI:**
  - Klik tombol "Export CSV" pada halaman Indikator Kinerja, pilih filter yang sesuai, file akan terunduh otomatis.

1. Install dependency:

```bash
go get github.com/gin-gonic/gin
```

2. Jalankan API:

```bash
go run ./cmd/api
```

3. Health check:

```bash
curl http://localhost:8080/health
```

4. Buka halaman modul CRUD terpisah:

```bash
open http://localhost:8080/ui
```

## Dokumen Penting

- Arsitektur: `docs/architecture.md`
- Reference Blueprint: `docs/reference-blueprint.md`
- Microservice Blueprint: `docs/microservice-architecture.md`
- Services Scaffold: `services/README.md`
- Contracts: `platform/contracts/openapi` and `platform/contracts/events`
- Endpoint: `docs/api-endpoints.md`
- Struktur Proyek: `docs/project-structure.md`
- DDL awal: `migrations/001_init_schema.sql`

## Contoh Endpoint CRUD

- `GET /api/v1/visi`
- `POST /api/v1/visi`
- `GET /api/v1/visi/:id`
- `PUT /api/v1/visi/:id`
- `DELETE /api/v1/visi/:id`

Pola endpoint yang sama tersedia untuk semua modul sampai `target-realisasi`.

## Contoh Endpoint Client (JWT Context)

Catatan penting:

- Endpoint mutasi Client mengambil actor dari JWT claim (`auth.user_id`, `auth.role`, `auth.full_name`).
- Payload request tidak perlu `actor_id` atau `actor_role`.

1. List client:

```bash
curl "http://localhost:8080/api/v1/clients?q=unit&status=DRAFT&page=1&limit=10" \
  -H "Authorization: Bearer change-me-in-production"
```

2. Create client:

```bash
curl -X POST http://localhost:8080/api/v1/clients \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer change-me-in-production" \
  -d '{
    "kode": "CL-2026-001",
    "nama": "Client Unit Pelayanan A",
    "unit_pengusul_id": 1
  }'
```

```bash
curl http://192.168.20.1:7002/health
```

curl -X PUT http://localhost:8080/api/v1/clients/1 \
 -H "Content-Type: application/json" \
 -H "Authorization: Bearer change-me-in-production" \
 -d '{
"nama": "Client Unit Pelayanan A (Revisi)"
}'

````

4. Submit client:

```bash
curl -X POST http://localhost:8080/api/v1/clients/1/submit \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer change-me-in-production" \
  -d '{"note":"siap diajukan"}'
````

5. Reject client:

```bash
curl -X POST http://localhost:8080/api/v1/clients/1/reject \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer change-me-in-production" \
  -d '{"reason":"dokumen belum lengkap"}'
```

6. Approve client:

```bash
curl -X POST http://localhost:8080/api/v1/clients/1/approve \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer change-me-in-production" \
  -d '{"note":"disetujui"}'
```

7. Delete client:

```bash
curl -X DELETE http://localhost:8080/api/v1/clients/1 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer change-me-in-production" \
  -d '{}'
```

## Contoh Update Target dan Realisasi

1. Create target dan realisasi:

```bash
curl -X POST http://localhost:8080/api/v1/performance/target-realisasi \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer change-me-in-production" \
  -d '{
    "indikator_rencana_kerja_id": 1,
    "tahun": 2026,
    "triwulan": 1,
    "target_nilai": 120,
    "realisasi_nilai": 80,
    "status": "ON_TRACK",
    "catatan": "Input baseline triwulan 1"
  }'
```

2. Update target dan realisasi per ID:

```bash
curl -X PUT http://localhost:8080/api/v1/performance/target-realisasi/1 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer change-me-in-production" \
  -d '{
    "target_nilai": 120,
    "realisasi_nilai": 100,
    "status": "ON_TRACK",
    "catatan": "Realisasi triwulan berjalan",
    "tanggal_verifikasi": "2026-03-08T10:30:00Z"
  }'
```

Catatan: `tanggal_verifikasi` menggunakan format RFC3339 (contoh: `2026-03-08T10:30:00Z`).

List `target-realisasi` mendukung query parameter:

- `q` untuk pencarian (`id`, `indikator_rencana_kerja_id`, `tahun`, `triwulan`, `status`)
- `tahun` untuk filter tahun
- `page` untuk nomor halaman
- `limit` untuk jumlah data per halaman (maks 100)

Contoh:

```bash
curl "http://localhost:8080/api/v1/performance/target-realisasi?q=on_track&tahun=2026&page=1&limit=10" \
  -H "Authorization: Bearer change-me-in-production"
```

## Contoh Hitung Persentase Capaian Kinerja

1. Hitung persentase capaian via query parameter (GET):

```bash
curl "http://localhost:8080/api/v1/performance/calculate-achievement?target_nilai=120&realisasi_nilai=102" \
  -H "Authorization: Bearer change-me-in-production"
```

2. Hitung persentase capaian via payload JSON (POST):

```bash
curl -X POST http://localhost:8080/api/v1/performance/calculate-achievement \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer change-me-in-production" \
  -d '{
    "target_nilai": 120,
    "realisasi_nilai": 102
  }'
```

Contoh hasil `capaian_persen`: `85.00`.

## Contoh Dashboard Summary

1. Ambil ringkasan total data dashboard:

```bash
curl http://localhost:8080/api/v1/performance/dashboard-summary \
  -H "Authorization: Bearer change-me-in-production"
```

Response berisi:

- `total_program`
- `total_kegiatan`
- `total_sub_kegiatan`
- `total_rencana_kerja`
- `total_anggaran`
- `total_realisasi_anggaran`
- `persentase_realisasi_anggaran`

## Contoh Performance Statistics

1. Statistik performa semua data:

```bash
curl http://localhost:8080/api/v1/performance/statistics \
  -H "Authorization: Bearer change-me-in-production"
```

2. Statistik performa dengan filter periode:

```bash
curl "http://localhost:8080/api/v1/performance/statistics?tahun=2026&triwulan=1" \
  -H "Authorization: Bearer change-me-in-production"
```

Response berisi:

- `total_data`
- `total_indikator`
- `total_status_on_track`
- `total_status_warning`
- `total_status_off_track`
- `total_target_nilai`
- `total_realisasi_nilai`
- `rata_rata_capaian_persen`
- `persentase_realisasi_target`

## Contoh Data Chart Target vs Realisasi

1. Data chart semua periode:

```bash
curl http://localhost:8080/api/v1/performance/chart-target-vs-realisasi \
  -H "Authorization: Bearer change-me-in-production"
```

2. Data chart per tahun:

```bash
curl "http://localhost:8080/api/v1/performance/chart-target-vs-realisasi?tahun=2026" \
  -H "Authorization: Bearer change-me-in-production"
```

Response berisi:

- `categories` (contoh: `2026-T1`, `2026-T2`)
- `series.target`
- `series.realisasi`

## Contoh Yearly Performance Summary

1. Ringkasan performa per tahun (semua tahun):

```bash
curl http://localhost:8080/api/v1/performance/yearly-summary \
  -H "Authorization: Bearer change-me-in-production"
```

2. Ringkasan performa per tahun (rentang tahun):

```bash
curl "http://localhost:8080/api/v1/performance/yearly-summary?tahun_start=2024&tahun_end=2026" \
  -H "Authorization: Bearer change-me-in-production"
```

Response per item berisi:

- `tahun`
- `total_data`
- `total_indikator`
- `total_target_nilai`
- `total_realisasi_nilai`
- `rata_rata_capaian_persen`
- `persentase_realisasi_target`
- `total_status_on_track`
- `total_status_warning`
- `total_status_off_track`

## Contoh Program Performance Ranking

1. Ranking performa program (default top 10):

```bash
curl http://localhost:8080/api/v1/performance/program-ranking \
  -H "Authorization: Bearer change-me-in-production"
```

2. Ranking performa program per periode (tahun/triwulan) dan limit:

```bash
curl "http://localhost:8080/api/v1/performance/program-ranking?tahun=2026&triwulan=1&limit=5" \
  -H "Authorization: Bearer change-me-in-production"
```

Response item berisi:

- `rank`
- `program_id`
- `program_kode`
- `program_nama`
- `total_indikator`
- `total_target_nilai`
- `total_realisasi_nilai`
- `rata_rata_capaian_persen`
- `persentase_realisasi_target`

## Contoh Endpoint Unit Pengusul

1. List Unit Pengusul:

```bash
curl "http://localhost:8080/api/v1/unit-pengusul?q=umum" \
  -H "Authorization: Bearer change-me-in-production"
```

2. Get detail Unit Pengusul:

```bash
curl http://localhost:8080/api/v1/unit-pengusul/1 \
  -H "Authorization: Bearer change-me-in-production"
```

3. Create Unit Pengusul:

```bash
curl -X POST http://localhost:8080/api/v1/unit-pengusul \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer change-me-in-production" \
  -d '{"kode":"UP-001","nama":"Unit Pengusul Umum","keterangan":"Layanan umum","aktif":true}'
```

4. Update Unit Pengusul:

```bash
curl -X PUT http://localhost:8080/api/v1/unit-pengusul/1 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer change-me-in-production" \
  -d '{"kode":"UP-001","nama":"Unit Pengusul Umum Revisi","keterangan":"Revisi nama unit","aktif":true}'
```

5. Delete Unit Pengusul:

```bash
curl -X DELETE http://localhost:8080/api/v1/unit-pengusul/1 \
  -H "Authorization: Bearer change-me-in-production"
```

Contoh response sukses list:

````json
{
```bash
open http://192.168.20.1:7002/ui
````

      {
      ```bash
      curl "http://192.168.20.1:7002/api/v1/clients?q=unit&status=DRAFT&page=1&limit=10" \
        -H "Authorization: Bearer change-me-in-production"
        "Keterangan": "Layanan umum",
        "Aktif": true,
        "CreatedAt": "2026-03-08T10:00:00Z",
        "UpdatedAt": "2026-03-08T10:00:00Z"
      }
      ```bash
      curl http://192.168.20.1:7002/api/v1/performance/dashboard-summary \
        -H "Authorization: Bearer change-me-in-production"

}

````bash
curl http://192.168.20.1:7002/api/v1/performance/statistics \
  -H "Authorization: Bearer change-me-in-production"

```json
{
```bash
curl http://192.168.20.1:7002/api/v1/performance/chart-target-vs-realisasi \
  -H "Authorization: Bearer change-me-in-production"
````

Status code error yang mungkin:

- `400` payload atau parameter tidak valid
- `404` data `unit_pengusul` tidak ditemukan
- `503` koneksi database tidak tersedia
- `500` error internal server

## Contoh Endpoint Unit Pelaksana

1. List Unit Pelaksana:

```bash
curl "http://localhost:8080/api/v1/unit-pelaksana?q=bedah" \
  -H "Authorization: Bearer change-me-in-production"
```

2. Get detail Unit Pelaksana:

```bash
curl http://localhost:8080/api/v1/unit-pelaksana/1 \
  -H "Authorization: Bearer change-me-in-production"
```

3. Create Unit Pelaksana:

```bash
curl -X POST http://localhost:8080/api/v1/unit-pelaksana \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer change-me-in-production" \
  -d '{"kode":"UPL-001","nama":"Unit Pelaksana Bedah","keterangan":"Pelaksana layanan bedah","aktif":true}'
```

4. Update Unit Pelaksana:

```bash
curl -X PUT http://localhost:8080/api/v1/unit-pelaksana/1 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer change-me-in-production" \
  -d '{"kode":"UPL-001","nama":"Unit Pelaksana Bedah Revisi","keterangan":"Perubahan nama unit","aktif":true}'
```

5. Delete Unit Pelaksana:

```bash
curl -X DELETE http://localhost:8080/api/v1/unit-pelaksana/1 \
  -H "Authorization: Bearer change-me-in-production"
```

Contoh response sukses list:

````json
{
```bash
curl http://192.168.20.1:7002/api/v1/performance/yearly-summary \
  -H "Authorization: Bearer change-me-in-production"
      {
      ```bash
      curl http://192.168.20.1:7002/api/v1/performance/program-ranking \
        -H "Authorization: Bearer change-me-in-production"
        "Keterangan": "Pelaksana layanan bedah",
        "Aktif": true,
        "CreatedAt": "2026-03-08T10:00:00Z",
        "UpdatedAt": "2026-03-08T10:00:00Z"
      }
      ```bash
      curl "http://192.168.20.1:7002/api/v1/unit-pengusul?q=umum" \
        -H "Authorization: Bearer change-me-in-production"
}
```bash
curl http://192.168.20.1:7002/api/v1/unit-pengusul/1 \
  -H "Authorization: Bearer change-me-in-production"

```json
{
```bash
curl "http://192.168.20.1:7002/api/v1/unit-pelaksana?q=bedah" \
  -H "Authorization: Bearer change-me-in-production"
````

Status code error yang mungkin:

- `400` payload atau parameter tidak valid
- `404` data `unit_pelaksana` tidak ditemukan
- `503` koneksi database tidak tersedia
- `500` error internal server

## Endpoint GORM Planning Lainnya

CRUD Gin+GORM juga tersedia untuk resource berikut:

- `misi`
- `tujuan`
- `indikator_tujuan`
- `sasaran`
- `indikator_sasaran`
- `program`
- `indikator_program`
- `kegiatan`
- `indikator_kegiatan`
- `sub_kegiatan`
- `indikator_sub_kegiatan`
- `rencana_kerja`
- `indikator_rencana_kerja`
- `realisasi_rencana_kerja`

Pola endpoint untuk masing-masing resource:

- `GET /api/v1/{resource}`
- `GET /api/v1/{resource}/:id`
- `POST /api/v1/{resource}`
- `PUT /api/v1/{resource}/:id`
- `DELETE /api/v1/{resource}/:id`

Contoh cepat (`misi`):

```bash
curl "http://localhost:8080/api/v1/misi?q=layanan" \
  -H "Authorization: Bearer change-me-in-production"

curl -X POST http://localhost:8080/api/v1/misi \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer change-me-in-production" \
  -d '{"visi_id":1,"kode":"MS-001","nama":"Peningkatan Layanan"}'
```

Catatan validasi payload:

- Field wajib tiap resource berbeda dan divalidasi di server (contoh `misi`: `visi_id`, `kode`, `nama`).
- Jika `AUTH_ENABLED=false`, header `Authorization` bisa dihapus.
- Jika koneksi DB gagal, endpoint mengembalikan `503`.

### Payload Minimum Per Resource

Gunakan field minimum berikut untuk request `POST` dan `PUT`:

| Resource                  | Field wajib minimum                                                                               |
| ------------------------- | ------------------------------------------------------------------------------------------------- |
| `misi`                    | `visi_id`, `kode`, `nama`                                                                         |
| `tujuan`                  | `misi_id`, `kode`, `nama`                                                                         |
| `indikator_tujuan`        | `tujuan_id`, `kode`, `nama`                                                                       |
| `sasaran`                 | `tujuan_id`, `kode`, `nama`                                                                       |
| `indikator_sasaran`       | `sasaran_id`, `kode`, `nama`                                                                      |
| `program`                 | `sasaran_id`, `unit_pengusul_id`, `kode`, `nama`                                                  |
| `indikator_program`       | `program_id`, `kode`, `nama`                                                                      |
| `kegiatan`                | `program_id`, `unit_pelaksana_id`, `kode`, `nama`                                                 |
| `indikator_kegiatan`      | `kegiatan_id`, `kode`, `nama`                                                                     |
| `sub_kegiatan`            | `kegiatan_id`, `kode`, `nama`                                                                     |
| `indikator_sub_kegiatan`  | `sub_kegiatan_id`, `kode`, `nama`                                                                 |
| `rencana_kerja`           | `indikator_sub_kegiatan_id`, `kode`, `nama`, `tahun`, `unit_pengusul_id`, `status`, `dibuat_oleh` |
| `indikator_rencana_kerja` | `rencana_kerja_id`, `kode`, `nama`                                                                |
| `realisasi_rencana_kerja` | `indikator_rencana_kerja_id`, `tahun`, `nilai_realisasi`, `realisasi_anggaran`, `diinput_oleh`    |

Contoh payload minimum `rencana_kerja`:

````json
{
```bash
curl http://192.168.20.1:7002/api/v1/unit-pelaksana/1 \
  -H "Authorization: Bearer change-me-in-production"
  "tahun": 2026,
  "unit_pengusul_id": 1,
  "status": "DRAFT",
  "dibuat_oleh": 101
}
```bash
curl "http://192.168.20.1:7002/api/v1/misi?q=layanan" \
  -H "Authorization: Bearer change-me-in-production"

1. Cek overview modul Renja:

```bash
curl http://localhost:8080/api/v1/renja/overview \
	-H "Authorization: Bearer change-me-in-production"
````

2. Submit Renja:

```bash
curl -X POST http://localhost:8080/api/v1/renja/1/submit \
	-H "Content-Type: application/json" \
	-H "Authorization: Bearer change-me-in-production" \
	-d '{"actor_id":101}'
```

3. Approve Renja:

```bash
curl -X POST http://localhost:8080/api/v1/renja/1/approve \
	-H "Content-Type: application/json" \
	-H "Authorization: Bearer change-me-in-production" \
	-d '{"actor_id":201}'
```

4. Reject Renja:

```bash
curl -X POST http://localhost:8080/api/v1/renja/1/reject \
	-H "Content-Type: application/json" \
	-H "Authorization: Bearer change-me-in-production" \
	-d '{"actor_id":201,"reason":"target belum realistis"}'
```

Catatan:

- Jika `AUTH_ENABLED=false`, header `Authorization` bisa dihapus.
- Jika koneksi DB belum tersedia, endpoint aksi Renja akan mengembalikan status `503`.

## Contoh Response Endpoint Renja

1. Response sukses `GET /api/v1/renja/overview`:

````json
{
```bash
curl -X POST http://192.168.20.1:7002/api/v1/misi \
  -H "Content-Type: application/json" \
    "scope": "Perencanaan kerja tahunan RSUD",
    "status": "OK",
    "storage": "mysql"
  }
  ```bash
  curl http://192.168.20.1:7002/api/v1/renja/overview \
    -H "Authorization: Bearer change-me-in-production"
2. Response sukses `POST /api/v1/renja/:id/submit`:

```json
{
```bash
curl -X GET "http://192.168.20.1:7002/api/v1/renja/export/indikator-csv?rencana_kerja_id=1&unit_pengusul_id=1" \
  -H "Authorization: Bearer <TOKEN-ANDA>" -o indikator_kinerja.csv
    "action": "submit"
  }
}
````

3. Response sukses `POST /api/v1/renja/:id/approve`:

```json
{
  "success": true,
  "data": {
    "id": 1,
    "action": "approve"
  }
}
```

4. Response sukses `POST /api/v1/renja/:id/reject`:

```json
{
  "success": true,
  "data": {
    "id": 1,
    "action": "reject"
  }
}
```

5. Contoh response gagal umum:

```json
{
  "success": false,
  "error": "invalid id"
}
```

Status code error yang mungkin:

- `400` payload/transition tidak valid (mis. `actor_id` kosong, reject tanpa reason, atau status tidak bisa ditransisikan)
- `404` data Renja tidak ditemukan
- `503` service Renja belum tersedia karena koneksi database gagal
- `500` error internal saat proses workflow

## Quick Test Script Renja

Gunakan skrip berikut untuk uji cepat endpoint Renja secara berurutan (overview -> submit -> approve -> reject):

```bash
#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"
TOKEN="${TOKEN:-change-me-in-production}"
RENJA_ID="${RENJA_ID:-1}"
RENJA_REJECT_ID="${RENJA_REJECT_ID:-2}"

AUTH_HEADER=( -H "Authorization: Bearer ${TOKEN}" )

echo "[1] Overview"
curl -sS "${BASE_URL}/api/v1/renja/overview" "${AUTH_HEADER[@]}" | jq . || true

echo "[2] Submit RENJA_ID=${RENJA_ID}"
curl -sS -X POST "${BASE_URL}/api/v1/renja/${RENJA_ID}/submit" \
  -H "Content-Type: application/json" \
  "${AUTH_HEADER[@]}" \
  -d '{"actor_id":101}' | jq . || true

echo "[3] Approve RENJA_ID=${RENJA_ID}"
curl -sS -X POST "${BASE_URL}/api/v1/renja/${RENJA_ID}/approve" \
  -H "Content-Type: application/json" \
  "${AUTH_HEADER[@]}" \
  -d '{"actor_id":201}' | jq . || true

echo "[4] Reject RENJA_REJECT_ID=${RENJA_REJECT_ID}"
curl -sS -X POST "${BASE_URL}/api/v1/renja/${RENJA_REJECT_ID}/reject" \
  -H "Content-Type: application/json" \
  "${AUTH_HEADER[@]}" \
  -d '{"actor_id":201,"reason":"target belum realistis"}' | jq . || true
```

Catatan penggunaan skrip:

- Secara default skrip memakai header `Authorization`.
- Jika `AUTH_ENABLED=false`, hapus `"${AUTH_HEADER[@]}"` dari setiap command `curl`.
- Install `jq` untuk output JSON yang rapi; jika belum ada, tetap bisa jalan tanpa format JSON.

## Environment Variables

- `APP_NAME` (default: `e-plan-ai`)
- `HTTP_ADDR` (default: `localhost:8080`)
- `MYSQL_DSN` (default: `root:root@tcp(localhost:3306)/e-plan-ai?parseTime=true`)
- `DB_MAX_OPEN_CONNS` (default: `20`)
- `DB_MAX_IDLE_CONNS` (default: `10`)
- `DB_CONN_MAX_LIFETIME_MINUTES` (default: `30`)
- `DB_CONN_MAX_IDLE_TIME_MINUTES` (default: `10`)

Contoh membuat database di MySQL localhost:

```bash
mysql -h localhost -u root -p -e "CREATE DATABASE IF NOT EXISTS \`e-plan-ai\`;"
```

## Catatan CRUD Storage

- Jika MySQL tersedia, CRUD modul menggunakan tabel `crud_records`.
- Jika MySQL belum tersedia, aplikasi fallback otomatis ke in-memory store.
- Endpoint list mendukung query: `page`, `limit`, `q`.
