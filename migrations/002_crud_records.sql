CREATE TABLE IF NOT EXISTS crud_records (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    resource VARCHAR(64) NOT NULL,
    code VARCHAR(50) NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT NULL,
    attributes JSON NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_crud_records_resource (resource),
    INDEX idx_crud_records_name (name),
    INDEX idx_crud_records_code (code)
);
