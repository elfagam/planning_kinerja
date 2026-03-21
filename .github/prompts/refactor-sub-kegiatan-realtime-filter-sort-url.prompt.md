---
name: Refactor Sub Kegiatan Realtime Filter Sort URL
description: "Refactor web/templates/sub-kegiatan.html untuk filter realtime, sort kolom, dan sinkronisasi URL query state"
argument-hint: "Opsional: batasan tambahan atau field sort yang diizinkan"
agent: agent
---

Refactor `web/templates/sub-kegiatan.html` agar tabel/list mendukung stateful filtering dan sorting tanpa mengubah struktur visual utama.

Target perilaku:

- Filter realtime dari input pencarian dengan debounce `300ms`.
- Sort kolom dengan toggle `asc/desc` saat header diklik.
- Gunakan query param tetap: `q`, `sort_by`, `order`.
- Default awal saat query kosong: `sort_by=id` dan `order=desc`.
- Prioritaskan sort server-side: request list harus mengirim `sort_by` dan `order`.

Batasan:

- Jangan ubah struktur visual utama (layout/card/table/form tetap).
- Pertahankan class/style Bootstrap dan hierarchy HTML yang ada.
- Fokus perubahan pada JavaScript state/fetch/event handling.
- Hindari event listener duplikat dan request berulang yang tidak perlu.

Langkah kerja:

1. Audit alur list fetch, parsing query, dan binding event saat ini.
2. Tambahkan parser/writer `URLSearchParams` untuk `q`, `sort_by`, `order`.
3. Terapkan debounce `300ms` pada filter realtime.
4. Implement state sort per kolom dan indikator arah sort ringan (tanpa redesign).
5. Saat filter/sort berubah, update URL state lalu fetch data sinkron.
6. Jika backend belum siap sort server-side, aktifkan fallback sort client-side yang konsisten.
7. Verifikasi alur: load awal, ketik filter, klik sort, refresh, back/forward, dan share URL.

Output yang diminta:

- Edit langsung file terkait.
- Ringkasan perubahan state management dan event flow.
- Kontrak query param final (`q`, `sort_by`, `order`) serta default behavior.
- Catatan kompatibilitas backend/frontend.
- Checklist uji manual singkat.
