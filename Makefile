# Variables
API_CMD := ./cmd/api
WORKER_CMD := ./cmd/worker
TMP_DIR := ./tmp
DOCKER_COMPOSE_FILE := docker/docker-compose.dev.yaml
ENV_FILE := .env.local

# Default target
.DEFAULT_GOAL := help

# ----- Local Dev ----- #
.PHONY: dev-api dev-worker
dev-api: ## Run API with Air
	air -c .air.toml \
		-build.cmd "go build -o $(TMP_DIR)/api $(API_CMD)" \
		-build.bin "$(TMP_DIR)/api"

dev-worker: ## Run worker locally
	air -c .air.toml \
		-build.cmd "go build -o $(TMP_DIR)/worker $(WORKER_CMD)" \
		-build.bin "$(TMP_DIR)/worker"

# ----- AutoGen code ----- #
.PHONY: sqlc genqlient openapi migrate
sqlc: ## Generate SQLC code
	@command -v sqlc >/dev/null || (echo "sqlc not installed" && exit 1)
	sqlc generate

genqlient: ## Generate GraphQL client code
	@command -v genqlient >/dev/null || (echo "genqlient not installed" && exit 1)
	@if [ ! -f schema.graphql ]; then \
		npx --yes graphqurl https://graphql.anilist.co --introspect > schema.graphql; \
		if [ $$? -ne 0 ]; then \
			echo "Failed to fetch schema.graphql"; \
			exit 1; \
		fi \
	fi
	genqlient genqlient.yaml

openapi: ## Generate full openapi docs + ts client
	@./scripts/create-openapi.sh

migrate: ## Generate migration files
	@# Check if API or Worker binaries are running
	@API_RUNNING=$$(pgrep -f "tmp/api" >/dev/null 2>&1 && echo "yes" || echo "no"); \
	WORKER_RUNNING=$$(pgrep -f "tmp/worker" >/dev/null 2>&1 && echo "yes" || echo "no"); \
	if [ "$$API_RUNNING" = "yes" ] || [ "$$WORKER_RUNNING" = "yes" ]; then \
		echo "❌ Error: API or Worker binary is currently running."; \
		echo ""; \
		echo "Please close the running binaries (stop air/dev processes) before creating migration files."; \
		echo "This prevents air from auto-running blank migrations."; \
		echo ""; \
		if [ "$$API_RUNNING" = "yes" ]; then \
			echo "  - API is running (found process matching 'tmp/api')"; \
		fi; \
		if [ "$$WORKER_RUNNING" = "yes" ]; then \
			echo "  - Worker is running (found process matching 'tmp/worker')"; \
		fi; \
		echo ""; \
		echo "Run: pkill -f 'tmp/api' or pkill -f 'tmp/worker' to stop them."; \
		exit 1; \
	fi
	@./scripts/migrate.sh

# ----- Docker Compose (Dev) ----- #
.PHONY: dev-docker-up dev-docker-down dev-docker-logs
dev-docker-up: ## Start dev containers
	@test -f $(ENV_FILE) || (echo "$(ENV_FILE) not found" && exit 1)
	docker compose -p aniways -f $(DOCKER_COMPOSE_FILE) --env-file $(ENV_FILE) up -d

dev-docker-down: ## Stop dev containers
	docker compose -p aniways -f $(DOCKER_COMPOSE_FILE) --env-file $(ENV_FILE) down

dev-docker-logs: ## View logs for all containers
	docker compose -p aniways -f $(DOCKER_COMPOSE_FILE) --env-file $(ENV_FILE) logs -f

# ----- Tmux Commands ----- #
.PHONY: tmux
tmux: ## Start tmux session
	@command -v ntmux >/dev/null 2>&1 && \
		{ echo "ntmux found, applying template..."; ntmux apply ntmux.json; } || \
		{ echo "ntmux not found, falling back to script..."; ./scripts/aniways-tmux.sh; }

# ----- Setup ----- #
.PHONY: setup
setup: ## Run setup script to install dependencies and start services
	@./scripts/setup.sh

# ----- Help Menu ----- #
.PHONY: help
help: ## Show help
	@echo "Available commands:"
	@echo ""
	@awk '/^# ----- .* ----- #/ { \
		gsub(/^# ----- | ----- #$$/, "", $$0); \
		printf "\033[33m%s:\033[0m\n", $$0; \
	} \
	/^[a-zA-Z_-]+:.*?##/ { \
		split($$0, parts, ":.*?## "); \
		printf "  \033[36m%-18s\033[0m %s\n", parts[1], parts[2]; \
	}' $(MAKEFILE_LIST)
