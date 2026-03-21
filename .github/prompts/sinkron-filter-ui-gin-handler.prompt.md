---
name: Sinkron Filter UI Gin Handler
description: "Sinkronkan parameter filter/sort dari UI template dengan endpoint Gin handler agar perilaku konsisten end-to-end"
argument-hint: "Masukkan path template + path handler/repository target"
agent: agent
---

Sinkronkan filter dan sort antara UI template dan backend Gin handler secara end-to-end.

Konteks target:

- Frontend mengirim parameter query tetap: `q`, `sort_by`, `order`.
- Handler Gin harus membaca, memvalidasi, dan meneruskan parameter ke layer query/repository.
- Sort diprioritaskan server-side dengan kontrak `sort_by` + `order` pada setiap request list.

Tujuan:

- Kontrak parameter frontend-backend konsisten dan terdokumentasi.
- Nilai default aman ditetapkan di backend jika parameter kosong/tidak valid.
- Default sort fallback backend: `sort_by=id`, `order=desc` (kecuali modul mendefinisikan override eksplisit).
- Hindari mismatch nama field sort antara UI dan database model.

Langkah kerja:

1. Telaah template/JS yang membentuk URL request list.
2. Telaah handler Gin, service, dan repository query builder.
3. Terapkan validasi `sort_by` berbasis allowlist field yang aman.
4. Terapkan validasi `order` hanya `asc|desc` (default sesuai kebutuhan modul).
5. Terapkan pencarian `q` yang konsisten pada field yang relevan.
6. Pastikan response menyertakan metadata yang berguna untuk UI bila tersedia.
7. Perbarui dokumentasi endpoint bila kontrak berubah.

Kriteria hasil:

- UI dan backend menggunakan nama query param yang sama.
- Tidak ada SQL injection risk dari dynamic sort.
- Default behavior jelas dan konsisten lintas layer.

Output:

- Edit file terkait pada frontend dan backend.
- Ringkasan perubahan per layer (template, handler, repository).
- Daftar allowlist `sort_by` yang dipakai.
- Checklist verifikasi manual request URL dan hasil response.
