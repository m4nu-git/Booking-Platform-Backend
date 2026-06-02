################################################################################
# Airbnb Microservices — Root Makefile
#
# Primary workflow (local dev, services run natively with hot-reload):
#   make dev          → start Redis (Docker), then all 5 services natively
#   make stop         → stop Docker containers
#   make migrate      → run ALL database migrations against local MySQL (3306)
#
# Port layout:
#   Local MySQL  → 3306  (your existing installation, used by native services)
#   Docker MySQL → 3307  (used only by the full Docker stack)
#   Docker Redis → 6379  (used by native services via localhost:6379)
#
# Full Docker workflow (all services containerised):
#   make docker-up    → build images and start everything
#   make docker-down  → stop and remove containers
#   make docker-reset → wipe volumes and rebuild from scratch
################################################################################

SHELL := /bin/bash
.DEFAULT_GOAL := help

# ─── Colours ───────────────────────────────────────────────────────────────────
GREEN  := $(shell printf '\033[0;32m')
YELLOW := $(shell printf '\033[0;33m')
BLUE   := $(shell printf '\033[0;36m')
RED    := $(shell printf '\033[0;31m')
RESET  := $(shell printf '\033[0m')

# ─── Database credentials ──────────────────────────────────────────────────────
# DB_PORT is the LOCAL MySQL port (used by 'make migrate' and native services).
# Docker MySQL is exposed on 3307 to avoid conflict — it is used only by 'make docker-up'.
DB_USER          := root
DB_PASSWORD      := root1234
DB_HOST           := 127.0.0.1
DB_PORT           := 3306
DOCKER_MYSQL_PORT := 3307

# ─── Help ──────────────────────────────────────────────────────────────────────
.PHONY: help
help: ## Show this help
	@echo ""
	@echo "  $(BLUE)Airbnb Microservices$(RESET)"
	@echo "  $(BLUE)═══════════════════════════════════════$(RESET)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-22s$(RESET) %s\n", $$1, $$2}'
	@echo ""

# ─── Local Dev (native hot-reload) ────────────────────────────────────────────
.PHONY: dev
dev: infra-up ## Start infra + all 5 services with live reload
	@echo "$(GREEN)▶ Starting all 5 services...$(RESET)"
	npm run dev:services

.PHONY: start
start: dev ## Alias for dev

.PHONY: stop
stop: infra-down ## Stop infrastructure containers

# ─── Infrastructure ───────────────────────────────────────────────────────────
.PHONY: infra-up
infra-up: ## Start Docker Redis (and MySQL on :3307 for full-Docker use)
	@echo "$(BLUE)▶ Starting infrastructure containers...$(RESET)"
	docker compose -f docker-compose.infra.yml up -d
	@echo "$(YELLOW)  Waiting for Redis to be ready...$(RESET)"
	@until docker compose -f docker-compose.infra.yml exec -T redis \
		redis-cli ping 2>/dev/null | grep -q PONG; do \
		printf '.'; sleep 1; \
	done
	@echo ""
	@echo "$(GREEN)✔ Redis is ready on localhost:6379$(RESET)"
	@echo "$(GREEN)✔ Docker MySQL available on localhost:$(DOCKER_MYSQL_PORT) (full-Docker use only)$(RESET)"
	@echo "$(YELLOW)  Native services will use your local MySQL on localhost:$(DB_PORT)$(RESET)"

.PHONY: infra-down
infra-down: ## Stop infrastructure containers
	docker compose -f docker-compose.infra.yml down
	@echo "$(YELLOW)▶ Infrastructure stopped$(RESET)"

.PHONY: infra-logs
infra-logs: ## Follow MySQL + Redis logs
	docker compose -f docker-compose.infra.yml logs -f

.PHONY: infra-reset
infra-reset: ## Wipe MySQL + Redis data and restart (WARNING: deletes all local data)
	@echo "$(RED)⚠ This will delete all local database data. Press Ctrl+C to cancel.$(RESET)"
	@sleep 3
	docker compose -f docker-compose.infra.yml down -v
	docker compose -f docker-compose.infra.yml up -d
	@echo "$(GREEN)✔ Infrastructure reset complete$(RESET)"

