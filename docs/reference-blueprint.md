# Reference Blueprint: Enterprise E-Planning (Golang)

Dokumen ini adalah acuan implementasi untuk sistem perencanaan kinerja pemerintah dengan karakteristik:

- domain kompleks,
- audit ketat,
- workflow persetujuan berlapis,
- requirement perubahan regulasi periodik.

## 1. Pattern yang Direkomendasikan

- **Modular Monolith** sebagai default deployment model.
- **Clean Architecture** di level modul (`delivery -> usecase -> repository -> domain`).
- **DDD-style boundaries** untuk menghindari coupling antarmodul.

Alasan:

- Lebih aman untuk governance dan audit dibanding microservices dini.
- Biaya operasional lebih rendah.
- Tetap bisa diekstrak per modul ke service terpisah jika diperlukan.

## 2. Bounded Context Utama

1. `strategic-planning`

- Visi, Misi, Tujuan, Sasaran.

2. `program-management`

- Program, Kegiatan, Sub Kegiatan.

3. `performance-management`

- Indikator Kinerja, Target, Realisasi, capaian.

4. `renja-workflow`

- Draft -> Submit -> Approve/Reject -> Locked.

5. `iam`

- User, Role, Policy (RBAC).

6. `audit-reporting`

- Audit log, dashboard, export laporan.

## 3. Layering per Modul

```text
internal/modules/<module>/
  delivery/http/            # handler, DTO transport, route binding
  usecase/                  # orchestration + transaction boundary
  repository/               # SQL persistence
  domain/                   # entity, invariant, domain rule
```

Aturan:

- `delivery` tidak menyentuh SQL.
- `usecase` tidak tahu detail Gin.
- `repository` tidak mengandung aturan bisnis.
- `domain` tidak tahu DB/HTTP.

## 4. Transaction Boundaries (Kritis)

Gunakan transaksi di usecase untuk operasi yang harus atomik:

- Submit Renja (ubah status + catat jejak audit).
- Approve Renja (ubah status + lock period optional + audit).
- Reject Renja (ubah status + reason + audit).
- Revisi target indikator (history + nilai baru + audit).

## 5. Data Governance Rules

- Semua entitas planning punya `created_at`, `updated_at`, `deleted_at` (soft delete).
- Entitas workflow punya `status`, `approved_by`, `approved_at`, `rejected_reason`.
- Data periodik punya mekanisme `period lock`.
- Semua perubahan status wajib masuk `audit_logs`.

## 6. Error Contract API

Format response konsisten:

- success: `{ "success": true, "data": ... }`
- error: `{ "success": false, "error": "message" }`

Gunakan mapping error:

- domain validation -> `400`
- unauthorized -> `401`
- forbidden -> `403`
- not found -> `404`
- conflict/state invalid -> `409`
- unexpected -> `500`

## 7. Security Baseline

- JWT auth + RBAC middleware.
- Policy check di usecase untuk action sensitif.
- Audit actor identity di semua mutasi data.
- Sanitasi input + limit payload.

## 8. Path to Microservices (Jika Diperlukan)

Ekstraksi dilakukan per bounded context saat ada sinyal jelas:

- bottleneck scaling spesifik modul,
- tim ownership independen,
- siklus release terganggu karena coupling.

Kandidat pertama biasanya `audit-reporting`.
