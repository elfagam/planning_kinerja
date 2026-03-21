-- Add OPERATOR role option to users.role enum.
-- Keep existing default as PERENCANA.

ALTER TABLE users
  MODIFY COLUMN role ENUM('ADMIN','OPERATOR','PERENCANA','VERIFIKATOR','PIMPINAN')
  NOT NULL
  DEFAULT 'PERENCANA';