# ─── Migrations (run once after fresh install or infra-reset) ─────────────────
.PHONY: migrate
migrate: ## Run all database migrations against the local MySQL
	@echo "$(BLUE)▶ Running AuthService (auth_dev) migrations...$(RESET)"
	cd AuthServiceInGo && goose -dir db/migrations mysql \
		"$(DB_USER):$(DB_PASSWORD)@tcp($(DB_HOST):$(DB_PORT))/auth_dev" up
	@echo "$(BLUE)▶ Running HotelService (airbnb_dev) migrations...$(RESET)"
	npm --prefix HotelService run migrate
	@echo "$(BLUE)▶ Running BookingService (airbnb_booking_dev) migrations...$(RESET)"
	cd BookingService && npx prisma migrate deploy --schema src/prisma/schema.prisma
	@echo "$(BLUE)▶ Running ReviewService (review_dev) migrations...$(RESET)"
	cd ReviewService && goose -dir db/migrations mysql \
		"$(DB_USER):$(DB_PASSWORD)@tcp($(DB_HOST):$(DB_PORT))/review_dev" up
	@echo "$(GREEN)✔ All migrations complete$(RESET)"

.PHONY: migrate-status
migrate-status: ## Show migration status for Auth and Review services
	@echo "$(BLUE)AuthService:$(RESET)"
	cd AuthServiceInGo && goose -dir db/migrations mysql \
		"$(DB_USER):$(DB_PASSWORD)@tcp($(DB_HOST):$(DB_PORT))/auth_dev" status
	@echo "$(BLUE)ReviewService:$(RESET)"
	cd ReviewService && goose -dir db/migrations mysql \
		"$(DB_USER):$(DB_PASSWORD)@tcp($(DB_HOST):$(DB_PORT))/review_dev" status

# ─── Install Dependencies ─────────────────────────────────────────────────────
.PHONY: install
install: ## Install all Node.js dependencies across services
	@echo "$(BLUE)▶ Installing root dependencies...$(RESET)"
	npm install
	@echo "$(BLUE)▶ Installing HotelService dependencies...$(RESET)"
	npm --prefix HotelService install
	@echo "$(BLUE)▶ Installing BookingService dependencies...$(RESET)"
	npm --prefix BookingService install
	@echo "$(BLUE)▶ Installing NotificationService dependencies...$(RESET)"
	npm --prefix NotificationService install
	@echo "$(GREEN)✔ All Node.js dependencies installed$(RESET)"

# ─── Full Docker Stack ────────────────────────────────────────────────────────
.PHONY: docker-up
docker-up: ## Build and start the full stack in Docker (all 5 services + infra)
	@echo "$(BLUE)▶ Building and starting full Docker stack...$(RESET)"
	docker compose up --build -d
	@echo "$(GREEN)✔ All containers started. Run 'make docker-logs' to follow logs.$(RESET)"

.PHONY: docker-down
docker-down: ## Stop and remove Docker containers (keeps volumes)
	docker compose down
	@echo "$(YELLOW)▶ Docker stack stopped$(RESET)"

.PHONY: docker-logs
docker-logs: ## Follow logs for all Docker services
	docker compose logs -f

.PHONY: docker-ps
docker-ps: ## Show status of all Docker containers
	docker compose ps

.PHONY: docker-reset
docker-reset: ## Wipe all Docker volumes and rebuild from scratch
	@echo "$(RED)⚠ This will delete ALL Docker data. Press Ctrl+C to cancel.$(RESET)"
	@sleep 3
	docker compose down -v
	docker compose up --build -d
	@echo "$(GREEN)✔ Full Docker reset complete$(RESET)"

# ─── Convenience ─────────────────────────────────────────────────────────────
.PHONY: health
health: ## Hit the gateway health endpoint
	@curl -s http://localhost:8080/health | python3 -m json.tool 2>/dev/null || \
		curl -s http://localhost:8080/health

.PHONY: setup
setup: install infra-up migrate ## First-time setup: install deps, start infra, run all migrations
	@echo ""
	@echo "$(GREEN)╔══════════════════════════════════════╗$(RESET)"
	@echo "$(GREEN)║  Setup complete! Run 'make dev' next  ║$(RESET)"
	@echo "$(GREEN)╚══════════════════════════════════════╝$(RESET)"
