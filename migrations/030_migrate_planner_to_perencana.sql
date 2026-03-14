-- Normalize legacy PLANNER role into PERENCANA.
-- Safe to run multiple times.

START TRANSACTION;

-- Ensure PERENCANA role exists.
INSERT INTO roles (code, name)
SELECT 'PERENCANA', 'Perencana Unit'
WHERE NOT EXISTS (
  SELECT 1 FROM roles WHERE code = 'PERENCANA'
);

-- Move user_roles bindings from PLANNER to PERENCANA.
INSERT IGNORE INTO user_roles (user_id, role_id)
SELECT ur.user_id, target.id
FROM user_roles ur
JOIN roles legacy ON legacy.id = ur.role_id AND legacy.code = 'PLANNER'
JOIN roles target ON target.code = 'PERENCANA';

-- Move approval_flows bindings from PLANNER to PERENCANA.
UPDATE approval_flows af
JOIN roles legacy ON legacy.id = af.role_id AND legacy.code = 'PLANNER'
JOIN roles target ON target.code = 'PERENCANA'
SET af.role_id = target.id;

-- Remove old bindings.
DELETE ur
FROM user_roles ur
JOIN roles legacy ON legacy.id = ur.role_id
WHERE legacy.code = 'PLANNER';

-- Normalize users.role enum values if any legacy data still exists.
UPDATE users
SET role = 'PERENCANA'
WHERE role = 'PLANNER';

-- Delete PLANNER role record if still present.
DELETE FROM roles
WHERE code = 'PLANNER';

COMMIT;
