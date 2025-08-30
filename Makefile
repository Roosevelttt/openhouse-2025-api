SHELL := bash

# Load .env if present
ifneq (,$(wildcard .env))
  include .env
  export
endif

# Defaults (can be overridden by .env)
DB_HOST ?= 127.0.0.1
DB_PORT ?= 3306
DB_DATABASE ?= openhouse_2025
DB_USERNAME ?= root
DB_PASSWORD ?=

GO ?= go
GOOSE ?= goose
MIGRATIONS_DIR ?= db/migrations

DB_DSN := $(DB_USERNAME):$(DB_PASSWORD)@tcp($(DB_HOST):$(DB_PORT))/$(DB_DATABASE)?parseTime=true&charset=utf8mb4&loc=Local

.PHONY: help tools goose-status migrate-up migrate-down migrate-redo migrate-create seed-ukm seed-division run-server tidy

help:
	@echo "Targets:"
	@echo "  tools            Install goose CLI"
	@echo "  goose-status     Show migration status"
	@echo "  migrate-up       Apply all pending migrations"
	@echo "  migrate-down     Rollback the most recent migration"
	@echo "  migrate-redo     Rollback then re-apply the most recent migration"
	@echo "  migrate-create name=<snake_case>  Create a new SQL migration"
	@echo "  seed-ukm         Run the UKM seeder"
	@echo "  seed-division    Run the Division seeder"
	@echo "  run-server       Run the API server"
	@echo "  tidy             go mod tidy"

tools:
	$(GO) install github.com/pressly/goose/v3/cmd/goose@latest

goose-status:
	$(GOOSE) -dir $(MIGRATIONS_DIR) mysql "$(DB_DSN)" status

migrate-up:
	$(GOOSE) -dir $(MIGRATIONS_DIR) mysql "$(DB_DSN)" up

migrate-down:
	$(GOOSE) -dir $(MIGRATIONS_DIR) mysql "$(DB_DSN)" down

migrate-redo:
	$(GOOSE) -dir $(MIGRATIONS_DIR) mysql "$(DB_DSN)" redo

# Usage: make migrate-create name=create_new_table
migrate-create:
	@if [ -z "$(name)" ]; then echo "Usage: make migrate-create name=create_new_table"; exit 1; fi
	$(GOOSE) -dir $(MIGRATIONS_DIR) create $(name) sql

seed-ukm:
	$(GO) run ./cmd/ukmseeder

seed-division:
	$(GO) run ./cmd/divisionseeder

run-server:
	$(GO) run ./cmd/server

tidy:
	$(GO) mod tidy

