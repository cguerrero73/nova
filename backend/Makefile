.PHONY: help dev build run test lint clean migrate-up migrate-down seed

# Variables
BINARY_NAME=nova-server
MAIN_PATH=./cmd/server
DB_HOST?=localhost
DB_PORT?=5432
DB_USER?=postgres
DB_PASSWORD?=postgres
DB_DATABASE?=nova
DB_SCHEMA?=public

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

dev: ## Run development server
	@DB_HOST=$(DB_HOST) DB_PORT=$(DB_PORT) DB_USER=$(DB_USER) DB_PASSWORD=$(DB_PASSWORD) DB_DATABASE=$(DB_DATABASE) DB_SCHEMA=$(DB_SCHEMA) go run $(MAIN_PATH)/main.go

build: ## Build binary
	go build -o $(BINARY_NAME) $(MAIN_PATH)/main.go

run: ## Run compiled binary
	./$(BINARY_NAME)

test: ## Run tests
	go test -v ./...

lint: ## Run linter
	golangci-lint run

clean: ## Clean build artifacts
	rm -f $(BINARY_NAME)

# Database migrations
migrate-up: ## Run pending migrations
	@echo "Running migrations..."
	@read -p "Migration type (global/tenant): " TYPE; \
	if [ "$$TYPE" = "global" ]; then \
		psql "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_DATABASE)?search_path=$(DB_SCHEMA)" -f migrations/global/001_init_global.up.sql; \
	elif [ "$$TYPE" = "tenant" ]; then \
		read -p "Tenant code: " TENANT; \
		psql "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_DATABASE)" -c "CREATE SCHEMA IF NOT EXISTS tenant_$$TENANT"; \
		psql "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_DATABASE)?search_path=tenant_$$TENANT" -f migrations/tenant/001_init_tenant.up.sql; \
	else \
		echo "Invalid type"; \
	fi

migrate-down: ## Rollback last migration
	@echo "Rolling back migrations..."
	@read -p "Migration type (global/tenant): " TYPE; \
	if [ "$$TYPE" = "global" ]; then \
		psql "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_DATABASE)?search_path=$(DB_SCHEMA)" -f migrations/global/001_init_global.down.sql; \
	elif [ "$$TYPE" = "tenant" ]; then \
		read -p "Tenant code: " TENANT; \
		psql "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_DATABASE)?search_path=tenant_$$TENANT" -f migrations/tenant/001_init_tenant.down.sql; \
		psql "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_DATABASE)" -c "DROP SCHEMA IF EXISTS tenant_$$TENANT CASCADE"; \
	else \
		echo "Invalid type"; \
	fi

# Seed data
seed: ## Run seeds
	@echo "Running seeds..."
	@read -p "Tenant code: " TENANT; \
	psql "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_DATABASE)?search_path=tenant_$$TENANT" -f seeds/001_syscodes.sql; \
	psql "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_DATABASE)?search_path=tenant_$$TENANT" -f seeds/002_default_roles.sql; \
	echo "Seeds applied successfully"

# Create a new tenant
create-tenant: ## Create new tenant (usage: make create-tenant TENANT=acme)
ifndef TENANT
	@echo "TENANT is required. Usage: make create-tenant TENANT=acme"
else
	@echo "Creating tenant $(TENANT)..."
	@psql "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_DATABASE)" -c "INSERT INTO eamtenants (ten_code, ten_name) VALUES ('$(TENANT)', '$(TENANT)') ON CONFLICT (ten_code) DO NOTHING"
	@psql "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_DATABASE)" -c "CREATE SCHEMA IF NOT EXISTS tenant_$(TENANT)"
	@echo "Tenant schema created. Run 'make migrate-up TYPE=tenant TENANT=$(TENANT)' to apply schema"
endif

# Generate ent schemas (requires ent installed)
ent-generate: ## Generate ent schemas from database
	@go run entgo.io/cmd/ent generate ./schema

# Development helpers
db-reset: ## Reset database (DANGER: destroys all data)
	@echo "This will delete ALL data. Are you sure? (y/n)"
	@read ans; \
	if [ "$$ans" = "y" ]; then \
		psql "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_DATABASE)" -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"; \
		make migrate-up TYPE=global; \
	else \
		echo "Cancelled"; \
	fi