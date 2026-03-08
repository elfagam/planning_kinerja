-- Incremental extension for hospital performance planning
-- This migration complements 001_init_schema.sql and 002_crud_records.sql
-- by adding governance and hospital organization structures not yet present.

CREATE TABLE IF NOT EXISTS hospitals (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    code VARCHAR(30) NOT NULL,
    name VARCHAR(200) NOT NULL,
    address TEXT NULL,
    phone VARCHAR(50) NULL,
    email VARCHAR(120) NULL,
    is_active TINYINT(1) NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uq_hospitals_code (code)
);

CREATE TABLE IF NOT EXISTS departments (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    hospital_id BIGINT UNSIGNED NOT NULL,
    code VARCHAR(30) NOT NULL,
    name VARCHAR(150) NOT NULL,
    description TEXT NULL,
    is_active TINYINT(1) NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT fk_departments_hospital FOREIGN KEY (hospital_id) REFERENCES hospitals(id),
    UNIQUE KEY uq_departments_code_per_hospital (hospital_id, code)
);

CREATE TABLE IF NOT EXISTS unit_department_links (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    unit_id BIGINT UNSIGNED NOT NULL,
    department_id BIGINT UNSIGNED NOT NULL,
    is_primary TINYINT(1) NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT fk_unit_dept_links_unit FOREIGN KEY (unit_id) REFERENCES units(id),
    CONSTRAINT fk_unit_dept_links_department FOREIGN KEY (department_id) REFERENCES departments(id),
    UNIQUE KEY uq_unit_department_pair (unit_id, department_id)
);

CREATE TABLE IF NOT EXISTS approval_flows (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    module_name VARCHAR(60) NOT NULL,
    step_order SMALLINT NOT NULL,
    role_id BIGINT UNSIGNED NOT NULL,
    step_name VARCHAR(120) NOT NULL,
    is_required TINYINT(1) NOT NULL DEFAULT 1,
    sla_hours INT UNSIGNED NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT fk_approval_flows_role FOREIGN KEY (role_id) REFERENCES roles(id),
    UNIQUE KEY uq_approval_flow_step (module_name, step_order)
);

CREATE TABLE IF NOT EXISTS approval_histories (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    module_name VARCHAR(60) NOT NULL,
    entity_id BIGINT UNSIGNED NOT NULL,
    step_order SMALLINT NOT NULL,
    decision ENUM('SUBMITTED', 'APPROVED', 'REJECTED', 'REVISED') NOT NULL,
    decided_by BIGINT UNSIGNED NOT NULL,
    decided_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    comments TEXT NULL,
    CONSTRAINT fk_approval_histories_decider FOREIGN KEY (decided_by) REFERENCES users(id),
    INDEX idx_approval_histories_entity (module_name, entity_id),
    INDEX idx_approval_histories_decided_at (decided_at)
);

CREATE TABLE IF NOT EXISTS planning_versions (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    module_name VARCHAR(60) NOT NULL,
    entity_id BIGINT UNSIGNED NOT NULL,
    version_no INT NOT NULL,
    snapshot_json JSON NOT NULL,
    changed_by BIGINT UNSIGNED NOT NULL,
    changed_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    reason VARCHAR(255) NULL,
    CONSTRAINT fk_planning_versions_user FOREIGN KEY (changed_by) REFERENCES users(id),
    UNIQUE KEY uq_planning_version (module_name, entity_id, version_no),
    INDEX idx_planning_versions_entity (module_name, entity_id)
);

CREATE TABLE IF NOT EXISTS indicator_data_sources (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    indikator_kinerja_id BIGINT UNSIGNED NOT NULL,
    source_name VARCHAR(120) NOT NULL,
    source_type ENUM('MANUAL', 'EMR', 'LIS', 'RIS', 'FINANCE', 'OTHER') NOT NULL DEFAULT 'MANUAL',
    endpoint_url VARCHAR(255) NULL,
    refresh_frequency ENUM('DAILY', 'WEEKLY', 'MONTHLY', 'QUARTERLY', 'YEARLY') NOT NULL DEFAULT 'MONTHLY',
    is_active TINYINT(1) NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT fk_indicator_data_source_indicator FOREIGN KEY (indikator_kinerja_id) REFERENCES indikator_kinerjas(id),
    INDEX idx_indicator_data_sources_indicator (indikator_kinerja_id)
);
