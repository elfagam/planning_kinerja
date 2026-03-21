---
name: Checklist Regression Manual Template CRUD
description: "Buat checklist regression test manual untuk halaman template CRUD agar perubahan JS/UI tetap aman"
argument-hint: "Masukkan nama modul/template yang diuji"
agent: agent
---

Buat checklist regression test manual yang bisa dipakai berulang untuk halaman template CRUD.

Fokus checklist:

- Integritas load awal halaman dan data list.
- Alur CRUD utama (create, edit, delete, reset form).
- Filter/search, sort, dan state URL query string.
- Validasi input dan pesan error/sukses.
- Kesesuaian response API terhadap tampilan UI.

Format output wajib:

1. `Prasyarat`: data awal, akun, endpoint, dan environment.
2. `Daftar Skenario`: tabel atau bullet terstruktur berisi:
   - ID skenario
   - Langkah uji
   - Hasil yang diharapkan
3. `Uji Negatif`: input invalid, data tidak ditemukan, token/auth gagal, error server.
4. `Uji Kompatibilitas`: refresh, back/forward browser, copy-paste URL state.
5. `Catatan Risiko`: area yang belum tercover atau butuh automation.

Aturan:

- Gunakan bahasa Indonesia yang ringkas dan operasional.
- Checklist harus bisa dijalankan QA/dev tanpa penjelasan tambahan.
- Jika memungkinkan, petakan skenario ke endpoint yang diuji.

Output:

- Berikan checklist final siap pakai.
- Tambahkan rekomendasi 3 skenario prioritas tinggi untuk diotomasi sebagai integration test.
