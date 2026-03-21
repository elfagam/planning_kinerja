-- Migration: Add column 'nama' to dokumen_pdf
ALTER TABLE dokumen_pdf
ADD COLUMN nama CHAR(255) AFTER tahun;
