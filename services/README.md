# Services Monorepo Scaffold

Folder ini berisi baseline microservice untuk e-Planning:

- `api-gateway`
- `planning-service`
- `renja-service`
- `performance-service`

Masing-masing service menggunakan layering:

- `internal/controller`
- `internal/service`
- `internal/repository`
- `internal/models`

## Menjalankan per service

Contoh:

```bash
cd services/planning-service
go mod tidy
go run ./cmd/server
```

Port default:

- api-gateway: `:8080`
- planning-service: `:8081`
- renja-service: `:8082`
- performance-service: `:8083`
