-- Migration to drop foreign key constraint dan kolom dari tabel program

ALTER TABLE program
DROP FOREIGN KEY fk_id_program_unit_pengusul;

ALTER TABLE program
DROP COLUMN sasaran_id,
DROP COLUMN unit_pengusul_id;
