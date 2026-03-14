# API Endpoints E-Planning RSUD

Base path: `/api/v1`

Query params untuk endpoint list CRUD:

- `page` (default `1`)
- `limit` (default `10`, max `100`)
- `q` (search by `name` atau `code`)

## Health

- `GET /health`

## Auth

- `POST /auth/login`
- `POST /auth/refresh`
- `POST /auth/logout`
- `GET /auth/me`

## Master Hierarchy Planning

- `GET /planning/hierarchy`

### Visi

- `GET /visi`
- `POST /visi`
- `GET /visi/:id`
- `PUT /visi/:id`
- `DELETE /visi/:id`

### Misi

- `GET /misi`
- `POST /misi`
- `GET /misi/:id`
- `PUT /misi/:id`
- `DELETE /misi/:id`

### Tujuan

- `GET /tujuan`
- `POST /tujuan`
- `GET /tujuan/:id`
- `PUT /tujuan/:id`
- `DELETE /tujuan/:id`

### Indikator Tujuan

- `GET /indikator-tujuan`
- `POST /indikator-tujuan`
- `GET /indikator-tujuan/:id`
- `PUT /indikator-tujuan/:id`
- `DELETE /indikator-tujuan/:id`

### Sasaran

- `GET /sasaran`
- `POST /sasaran`
- `GET /sasaran/:id`
- `PUT /sasaran/:id`
- `DELETE /sasaran/:id`

### Indikator Sasaran

- `GET /indikator-sasaran`
- `POST /indikator-sasaran`
- `GET /indikator-sasaran/:id`
- `PUT /indikator-sasaran/:id`
- `DELETE /indikator-sasaran/:id`

### Program

- `GET /program`
- `POST /program`
- `GET /program/:id`
- `PUT /program/:id`
- `DELETE /program/:id`

### Indikator Program

- `GET /indikator-program`
- `POST /indikator-program`
- `GET /indikator-program/:id`
- `PUT /indikator-program/:id`
- `DELETE /indikator-program/:id`

### Kegiatan

- `GET /kegiatan`
- `POST /kegiatan`
- `GET /kegiatan/:id`
- `PUT /kegiatan/:id`
- `DELETE /kegiatan/:id`

### Indikator Kegiatan

- `GET /indikator-kegiatan`
- `POST /indikator-kegiatan`
- `GET /indikator-kegiatan/:id`
- `PUT /indikator-kegiatan/:id`
- `DELETE /indikator-kegiatan/:id`

### Sub Kegiatan

- `GET /sub-kegiatan`
- `POST /sub-kegiatan`
- `GET /sub-kegiatan/:id`
- `PUT /sub-kegiatan/:id`
- `DELETE /sub-kegiatan/:id`

### Indikator Sub Kegiatan

- `GET /indikator-sub-kegiatan`
- `POST /indikator-sub-kegiatan`
- `GET /indikator-sub-kegiatan/:id`
- `PUT /indikator-sub-kegiatan/:id`
- `DELETE /indikator-sub-kegiatan/:id`

## Renja

- `GET /renja`
- `POST /renja`
- `GET /renja/:id`
- `PUT /renja/:id`
- `DELETE /renja/:id`
- `GET /renja/overview`
- `POST /renja/:id/submit`
- `POST /renja/:id/approve`
- `POST /renja/:id/reject`

## Clients

- `GET /clients`
- `POST /clients`
- `GET /clients/:id`
- `PUT /clients/:id`
- `DELETE /clients/:id`
- `POST /clients/:id/submit`
- `POST /clients/:id/unsubmit`
- `POST /clients/:id/reject`
- `POST /clients/:id/re-evaluate`
- `POST /clients/:id/approve`
- `GET /clients/:id/status-history`

Catatan: endpoint mutasi Client mengambil actor dari JWT auth context (`auth.user_id`, `auth.role`, `auth.full_name`), bukan dari payload JSON.

## Indikator Kinerja

- `GET /indikator-kinerja`
- `POST /indikator-kinerja`
- `GET /indikator-kinerja/:id`
- `PUT /indikator-kinerja/:id`
- `DELETE /indikator-kinerja/:id`

## Target dan Realisasi

- `GET /performance/target-realisasi`
- `GET /performance/dashboard-summary`
- `GET /performance/statistics`
- `GET /performance/chart-target-vs-realisasi`
- `GET /performance/yearly-summary`
- `GET /performance/program-ranking`
- `GET /target-realisasi`
- `POST /target-realisasi`
- `GET /target-realisasi/:id`
- `PUT /target-realisasi/:id`
- `DELETE /target-realisasi/:id`
- `POST /target-realisasi/:id/verifikasi`

## Reporting

- `GET /reports/dashboard`
- `GET /reports/capaian?year=2026&period=Q1`
- `GET /reports/export/excel?year=2026`
- `GET /reports/export/pdf?year=2026`

## UI Pages (Separated CRUD)

- `GET /ui`
- `GET /ui/visi`
- `GET /ui/misi`
- `GET /ui/tujuan`
- `GET /ui/indikator-tujuan`
- `GET /ui/sasaran`
- `GET /ui/indikator-sasaran`
- `GET /ui/program`
- `GET /ui/indikator-program`
- `GET /ui/kegiatan`
- `GET /ui/indikator-kegiatan`
- `GET /ui/sub-kegiatan`
- `GET /ui/indikator-sub-kegiatan`
- `GET /ui/renja`
- `GET /ui/indikator-kinerja`
- `GET /ui/target-realisasi`
- `GET /ui/dashboard`
