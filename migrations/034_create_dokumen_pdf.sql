-- Migration: Create dokumen_pdf table
CREATE TABLE dokumen_pdf (
    id CHAR(36) PRIMARY KEY DEFAULT (UUID()),
    tahun INT NOT NULL,
    file_path VARCHAR(255) NOT NULL
);