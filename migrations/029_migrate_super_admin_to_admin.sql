-- Normalize legacy SUPER_ADMIN role into ADMIN.
-- Safe to run multiple times.

START TRANSACTION;

-- Ensure ADMIN role exists.
INSERT INTO roles (code, name)
SELECT 'ADMIN', 'Administrator'
WHERE NOT EXISTS (
  SELECT 1 FROM roles WHERE code = 'ADMIN'
);

-- Move user_roles bindings from SUPER_ADMIN to ADMIN.
INSERT IGNORE INTO user_roles (user_id, role_id)
SELECT ur.user_id, admin.id
FROM user_roles ur
JOIN roles legacy ON legacy.id = ur.role_id AND legacy.code = 'SUPER_ADMIN'
JOIN roles admin ON admin.code = 'ADMIN';

-- Remove old bindings.
DELETE ur
FROM user_roles ur
JOIN roles legacy ON legacy.id = ur.role_id
WHERE legacy.code = 'SUPER_ADMIN';

-- Normalize users.role enum values if any legacy data still exists.
UPDATE users
SET role = 'ADMIN'
WHERE role = 'SUPER_ADMIN';

-- Delete SUPER_ADMIN role record if still present.
DELETE FROM roles
WHERE code = 'SUPER_ADMIN';

COMMIT;
