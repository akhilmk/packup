# Makefile for production and full solution deployment


# Configuration
DOCKER_COMPOSE ?= docker compose

.PHONY: prod-up prod-down

# --- docker compose command for production and staging

prod-up:
	@echo "Starting production stack..."
	$(DOCKER_COMPOSE) -f compose.yaml --env-file .env.prod up -d
	@echo "✓ Production stack is up"

prod-down:
	@echo "Stopping production stack..."
	$(DOCKER_COMPOSE) -f compose.yaml --env-file .env.prod down
	@echo "✓ Production stack is down"

dev-up:
	@echo "Starting development stack..."
	$(DOCKER_COMPOSE) -f compose.yaml --env-file .env.dev up -d
	@echo "✓ Development stack is up"

dev-down:
	@echo "Stopping development stack..."
	$(DOCKER_COMPOSE) -f compose.yaml --env-file .env.dev down
	@echo "✓ Development stack is down"