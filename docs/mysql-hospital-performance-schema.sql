-- MySQL 8.0+ schema design for Hospital Performance Planning System
-- Domain coverage: strategic planning hierarchy, Renja, KPI target vs realization,
-- governance workflow, and auditability.

CREATE DATABASE IF NOT EXISTS hospital_performance_planning
  CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci;

USE hospital_performance_planning;

-- ------------------------------
-- Master & Organization
-- ------------------------------
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

CREATE TABLE IF NOT EXISTS units (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  department_id BIGINT UNSIGNED NOT NULL,
  code VARCHAR(30) NOT NULL,
  name VARCHAR(150) NOT NULL,
  unit_type ENUM('MEDICAL', 'NURSING', 'SUPPORT', 'ADMIN', 'LAB', 'PHARMACY') NOT NULL DEFAULT 'SUPPORT',
  is_active TINYINT(1) NOT NULL DEFAULT 1,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  CONSTRAINT fk_units_department FOREIGN KEY (department_id) REFERENCES departments(id),
  UNIQUE KEY uq_units_code_per_department (department_id, code)
);

CREATE TABLE IF NOT EXISTS periods (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  fiscal_year SMALLINT NOT NULL,
  quarter TINYINT NULL,
  month TINYINT NULL,
  status ENUM('OPEN', 'LOCKED', 'CLOSED') NOT NULL DEFAULT 'OPEN',
  start_date DATE NULL,
  end_date DATE NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uq_period_dimension (fiscal_year, quarter, month)
);

-- ------------------------------
-- IAM & Governance
-- ------------------------------
CREATE TABLE IF NOT EXISTS roles (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  code VARCHAR(40) NOT NULL,
  name VARCHAR(120) NOT NULL,
  description VARCHAR(255) NULL,
  UNIQUE KEY uq_roles_code (code)
);

CREATE TABLE IF NOT EXISTS users (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  unit_id BIGINT UNSIGNED NULL,
  full_name VARCHAR(150) NOT NULL,
  email VARCHAR(150) NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  is_active TINYINT(1) NOT NULL DEFAULT 1,
  last_login_at DATETIME NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  CONSTRAINT fk_users_unit FOREIGN KEY (unit_id) REFERENCES units(id),
  UNIQUE KEY uq_users_email (email)
);

CREATE TABLE IF NOT EXISTS user_roles (
  user_id BIGINT UNSIGNED NOT NULL,
  role_id BIGINT UNSIGNED NOT NULL,
  PRIMARY KEY (user_id, role_id),
  CONSTRAINT fk_user_roles_user FOREIGN KEY (user_id) REFERENCES users(id),
  CONSTRAINT fk_user_roles_role FOREIGN KEY (role_id) REFERENCES roles(id)
);

-- ------------------------------
-- Strategic Planning Hierarchy
-- ------------------------------
CREATE TABLE IF NOT EXISTS visions (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  hospital_id BIGINT UNSIGNED NOT NULL,
  code VARCHAR(30) NOT NULL,
  statement VARCHAR(255) NOT NULL,
  start_year SMALLINT NOT NULL,
  end_year SMALLINT NOT NULL,
  is_active TINYINT(1) NOT NULL DEFAULT 1,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at DATETIME NULL,
  CONSTRAINT fk_visions_hospital FOREIGN KEY (hospital_id) REFERENCES hospitals(id),
  UNIQUE KEY uq_visions_code_per_hospital (hospital_id, code)
);

CREATE TABLE IF NOT EXISTS missions (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  vision_id BIGINT UNSIGNED NOT NULL,
  code VARCHAR(30) NOT NULL,
  statement VARCHAR(255) NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at DATETIME NULL,
  CONSTRAINT fk_missions_vision FOREIGN KEY (vision_id) REFERENCES visions(id),
  UNIQUE KEY uq_missions_code_per_vision (vision_id, code)
);

CREATE TABLE IF NOT EXISTS goals (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  mission_id BIGINT UNSIGNED NOT NULL,
  code VARCHAR(30) NOT NULL,
  statement VARCHAR(255) NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at DATETIME NULL,
  CONSTRAINT fk_goals_mission FOREIGN KEY (mission_id) REFERENCES missions(id),
  UNIQUE KEY uq_goals_code_per_mission (mission_id, code)
);

CREATE TABLE IF NOT EXISTS goal_indicators (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  goal_id BIGINT UNSIGNED NOT NULL,
  code VARCHAR(40) NOT NULL,
  name VARCHAR(255) NOT NULL,
  unit VARCHAR(60) NULL,
  formula TEXT NULL,
  baseline DECIMAL(18,2) NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at DATETIME NULL,
  CONSTRAINT fk_goal_indicators_goal FOREIGN KEY (goal_id) REFERENCES goals(id),
  UNIQUE KEY uq_goal_indicators_code_per_goal (goal_id, code)
);

