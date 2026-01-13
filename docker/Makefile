# Makefile for production

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

.PHONY: docker-release prod-up prod-down build-all docker-build

# --- Pipeline / Release Commands (Run from project root) ---

# Complete release: build everything and create Docker image, then push to Hub
docker-release: build-all docker-build
	@echo "Tagging and pushing image to Docker Hub..."
	docker tag $(IMAGE_NAME):latest $(DOCKER_HUB_USER)/$(IMAGE_NAME):latest
	docker push $(DOCKER_HUB_USER)/$(IMAGE_NAME):latest
	@echo "✓ Release $(DOCKER_HUB_USER)/$(IMAGE_NAME):latest pushed successfully"

# Delegate source building to the root Makefile to maintain context
build-all:
	$(MAKE) build-all

docker-build:
	$(MAKE) docker-build

# --- Production Deployment Commands (Run from project root) ---

prod-up:
	@echo "Starting production stack..."
	$(DOCKER_COMPOSE) -f docker/docker-compose.yml --env-file docker/.env.prod up -d
	@echo "✓ Production stack is up"

prod-down:
	@echo "Stopping production stack..."
	$(DOCKER_COMPOSE) -f docker/docker-compose.yml --env-file docker/.env.prod down
	@echo "✓ Production stack is down"