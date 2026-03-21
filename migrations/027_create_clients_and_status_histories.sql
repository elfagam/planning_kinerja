-- Create clients and client_status_histories tables for client state machine workflow.
-- This migration is idempotent via CREATE TABLE IF NOT EXISTS.

CREATE TABLE IF NOT EXISTS clients (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    kode VARCHAR(50) NOT NULL,
    nama VARCHAR(255) NOT NULL,
    status ENUM('DRAFT', 'DIAJUKAN', 'DISETUJUI', 'DITOLAK') NOT NULL DEFAULT 'DRAFT',
    unit_pengusul_id BIGINT UNSIGNED NULL,
    created_by BIGINT UNSIGNED NULL,
    updated_by BIGINT UNSIGNED NULL,
    approved_by BIGINT UNSIGNED NULL,
    approved_at DATETIME NULL,
    rejected_by BIGINT UNSIGNED NULL,
    rejected_at DATETIME NULL,
    rejected_reason TEXT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME NULL,
    UNIQUE KEY uq_clients_kode (kode),
    INDEX idx_clients_status (status),
    INDEX idx_clients_unit_pengusul_id (unit_pengusul_id),
    INDEX idx_clients_deleted_at (deleted_at),
    CONSTRAINT fk_clients_unit_pengusul FOREIGN KEY (unit_pengusul_id) REFERENCES unit_pengusul(id),
    CONSTRAINT fk_clients_created_by FOREIGN KEY (created_by) REFERENCES users(id),
    CONSTRAINT fk_clients_updated_by FOREIGN KEY (updated_by) REFERENCES users(id),
    CONSTRAINT fk_clients_approved_by FOREIGN KEY (approved_by) REFERENCES users(id),
    CONSTRAINT fk_clients_rejected_by FOREIGN KEY (rejected_by) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS client_status_histories (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    client_id BIGINT UNSIGNED NOT NULL,
    from_status ENUM('DRAFT', 'DIAJUKAN', 'DISETUJUI', 'DITOLAK') NULL,
    to_status ENUM('DRAFT', 'DIAJUKAN', 'DISETUJUI', 'DITOLAK') NOT NULL,
    action VARCHAR(32) NOT NULL,
    reason TEXT NULL,
    note TEXT NULL,
    actor_id BIGINT UNSIGNED NULL,
    actor_name VARCHAR(255) NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_client_status_histories_client_id (client_id),
    INDEX idx_client_status_histories_created_at (created_at),
    CONSTRAINT fk_client_status_histories_client FOREIGN KEY (client_id) REFERENCES clients(id) ON DELETE CASCADE,
    CONSTRAINT fk_client_status_histories_actor FOREIGN KEY (actor_id) REFERENCES users(id)
);