CREATE TABLE IF NOT EXISTS objectives (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  goal_id BIGINT UNSIGNED NOT NULL,
  code VARCHAR(30) NOT NULL,
  statement VARCHAR(255) NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at DATETIME NULL,
  CONSTRAINT fk_objectives_goal FOREIGN KEY (goal_id) REFERENCES goals(id),
  UNIQUE KEY uq_objectives_code_per_goal (goal_id, code)
);

CREATE TABLE IF NOT EXISTS objective_indicators (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  objective_id BIGINT UNSIGNED NOT NULL,
  code VARCHAR(40) NOT NULL,
  name VARCHAR(255) NOT NULL,
  unit VARCHAR(60) NULL,
  formula TEXT NULL,
  baseline DECIMAL(18,2) NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at DATETIME NULL,
  CONSTRAINT fk_objective_indicators_objective FOREIGN KEY (objective_id) REFERENCES objectives(id),
  UNIQUE KEY uq_obj_indicators_code_per_objective (objective_id, code)
);

CREATE TABLE IF NOT EXISTS programs (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  objective_id BIGINT UNSIGNED NOT NULL,
  owner_unit_id BIGINT UNSIGNED NOT NULL,
  code VARCHAR(40) NOT NULL,
  name VARCHAR(255) NOT NULL,
  description TEXT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at DATETIME NULL,
  CONSTRAINT fk_programs_objective FOREIGN KEY (objective_id) REFERENCES objectives(id),
  CONSTRAINT fk_programs_owner_unit FOREIGN KEY (owner_unit_id) REFERENCES units(id),
  UNIQUE KEY uq_programs_code_per_unit (owner_unit_id, code)
);

CREATE TABLE IF NOT EXISTS program_indicators (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  program_id BIGINT UNSIGNED NOT NULL,
  code VARCHAR(40) NOT NULL,
  name VARCHAR(255) NOT NULL,
  unit VARCHAR(60) NULL,
  formula TEXT NULL,
  baseline DECIMAL(18,2) NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at DATETIME NULL,
  CONSTRAINT fk_program_indicators_program FOREIGN KEY (program_id) REFERENCES programs(id),
  UNIQUE KEY uq_prog_indicators_code_per_program (program_id, code)
);

CREATE TABLE IF NOT EXISTS activities (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  program_id BIGINT UNSIGNED NOT NULL,
  code VARCHAR(40) NOT NULL,
  name VARCHAR(255) NOT NULL,
  description TEXT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at DATETIME NULL,
  CONSTRAINT fk_activities_program FOREIGN KEY (program_id) REFERENCES programs(id),
  UNIQUE KEY uq_activities_code_per_program (program_id, code)
);

CREATE TABLE IF NOT EXISTS activity_indicators (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  activity_id BIGINT UNSIGNED NOT NULL,
  code VARCHAR(40) NOT NULL,
  name VARCHAR(255) NOT NULL,
  unit VARCHAR(60) NULL,
  formula TEXT NULL,
  baseline DECIMAL(18,2) NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at DATETIME NULL,
  CONSTRAINT fk_activity_indicators_activity FOREIGN KEY (activity_id) REFERENCES activities(id),
  UNIQUE KEY uq_act_indicators_code_per_activity (activity_id, code)
);

CREATE TABLE IF NOT EXISTS sub_activities (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  activity_id BIGINT UNSIGNED NOT NULL,
  code VARCHAR(40) NOT NULL,
  name VARCHAR(255) NOT NULL,
  description TEXT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at DATETIME NULL,
  CONSTRAINT fk_sub_activities_activity FOREIGN KEY (activity_id) REFERENCES activities(id),
  UNIQUE KEY uq_sub_activities_code_per_activity (activity_id, code)
);

CREATE TABLE IF NOT EXISTS sub_activity_indicators (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  sub_activity_id BIGINT UNSIGNED NOT NULL,
  code VARCHAR(40) NOT NULL,
  name VARCHAR(255) NOT NULL,
  unit VARCHAR(60) NULL,
  formula TEXT NULL,
  baseline DECIMAL(18,2) NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at DATETIME NULL,
  CONSTRAINT fk_sub_activity_indicators_sub_activity FOREIGN KEY (sub_activity_id) REFERENCES sub_activities(id),
  UNIQUE KEY uq_sub_act_indicators_code_per_sub_activity (sub_activity_id, code)
);

-- ------------------------------
-- Annual Work Plan (Renja)
-- ------------------------------
CREATE TABLE IF NOT EXISTS renja_documents (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  period_id BIGINT UNSIGNED NOT NULL,
  unit_id BIGINT UNSIGNED NOT NULL,
  doc_number VARCHAR(60) NOT NULL,
  title VARCHAR(255) NOT NULL,
  status ENUM('DRAFT', 'SUBMITTED', 'UNDER_REVIEW', 'APPROVED', 'REJECTED') NOT NULL DEFAULT 'DRAFT',
  notes TEXT NULL,
  created_by BIGINT UNSIGNED NOT NULL,
  updated_by BIGINT UNSIGNED NULL,
  submitted_at DATETIME NULL,
  approved_at DATETIME NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at DATETIME NULL,
  CONSTRAINT fk_renja_docs_period FOREIGN KEY (period_id) REFERENCES periods(id),
  CONSTRAINT fk_renja_docs_unit FOREIGN KEY (unit_id) REFERENCES units(id),
  CONSTRAINT fk_renja_docs_created_by FOREIGN KEY (created_by) REFERENCES users(id),
  CONSTRAINT fk_renja_docs_updated_by FOREIGN KEY (updated_by) REFERENCES users(id),
  UNIQUE KEY uq_renja_doc_number (doc_number)
);

