-- users.role is the single source of truth for runtime authorization.
-- user_roles is kept for backward compatibility and synced from users.role.
-- Idempotent migration: safe to run multiple times.

-- Ensure users.role exists (temporary VARCHAR to allow normalization before ENUM cast).
SET @role_col_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'users'
    AND COLUMN_NAME = 'role'
);
SET @ddl := IF(
  @role_col_exists = 0,
  'ALTER TABLE users ADD COLUMN role VARCHAR(30) NOT NULL DEFAULT ''PERENCANA'' AFTER email',
  'SELECT 1'
);
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- Normalize legacy role values to canonical runtime role codes.
UPDATE users
SET role = CASE
  WHEN UPPER(TRIM(COALESCE(role, ''))) IN ('ADMIN', 'SUPER_ADMIN', 'SUPERADMIN') THEN 'ADMIN'
  WHEN UPPER(TRIM(COALESCE(role, ''))) IN ('OPERATOR') THEN 'OPERATOR'
  WHEN UPPER(TRIM(COALESCE(role, ''))) IN ('PERENCANA', 'PLANNER') THEN 'PERENCANA'
  WHEN UPPER(TRIM(COALESCE(role, ''))) IN ('VERIFIKATOR', 'VERIFIER', 'REVIEWER') THEN 'VERIFIKATOR'
  WHEN UPPER(TRIM(COALESCE(role, ''))) IN ('PIMPINAN', 'APPROVER') THEN 'PIMPINAN'
  ELSE 'PERENCANA'
END;

-- Enforce constrained runtime role set.
ALTER TABLE users
  MODIFY COLUMN role ENUM('ADMIN','OPERATOR','PERENCANA','VERIFIKATOR','PIMPINAN')
  NOT NULL
  DEFAULT 'PERENCANA';

-- Ensure role master rows exist.
INSERT INTO roles (code, name)
SELECT 'ADMIN', 'Administrator'
WHERE NOT EXISTS (SELECT 1 FROM roles WHERE code = 'ADMIN');

INSERT INTO roles (code, name)
SELECT 'OPERATOR', 'Operator Unit'
WHERE NOT EXISTS (SELECT 1 FROM roles WHERE code = 'OPERATOR');

INSERT INTO roles (code, name)
SELECT 'PERENCANA', 'Perencana Unit'
WHERE NOT EXISTS (SELECT 1 FROM roles WHERE code = 'PERENCANA');

INSERT INTO roles (code, name)
SELECT 'VERIFIKATOR', 'Verifikator'
WHERE NOT EXISTS (SELECT 1 FROM roles WHERE code = 'VERIFIKATOR');

INSERT INTO roles (code, name)
SELECT 'PIMPINAN', 'Pimpinan'
WHERE NOT EXISTS (SELECT 1 FROM roles WHERE code = 'PIMPINAN');

-- Keep user_roles as deprecated compatibility data synced from users.role.
SET @user_roles_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.TABLES
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'user_roles'
);

SET @dml := IF(
  @user_roles_exists = 1,
  'DELETE ur FROM user_roles ur JOIN roles r ON r.id = ur.role_id WHERE r.code IN (''ADMIN'',''OPERATOR'',''PERENCANA'',''VERIFIKATOR'',''PIMPINAN'')',
  'SELECT 1'
);
PREPARE stmt FROM @dml; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @dml := IF(
  @user_roles_exists = 1,
  'INSERT IGNORE INTO user_roles (user_id, role_id) SELECT u.id, r.id FROM users u JOIN roles r ON r.code = u.role WHERE u.role IN (''ADMIN'',''OPERATOR'',''PERENCANA'',''VERIFIKATOR'',''PIMPINAN'')',
  'SELECT 1'
);
PREPARE stmt FROM @dml; EXECUTE stmt; DEALLOCATE PREPARE stmt;
