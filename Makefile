.PHONY: help dev devFrontend devBackend build test lint lintFix
.PHONY: dbMigrate dbGenerate dbPush dbStudio dbSeed clean

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
	@echo "  make dbMigrate   - Run Prisma migrations"
	@echo "  make dbGenerate  - Generate Prisma client"
	@echo "  make dbPush      - Push schema to database"
	@echo "  make dbStudio    - Open Prisma Studio"
	@echo "  make dbSeed      - Seed database"
	@echo "  make clean       - Remove build artifacts"

dev:
	@echo "Starting Nova development environment..."
	@echo "Frontend: http://localhost:4200"
	@echo "Backend:  http://localhost:4000"
	@echo "Prisma Studio: http://localhost:5555"
	cd backend && pnpm run dev &
	cd frontend && pnpm start &

devFrontend:
	cd frontend && pnpm start

devBackend:
	cd backend && pnpm run dev

build:
	cd backend && pnpm run build
	cd frontend && pnpm run build

test:
	cd backend && pnpm test
	cd frontend && pnpm test

lint:
	cd backend && pnpm run lint
	cd frontend && pnpm run lint

lintFix:
	cd backend && pnpm run lintFix
	cd frontend && pnpm run lintFix

dbMigrate:
	cd backend && npx prisma migrate dev

dbGenerate:
	cd backend && npx prisma generate

dbPush:
	cd backend && npx prisma db push

dbStudio:
	cd backend && npx prisma studio

dbSeed:
	cd backend && npx prisma db seed

clean:
	cd frontend && pnpm run clean || true
	cd backend && pnpm run clean || true
