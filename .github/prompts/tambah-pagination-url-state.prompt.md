---
name: Tambah Pagination URL State
description: "Tambahkan pagination pada halaman/template CRUD dengan sinkronisasi state di URL query string"
argument-hint: "Opsional: param page/limit, default limit, dan strategi server/client"
agent: agent
---

Tambahkan fitur pagination pada halaman/template CRUD yang sudah memiliki list data, dengan state URL query string yang konsisten.

Tujuan:

- Gunakan query parameter URL untuk pagination: `page` dan `limit`.
- Pertahankan query parameter yang sudah ada dengan nama tetap `q`, `sort_by`, `order` agar tidak hilang saat pindah halaman.
- Tetapkan default sort awal `sort_by=id` dan `order=desc` jika query belum ada.
- Pastikan state tetap pulih saat halaman di-refresh atau URL dibuka ulang.

Batasan:

- Jangan ubah struktur visual utama halaman.
- Gunakan komponen UI yang sudah ada; jika perlu tambah kontrol pagination, buat minimal dan konsisten dengan Bootstrap yang dipakai.
- Prioritaskan pagination server-side (kirim `page`/`limit`), dengan fallback client-side jika endpoint belum mendukung.

Langkah kerja:

1. Audit alur fetch list yang ada dan sumber metadata total data.
2. Tambahkan state parser/writer URL untuk `page` dan `limit` tanpa merusak param lain.
3. Implement handler next/prev dan klik nomor halaman.
4. Pastikan perubahan filter/sort mereset `page` ke `1`.
5. Tampilkan info ringkas: halaman aktif, jumlah data, total jika tersedia.
6. Verifikasi alur manual: ubah page, ubah limit, refresh, copy URL, kembali dari browser history.

Output:

- Edit langsung file terkait.
- Ringkas perubahan state management pagination.
- Sebutkan kontrak parameter request/response yang dipakai (`page`, `limit`, total items/pages jika ada).
- Sertakan checklist uji manual singkat.
