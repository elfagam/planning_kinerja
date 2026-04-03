-- Add dibuat_oleh column to indikator_rencana_kerja table
ALTER TABLE `indikator_rencana_kerja` 
ADD COLUMN `dibuat_oleh` BIGINT UNSIGNED NOT NULL AFTER `anggaran_tahunan`;

-- Add index for better search and filter performance
CREATE INDEX `idx_indikator_rk_dibuat_oleh` ON `indikator_rencana_kerja` (`dibuat_oleh`);
