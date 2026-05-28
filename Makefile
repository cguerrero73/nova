.PHONY: help dev devFrontend devBackend build test lint lintFix
.PHONY: dbMigrate dbSeed clean

help:
	@echo "Available commands:"
	@echo ""
	@echo "  make dev         - Start all services (frontend + backend)"
	@echo "  make devFrontend - Start frontend only"
	@echo "  make devBackend  - Start backend only"
	@echo "  make build       - Build all projects"
	@echo "  make test        - Run all tests"
	@echo "  make lint        - Lint all projects"
	@echo "  make lintFix     - Auto-fix linting issues"
	@echo "  make dbMigrate   - Run database migrations (golang-migrate)"
	@echo "  make dbSeed      - Seed database"
	@echo "  make clean       - Remove build artifacts"

dev:
	@echo "Starting Nova development environment..."
	@echo "Frontend: http://localhost:4200"
	@echo "Backend:  http://localhost:4000"
	@cd backend && make dev &
	@cd frontend && pnpm start &

devFrontend:
	cd frontend && pnpm start

devBackend:
	cd backend && make dev

build:
	cd backend && make build
	cd frontend && pnpm run build

test:
	cd backend && make test
	cd frontend && pnpm test

lint:
	cd backend && make lint
	cd frontend && pnpm run lint

lintFix:
	cd backend && golangci-lint run --fix
	cd frontend && pnpm run lint --fix

dbMigrate:
	cd backend && make migrate-up

dbSeed:
	cd backend && make seed

clean:
	cd frontend && pnpm run clean || true
	cd backend && make clean || true