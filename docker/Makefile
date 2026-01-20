# Makefile for production and full solution deployment

# Load environment variables from .env.prod
# This handles both running from root or from inside the docker/ folder
ifneq (,$(wildcard docker/.env.prod))
    include docker/.env.prod
    export
else ifneq (,$(wildcard .env.prod))
    include .env.prod
    export
endif

# Configuration
DOCKER_COMPOSE ?= docker compose
IMAGE_NAME := packup
# Fallback if not in env file
DOCKER_HUB_USER ?= akhilmk

.PHONY: prod-up prod-down

# --- docker compose command for production and staging

prod-up:
	@echo "Starting production stack..."
	$(DOCKER_COMPOSE) -f docker-compose.yml --env-file .env.prod up -d
	@echo "✓ Production stack is up"

prod-down:
	@echo "Stopping production stack..."
	$(DOCKER_COMPOSE) -f docker-compose.yml --env-file .env.prod down
	@echo "✓ Production stack is down"

dev-up:
	@echo "Starting development stack..."
	$(DOCKER_COMPOSE) -f docker-compose.yml --env-file .env.dev up -d
	@echo "✓ Development stack is up"

dev-down:
	@echo "Stopping development stack..."
	$(DOCKER_COMPOSE) -f docker-compose.yml --env-file .env.dev down
	@echo "✓ Development stack is down"