# Clean Project Structure (Gin Performance Planning)

Tujuan struktur ini adalah menjaga separation of concerns, mudah scaling modul, dan mudah diuji.

```text
cmd/
  api/                      # Main HTTP service (Gin)
  migrate/                  # CLI migration runner

configs/                    # Config templates (yaml/json/toml)

docs/                       # Arsitektur, API contract, operational docs

internal/
  app/                      # App wiring and composition root
  bootstrap/                # Router bootstrap and registration
  config/                   # Environment loader
  modules/                  # Business modules (bounded contexts)
    planning/
      delivery/http/
      usecase/
      domain/
      repository/
    renja/
      delivery/http/
      usecase/
      domain/
      repository/
    performance/
      delivery/http/
      usecase/
      domain/
      repository/
    crud/
      delivery/http/
      usecase/
      domain/
      repository/
    ui/
      delivery/http/
      usecase/
  shared/
    database/               # DB connection and tx helper
    middleware/             # auth, recovery, audit, tracing
    response/               # standard API response

migrations/                 # SQL migration files

pkg/                        # Reusable packages outside internal
  logger/
  validator/

scripts/                    # Dev scripts (seed, backup, reset)

tests/
  integration/              # Integration/E2E tests

web/
  assets/
    css/
    js/
  templates/
```

## Rules of Thumb

- `delivery/http` only handles transport concerns (Gin context, binding, status code).
- `usecase` holds business workflow and transactions.
- `domain` contains entities and business invariants.
- `repository` encapsulates database queries.
- `shared` contains cross-cutting utilities used by many modules.
- `pkg` only for generic reusable code that does not depend on internal business details.
