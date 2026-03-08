MYSQL_DSN ?= root@tcp(localhost:3306)/e-plan-ai?parseTime=true
MYSQL_CMD ?= mysql
MYSQL_USER ?= root
MYSQL_DB ?= e-plan-ai

.PHONY: gorm-migrate gorm-check migrate seed

gorm-migrate:
	MYSQL_DSN='$(MYSQL_DSN)' go run ./cmd/gorm-migrate

gorm-check:
	MYSQL_DSN='$(MYSQL_DSN)' go run ./cmd/gorm-check

migrate:
	MYSQL_DSN='$(MYSQL_DSN)' go run ./cmd/migrate

seed:
	$(MYSQL_CMD) -u $(MYSQL_USER) $(MYSQL_DB) < docs/performance-planning-seed.sql
