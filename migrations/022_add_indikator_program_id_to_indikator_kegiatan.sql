-- Migration to add indikator_program_id to indikator_kegiatan and create FK
ALTER TABLE indikator_kegiatan
  ADD COLUMN indikator_program_id BIGINT UNSIGNED NULL AFTER id,
  ADD CONSTRAINT fk_indikator_kegiatan_program FOREIGN KEY (indikator_program_id) REFERENCES indikator_program(id);

-- Optional: create index for faster lookup
CREATE INDEX idx_indikator_kegiatan_program ON indikator_kegiatan(indikator_program_id);