MYSQL_DSN ?= root:rootpassword@tcp(localhost:3307)/e-plan-ai?parseTime=true
MYSQL_CMD ?= mysql
MYSQL_USER ?= root
MYSQL_DB ?= e-plan-ai
MYSQL_PWD ?= rootpassword

.PHONY: gorm-migrate gorm-check migrate seed

gorm-migrate:
	MYSQL_DSN='$(MYSQL_DSN)' go run ./cmd/gorm-migrate

gorm-check:
	MYSQL_DSN='$(MYSQL_DSN)' go run ./cmd/gorm-check

migrate:
	go run ./cmd/migrate

seed:
	$(MYSQL_CMD) -u $(MYSQL_USER) $(MYSQL_DB) < docs/performance-planning-seed.sql

# Docker Targets
.PHONY: docker-build docker-up docker-down docker-logs

docker-build:
	docker compose build

docker-up:
	docker compose up -d

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f app
