---
name: Refactor Indikator Kegiatan Stateful Table
description: "Refactor web/templates/indikator-kegiatan.html untuk realtime filter, sort kolom, dan URL query state tanpa ubah struktur visual utama"
argument-hint: "Opsional: override kecil jika diperlukan (default: id desc, debounce 300ms)"
agent: agent
---

Refactor `web/templates/indikator-kegiatan.html` agar daftar data mendukung:

- Filter realtime dari input pencarian (`#query`) dengan debounce `300ms`.
- Sort kolom yang bisa di-toggle asc/desc saat header kolom diklik.
- Sinkronisasi state ke URL query string dengan parameter tetap: `q`, `sort_by`, `order` agar bisa di-refresh/share tanpa kehilangan state.
- Tetap mempertahankan struktur visual utama, class Bootstrap, dan hierarchy HTML yang ada.

Batasan implementasi:

- Jangan ubah tampilan utama atau layout card/form/table secara signifikan.
- Utamakan perubahan pada JavaScript inline yang sudah ada.
- Pertahankan kompatibilitas endpoint `data-api-endpoint` saat ini.
- Prioritaskan sort server-side dengan selalu mengirim `sort_by` dan `order` pada request list.
- Jika backend belum dukung sort server-side, lakukan fallback sort di client-side secara konsisten.
- Hindari duplikasi event listener dan pastikan tidak memicu fetch berulang yang tidak perlu.
- Gunakan default state awal: `sort_by=id` dan `order=desc`.

Langkah kerja yang diharapkan:

1. Analisis state saat ini pada query/search/list rendering.
2. Tambahkan parser+writer URL state (`URLSearchParams`) untuk inisialisasi dan update state.
3. Implement debounce pada filter realtime.
4. Implement sort state, indikator arah sort di header tabel (tanpa redesign layout), lalu kirim `sort_by/order` ke backend saat fetch list.
5. Pastikan aksi `Cari / Refresh` tetap berfungsi dan sinkron dengan state baru.
6. Pastikan load awal mengisi state default `id desc` bila query param kosong.
7. Uji manual alur: load awal, ketik filter, klik sort, refresh halaman, dan copy URL.

Output yang diminta:

- Lakukan edit langsung pada file terkait.
- Berikan ringkasan perubahan berisi:
  - Perubahan state management.
  - Daftar query parameter yang digunakan.
  - Potensi dampak kompatibilitas backend/frontend.
  - Checklist uji manual singkat.
