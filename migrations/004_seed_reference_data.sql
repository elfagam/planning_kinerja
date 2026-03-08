-- Seed reference data for hospital performance planning
-- Safe to run multiple times (idempotent pattern).

-- ------------------------------
-- Default Roles
-- ------------------------------
INSERT INTO roles (code, name)
SELECT 'SUPER_ADMIN', 'Super Administrator'
WHERE NOT EXISTS (SELECT 1 FROM roles WHERE code = 'SUPER_ADMIN');

INSERT INTO roles (code, name)
SELECT 'PLANNER', 'Perencana Unit'
WHERE NOT EXISTS (SELECT 1 FROM roles WHERE code = 'PLANNER');

INSERT INTO roles (code, name)
SELECT 'REVIEWER', 'Reviewer Perencanaan'
WHERE NOT EXISTS (SELECT 1 FROM roles WHERE code = 'REVIEWER');

INSERT INTO roles (code, name)
SELECT 'APPROVER', 'Approver Perencanaan'
WHERE NOT EXISTS (SELECT 1 FROM roles WHERE code = 'APPROVER');

INSERT INTO roles (code, name)
SELECT 'VERIFIER', 'Verifier Kinerja'
WHERE NOT EXISTS (SELECT 1 FROM roles WHERE code = 'VERIFIER');

-- ------------------------------
-- Starter Hospital Organization
-- ------------------------------
INSERT INTO hospitals (code, name, address, phone, email, is_active)
SELECT 'RSUD-001', 'RSUD Contoh', 'Jl. Kesehatan No. 1', '021-5550001', 'admin@rsudcontoh.go.id', 1
WHERE NOT EXISTS (SELECT 1 FROM hospitals WHERE code = 'RSUD-001');

INSERT INTO departments (hospital_id, code, name, description, is_active)
SELECT h.id, 'MED', 'Pelayanan Medis', 'Departemen pelayanan medis', 1
FROM hospitals h
WHERE h.code = 'RSUD-001'
  AND NOT EXISTS (
    SELECT 1 FROM departments d
    WHERE d.hospital_id = h.id AND d.code = 'MED'
  );

INSERT INTO departments (hospital_id, code, name, description, is_active)
SELECT h.id, 'NUR', 'Keperawatan', 'Departemen pelayanan keperawatan', 1
FROM hospitals h
WHERE h.code = 'RSUD-001'
  AND NOT EXISTS (
    SELECT 1 FROM departments d
    WHERE d.hospital_id = h.id AND d.code = 'NUR'
  );

INSERT INTO departments (hospital_id, code, name, description, is_active)
SELECT h.id, 'ADM', 'Administrasi & Keuangan', 'Departemen administrasi dan keuangan', 1
FROM hospitals h
WHERE h.code = 'RSUD-001'
  AND NOT EXISTS (
    SELECT 1 FROM departments d
    WHERE d.hospital_id = h.id AND d.code = 'ADM'
  );

-- ------------------------------
-- Starter Periods
-- ------------------------------
INSERT INTO periods (year, quarter, month, status)
SELECT 2026, 1, NULL, 'OPEN'
WHERE NOT EXISTS (
    SELECT 1 FROM periods
    WHERE year = 2026 AND quarter = 1 AND month IS NULL
);

INSERT INTO periods (year, quarter, month, status)
SELECT 2026, 2, NULL, 'OPEN'
WHERE NOT EXISTS (
    SELECT 1 FROM periods
    WHERE year = 2026 AND quarter = 2 AND month IS NULL
);

INSERT INTO periods (year, quarter, month, status)
SELECT 2026, 3, NULL, 'OPEN'
WHERE NOT EXISTS (
    SELECT 1 FROM periods
    WHERE year = 2026 AND quarter = 3 AND month IS NULL
);

INSERT INTO periods (year, quarter, month, status)
SELECT 2026, 4, NULL, 'OPEN'
WHERE NOT EXISTS (
    SELECT 1 FROM periods
    WHERE year = 2026 AND quarter = 4 AND month IS NULL
);

-- ------------------------------
-- Approval Flow Templates
-- Uses roles seeded above and applies to core modules.
-- ------------------------------
INSERT INTO approval_flows (module_name, step_order, role_id, step_name, is_required, sla_hours)
SELECT 'RENJA', 1, r.id, 'Submit by Planner', 1, 24
FROM roles r
WHERE r.code = 'PLANNER'
  AND NOT EXISTS (
    SELECT 1 FROM approval_flows af
    WHERE af.module_name = 'RENJA' AND af.step_order = 1
  );

INSERT INTO approval_flows (module_name, step_order, role_id, step_name, is_required, sla_hours)
SELECT 'RENJA', 2, r.id, 'Review by Reviewer', 1, 48
FROM roles r
WHERE r.code = 'REVIEWER'
  AND NOT EXISTS (
    SELECT 1 FROM approval_flows af
    WHERE af.module_name = 'RENJA' AND af.step_order = 2
  );

INSERT INTO approval_flows (module_name, step_order, role_id, step_name, is_required, sla_hours)
SELECT 'RENJA', 3, r.id, 'Approval by Approver', 1, 48
FROM roles r
WHERE r.code = 'APPROVER'
  AND NOT EXISTS (
    SELECT 1 FROM approval_flows af
    WHERE af.module_name = 'RENJA' AND af.step_order = 3
  );

INSERT INTO approval_flows (module_name, step_order, role_id, step_name, is_required, sla_hours)
SELECT 'TARGET_REALISASI', 1, r.id, 'Input by Planner', 1, 24
FROM roles r
WHERE r.code = 'PLANNER'
  AND NOT EXISTS (
    SELECT 1 FROM approval_flows af
    WHERE af.module_name = 'TARGET_REALISASI' AND af.step_order = 1
  );

INSERT INTO approval_flows (module_name, step_order, role_id, step_name, is_required, sla_hours)
SELECT 'TARGET_REALISASI', 2, r.id, 'Verification by Verifier', 1, 48
FROM roles r
WHERE r.code = 'VERIFIER'
  AND NOT EXISTS (
    SELECT 1 FROM approval_flows af
    WHERE af.module_name = 'TARGET_REALISASI' AND af.step_order = 2
  );
