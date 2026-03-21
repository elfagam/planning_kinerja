-- Reset password untuk akun demo agar mudah login pada environment development.
-- Password default (plaintext): Admin123!
-- WARNING: jangan gunakan password ini di production.

UPDATE users
SET password_hash = '$2a$10$A13.hD0DxyCqvXHfUZ3bte.itOLmOYqHeXQTn5Rng7C./11A2YYaK'
WHERE email IN (
  'superadmin@rsudcontoh.go.id',
  'planner.med@rsudcontoh.go.id',
  'reviewer.nur@rsudcontoh.go.id',
  'approver@rsudcontoh.go.id',
  'verifier@rsudcontoh.go.id',
  'admin.planning@rsud.local',
  'verifier@rsud.local'
);
