-- Seed demo users and role assignments
-- Safe to run multiple times (idempotent pattern).
-- IMPORTANT: Demo password hash below must be replaced for production use.

-- ------------------------------
-- Starter Units (compatible with 001_init_schema.sql)
-- ------------------------------
INSERT INTO units (code, name)
SELECT 'UNIT-MED', 'Unit Pelayanan Medis'
WHERE NOT EXISTS (SELECT 1 FROM units WHERE code = 'UNIT-MED');

INSERT INTO units (code, name)
SELECT 'UNIT-NUR', 'Unit Keperawatan'
WHERE NOT EXISTS (SELECT 1 FROM units WHERE code = 'UNIT-NUR');

INSERT INTO units (code, name)
SELECT 'UNIT-ADM', 'Unit Administrasi dan Keuangan'
WHERE NOT EXISTS (SELECT 1 FROM units WHERE code = 'UNIT-ADM');

-- Optional linkage to departments (table introduced in 003).
INSERT INTO unit_department_links (unit_id, department_id, is_primary)
SELECT u.id, d.id, 1
FROM units u
JOIN departments d ON d.code = 'MED'
WHERE u.code = 'UNIT-MED'
  AND NOT EXISTS (
    SELECT 1 FROM unit_department_links l
    WHERE l.unit_id = u.id AND l.department_id = d.id
  );

INSERT INTO unit_department_links (unit_id, department_id, is_primary)
SELECT u.id, d.id, 1
FROM units u
JOIN departments d ON d.code = 'NUR'
WHERE u.code = 'UNIT-NUR'
  AND NOT EXISTS (
    SELECT 1 FROM unit_department_links l
    WHERE l.unit_id = u.id AND l.department_id = d.id
  );

INSERT INTO unit_department_links (unit_id, department_id, is_primary)
SELECT u.id, d.id, 1
FROM units u
JOIN departments d ON d.code = 'ADM'
WHERE u.code = 'UNIT-ADM'
  AND NOT EXISTS (
    SELECT 1 FROM unit_department_links l
    WHERE l.unit_id = u.id AND l.department_id = d.id
  );

-- ------------------------------
-- Demo Users
-- bcrypt hash placeholder for password rotation on first login.
-- ------------------------------
INSERT INTO users (unit_id, full_name, email, password_hash, is_active)
SELECT u.id, 'Super Admin RSUD', 'superadmin@rsudcontoh.go.id', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 1
FROM units u
WHERE u.code = 'UNIT-ADM'
  AND NOT EXISTS (
    SELECT 1 FROM users x WHERE x.email = 'superadmin@rsudcontoh.go.id'
  );

INSERT INTO users (unit_id, full_name, email, password_hash, is_active)
SELECT u.id, 'Planner Medis', 'planner.med@rsudcontoh.go.id', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 1
FROM units u
WHERE u.code = 'UNIT-MED'
  AND NOT EXISTS (
    SELECT 1 FROM users x WHERE x.email = 'planner.med@rsudcontoh.go.id'
  );

INSERT INTO users (unit_id, full_name, email, password_hash, is_active)
SELECT u.id, 'Reviewer Keperawatan', 'reviewer.nur@rsudcontoh.go.id', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 1
FROM units u
WHERE u.code = 'UNIT-NUR'
  AND NOT EXISTS (
    SELECT 1 FROM users x WHERE x.email = 'reviewer.nur@rsudcontoh.go.id'
  );

INSERT INTO users (unit_id, full_name, email, password_hash, is_active)
SELECT u.id, 'Approver Direksi', 'approver@rsudcontoh.go.id', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 1
FROM units u
WHERE u.code = 'UNIT-ADM'
  AND NOT EXISTS (
    SELECT 1 FROM users x WHERE x.email = 'approver@rsudcontoh.go.id'
  );

INSERT INTO users (unit_id, full_name, email, password_hash, is_active)
SELECT u.id, 'Verifier Kinerja', 'verifier@rsudcontoh.go.id', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 1
FROM units u
WHERE u.code = 'UNIT-ADM'
  AND NOT EXISTS (
    SELECT 1 FROM users x WHERE x.email = 'verifier@rsudcontoh.go.id'
  );

-- ------------------------------
-- Role Assignments
-- ------------------------------
INSERT INTO user_roles (user_id, role_id)
SELECT u.id, r.id
FROM users u
JOIN roles r ON r.code = 'ADMIN'
WHERE u.email = 'superadmin@rsudcontoh.go.id'
  AND NOT EXISTS (
    SELECT 1 FROM user_roles ur
    WHERE ur.user_id = u.id AND ur.role_id = r.id
  );

INSERT INTO user_roles (user_id, role_id)
SELECT u.id, r.id
FROM users u
JOIN roles r ON r.code = 'PERENCANA'
WHERE u.email = 'planner.med@rsudcontoh.go.id'
  AND NOT EXISTS (
    SELECT 1 FROM user_roles ur
    WHERE ur.user_id = u.id AND ur.role_id = r.id
  );

INSERT INTO user_roles (user_id, role_id)
SELECT u.id, r.id
FROM users u
JOIN roles r ON r.code = 'REVIEWER'
WHERE u.email = 'reviewer.nur@rsudcontoh.go.id'
  AND NOT EXISTS (
    SELECT 1 FROM user_roles ur
    WHERE ur.user_id = u.id AND ur.role_id = r.id
  );

INSERT INTO user_roles (user_id, role_id)
SELECT u.id, r.id
FROM users u
JOIN roles r ON r.code = 'APPROVER'
WHERE u.email = 'approver@rsudcontoh.go.id'
  AND NOT EXISTS (
    SELECT 1 FROM user_roles ur
    WHERE ur.user_id = u.id AND ur.role_id = r.id
  );

INSERT INTO user_roles (user_id, role_id)
SELECT u.id, r.id
FROM users u
JOIN roles r ON r.code = 'VERIFIER'
WHERE u.email = 'verifier@rsudcontoh.go.id'
  AND NOT EXISTS (
    SELECT 1 FROM user_roles ur
    WHERE ur.user_id = u.id AND ur.role_id = r.id
  );
