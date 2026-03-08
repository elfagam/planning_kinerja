# Microservice-Ready Architecture: Golang Planning System

Dokumen ini menjabarkan target arsitektur microservice untuk sistem e-Planning pemerintah (perencanaan kinerja, Renja, indikator, target/realisasi) dengan fokus pada scalability, auditability, dan governance.

## 1. Prinsip Utama

- Domain-driven service boundaries (bukan split per tabel).
- Database per service (data ownership tegas).
- API contract versioning (`/v1`, `/v2`).
- Event-driven integration untuk proses lintas domain.
- Strong observability (logs, metrics, traces, audit trail).
- Zero-trust service communication (authn/authz antar service).

## 2. Service Decomposition

1. `api-gateway`

- Entry point publik.
- TLS termination, rate limiting, auth token validation, request routing.

2. `iam-service`

- User, role, policy (RBAC/ABAC).
- Token issue/refresh dan permission introspection.

3. `strategic-planning-service`

- Visi, Misi, Tujuan, Sasaran + indikator level strategis.

4. `program-service`

- Program, Kegiatan, Sub Kegiatan + indikator operasional.

5. `renja-service`

- Siklus Renja: draft, submit, approve, reject, lock.
- Ownership terhadap workflow state machine.

6. `performance-service`

- Indikator Kinerja, Target, Realisasi, perhitungan capaian.

7. `audit-service`

- Immutable audit trail untuk semua mutasi penting.

8. `reporting-service`

- Read model agregat, dashboard, export PDF/Excel.

9. `notification-service` (opsional)

- Email/WA/SMS/internal notification untuk approval dan alert deviasi.

## 3. High-Level Topology

```text
Client (Web/Mobile)
   -> API Gateway
      -> IAM Service
      -> Strategic Planning Service
      -> Program Service
      -> Renja Service
      -> Performance Service
      -> Reporting Service

Async Event Bus (NATS/Kafka/RabbitMQ)
      <- domain events from services
      -> Audit Service
      -> Reporting Service
      -> Notification Service
```

## 4. Data Ownership & Storage

- Setiap service punya schema/database sendiri.
- Tidak ada query lintas DB langsung.
- Integrasi data lintas service melalui:
  - synchronous API (query real-time), atau
  - asynchronous event projection (untuk dashboard/reporting).

Contoh:

- `renja-service` tidak join tabel `performance-service` secara langsung.
- `reporting-service` membangun read model dari event `RenjaApproved`, `TargetUpdated`, `RealisasiSubmitted`.

## 5. Communication Pattern

### Sync (HTTP/gRPC)

- Dipakai untuk command/query yang butuh response langsung.
- Contoh: validasi permission sebelum approve Renja.

### Async (Event Bus)

- Dipakai untuk propagasi perubahan state lintas domain.
- Contoh event:
  - `renja.submitted`
  - `renja.approved`
  - `target.updated`
  - `realisasi.recorded`

### Reliability Pattern

- Outbox pattern per service untuk menjamin event publish tidak hilang.
- Retry + dead-letter queue untuk event consumer.
- Idempotent consumer (event key/sequence check).

## 6. Transaction & Consistency

- ACID transaction hanya di dalam satu service boundary.
- Lintas service gunakan eventual consistency.
- Proses multi-step gunakan Saga/Process Manager (orchestration sederhana).

Contoh Saga Approve Renja:

1. `renja-service` ubah status ke `APPROVED` + outbox event.
2. `performance-service` konsumsi event untuk enable target lock policy.
3. `audit-service` catat event immutable.
4. Jika step downstream gagal, lakukan retry; tidak rollback status Renja secara global kecuali ada kompensasi yang dirancang.

## 7. Security Architecture

- OIDC/JWT via `iam-service`.
- Service-to-service auth: mTLS atau signed service token.
- Policy enforcement:
  - coarse-grained di API Gateway.
  - fine-grained di masing-masing service usecase.
- Semua mutasi wajib bawa actor context (`user_id`, `role`, `unit_id`, `request_id`).

## 8. Observability & Operations

- Centralized logs (JSON structured).
- Metrics (Prometheus): latency, error rate, queue lag, throughput.
- Distributed tracing (OpenTelemetry + Jaeger/Tempo).
- Audit dashboard untuk aktivitas approval/revisi.
- SLO awal:
  - API critical p95 < 500ms
  - error rate < 1%

## 9. Deployment Blueprint

- Containerized deployment (Docker).
- Orchestration: Kubernetes (recommended for enterprise scale).
- Minimal environment:
  - `dev`: single cluster, shared managed DB (schema per service)
  - `staging`: mirror prod topology
  - `prod`: HA DB, autoscaling service, isolated network policy
- CI/CD:
  - test -> security scan -> build image -> deploy canary -> progressive rollout

## 10. Recommended Repository Strategy

### Option A: Monorepo (recommended in early stage)

```text
services/
  api-gateway/
  iam-service/
  strategic-planning-service/
  program-service/
  renja-service/
  performance-service/
  audit-service/
  reporting-service/
platform/
  contracts/            # protobuf/openapi/event schema
  helm/                 # k8s charts
  docker/
  scripts/
```

### Option B: Polyrepo

- Dipakai jika team ownership sudah sangat matang.
- Perlu governance kuat untuk schema/version consistency.

## 11. Contract Governance

- API spec mandatory (OpenAPI/protobuf).
- Event schema versioned (`event_name`, `schema_version`).
- Backward compatibility policy untuk semua perubahan contract.
- Consumer-driven contract test pada pipeline.

## 12. Migration Path from Current Modular Monolith

1. Tetapkan service boundary dan contract dulu (tanpa split fisik).
2. Terapkan outbox + event schema di monolith.
3. Ekstrak `audit-service` dan `reporting-service` terlebih dahulu.
4. Ekstrak `performance-service` jika beban kalkulasi/reporting meningkat.
5. Ekstrak `renja-service` setelah workflow stabil.

Pendekatan ini menurunkan risiko sekaligus menjaga continuity layanan publik.