CREATE TABLE IF NOT EXISTS renja_items (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  renja_document_id BIGINT UNSIGNED NOT NULL,
  program_id BIGINT UNSIGNED NULL,
  activity_id BIGINT UNSIGNED NULL,
  sub_activity_id BIGINT UNSIGNED NULL,
  budget DECIMAL(18,2) NOT NULL DEFAULT 0,
  volume DECIMAL(18,2) NOT NULL DEFAULT 0,
  uom VARCHAR(50) NULL,
  output_description VARCHAR(255) NULL,
  location VARCHAR(120) NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  CONSTRAINT fk_renja_items_doc FOREIGN KEY (renja_document_id) REFERENCES renja_documents(id),
  CONSTRAINT fk_renja_items_program FOREIGN KEY (program_id) REFERENCES programs(id),
  CONSTRAINT fk_renja_items_activity FOREIGN KEY (activity_id) REFERENCES activities(id),
  CONSTRAINT fk_renja_items_sub_activity FOREIGN KEY (sub_activity_id) REFERENCES sub_activities(id),
  INDEX idx_renja_items_doc (renja_document_id)
);

-- ------------------------------
-- Performance Measurement
-- ------------------------------
CREATE TABLE IF NOT EXISTS performance_indicators (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  code VARCHAR(40) NOT NULL,
  name VARCHAR(255) NOT NULL,
  source_type ENUM('GOAL', 'OBJECTIVE', 'PROGRAM', 'ACTIVITY', 'SUB_ACTIVITY', 'SERVICE') NOT NULL,
  source_id BIGINT UNSIGNED NOT NULL,
  unit VARCHAR(60) NULL,
  formula TEXT NULL,
  data_source VARCHAR(120) NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uq_performance_indicators_code (code)
);

CREATE TABLE IF NOT EXISTS target_realizations (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  performance_indicator_id BIGINT UNSIGNED NOT NULL,
  period_id BIGINT UNSIGNED NOT NULL,
  target_value DECIMAL(18,2) NOT NULL DEFAULT 0,
  realization_value DECIMAL(18,2) NOT NULL DEFAULT 0,
  deviation_value DECIMAL(18,2) GENERATED ALWAYS AS (realization_value - target_value) STORED,
  achievement_percent DECIMAL(8,2) GENERATED ALWAYS AS (
    CASE
      WHEN target_value = 0 THEN 0
      ELSE (realization_value / target_value) * 100
    END
  ) STORED,
  performance_status ENUM('ON_TRACK', 'WARNING', 'OFF_TRACK') NOT NULL DEFAULT 'ON_TRACK',
  verification_status ENUM('DRAFT', 'VERIFIED') NOT NULL DEFAULT 'DRAFT',
  verified_by BIGINT UNSIGNED NULL,
  verified_at DATETIME NULL,
  notes TEXT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  CONSTRAINT fk_target_realization_indicator FOREIGN KEY (performance_indicator_id) REFERENCES performance_indicators(id),
  CONSTRAINT fk_target_realization_period FOREIGN KEY (period_id) REFERENCES periods(id),
  CONSTRAINT fk_target_realization_verifier FOREIGN KEY (verified_by) REFERENCES users(id),
  UNIQUE KEY uq_target_realization (performance_indicator_id, period_id)
);

-- ------------------------------
-- Approval Workflow & Versioning
-- ------------------------------
CREATE TABLE IF NOT EXISTS approval_flows (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  module_name VARCHAR(60) NOT NULL,
  step_order SMALLINT NOT NULL,
  role_id BIGINT UNSIGNED NOT NULL,
  step_name VARCHAR(120) NOT NULL,
  is_required TINYINT(1) NOT NULL DEFAULT 1,
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
  INDEX idx_approval_histories_entity (module_name, entity_id)
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
  UNIQUE KEY uq_planning_version (module_name, entity_id, version_no)
);

-- ------------------------------
-- Audit Log
-- ------------------------------
CREATE TABLE IF NOT EXISTS audit_logs (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NULL,
  action VARCHAR(100) NOT NULL,
  module_name VARCHAR(60) NOT NULL,
  entity_id BIGINT UNSIGNED NULL,
  request_payload JSON NULL,
  response_code SMALLINT NULL,
  ip_address VARCHAR(45) NULL,
  user_agent VARCHAR(255) NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_audit_logs_user FOREIGN KEY (user_id) REFERENCES users(id),
  INDEX idx_audit_logs_module_entity (module_name, entity_id),
  INDEX idx_audit_logs_created_at (created_at)
);
