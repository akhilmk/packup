# makefile for development environment 

# Load environment variables from docker/.env.dev if it exists
ifneq (,$(wildcard docker/.env.dev))
    include docker/.env.dev
    export
endif

# Configuration
COMPOSE_FILE ?= docker/docker-compose.dev.yml
DOCKER_COMPOSE ?= docker compose
IMAGE_NAME := packup
CONTAINER_NAME := packup
NODE_IMAGE := node:24-alpine
USER_ID := $(shell id -u)
GROUP_ID := $(shell id -g)
COMPOSE_ENV := USER_ID=$(USER_ID) GROUP_ID=$(GROUP_ID)
DOCKER_EXEC_NODE := docker exec -i packup-node-dev

.PHONY: dev dev-down dev-logs dev-shell \
        frontend-install frontend-audit-fix frontend-build build-backend build-frontend build-all \
        docker run logs app-shell go-test \
        clean docker-clean help

# --- Database Commands ---

dev:
	@echo "Starting dev environment (db + node builder)..."
	$(COMPOSE_ENV) $(DOCKER_COMPOSE) -f $(COMPOSE_FILE) --env-file docker/.env.dev up -d

dev-down:
	@echo "Stopping dev environment..."
	$(COMPOSE_ENV) $(DOCKER_COMPOSE) -f $(COMPOSE_FILE) --env-file docker/.env.dev down -v

dev-logs:
	$(COMPOSE_ENV) $(DOCKER_COMPOSE) -f $(COMPOSE_FILE) --env-file docker/.env.dev logs -f

dev-shell:
	$(COMPOSE_ENV) $(DOCKER_COMPOSE) -f $(COMPOSE_FILE) --env-file docker/.env.dev exec postgres psql -U $(DB_USER) -d $(DB_NAME)

# --- Build Commands ---

frontend-install:
	@if [ $$(docker ps -q -f name=packup-node-dev) ]; then \
		$(DOCKER_EXEC_NODE) npm install; \
	else \
		docker run --rm -v $(PWD)/frontend:/app -w /app --user $(USER_ID):$(GROUP_ID) $(NODE_IMAGE) npm install; \
	fi

frontend-audit-fix:
	@if [ $$(docker ps -q -f name=packup-node-dev) ]; then \
		$(DOCKER_EXEC_NODE) npm audit fix; \
	else \
		docker run --rm -v $(PWD)/frontend:/app -w /app --user $(USER_ID):$(GROUP_ID) $(NODE_IMAGE) npm audit fix; \
	fi

frontend-build:
	@if [ $$(docker ps -q -f name=packup-node-dev) ]; then \
		$(DOCKER_EXEC_NODE) npm run build; \
	else \
		docker run --rm -v $(PWD)/frontend:/app -w /app --user $(USER_ID):$(GROUP_ID) $(NODE_IMAGE) npm run build; \
	fi

build-frontend: frontend-build
	@echo "Preparing frontend artifacts..."
	mkdir -p bin/frontend
	rm -rf bin/frontend/dist
	cp -r frontend/dist bin/frontend/

build-backend:
	@echo "Building Go backend..."
	cd backend && CGO_ENABLED=0 GOOS=linux go build -o ../bin/packup ./cmd/server
	@echo "Copying migrations..."
	mkdir -p bin/migrations
	cp -r backend/migrations/* bin/migrations/

build-all: build-backend build-frontend
	@echo "✓ All artifacts built in bin/"

docker: docker-clean clean build-all
	@echo "Building Docker image $(IMAGE_NAME):latest..."
	docker build -t $(IMAGE_NAME):latest -f docker/Dockerfile .
	@echo "✓ Docker image built"

# --- Runtime Commands ---

# "--network packup-dev" - use network of dev docker compose.
# "-e DB_HOST=packup-dev-db" - use db host name from dev docker compose.
run:
	@echo "Starting $(CONTAINER_NAME) container..."
	-docker rm -f $(CONTAINER_NAME) 2>/dev/null
	docker run -d \
		--name $(CONTAINER_NAME) \
		-p 8080:8080 \
		--network packup-dev \
		-e DB_USER=$(DB_USER) \
		-e DB_PASS=$(DB_PASS) \
		-e DB_NAME=$(DB_NAME) \
		-e DB_HOST=packup-dev-db \
		-e DB_PORT=5432 \
		-e SSLMODE=$(SSLMODE) \
		-e PORT=8080 \
		-e ADMIN_EMAILS=$(ADMIN_EMAILS) \
		-e GOOGLE_CLIENT_ID=$(GOOGLE_CLIENT_ID) \
		-e GOOGLE_CLIENT_SECRET=$(GOOGLE_CLIENT_SECRET) \
		-e GOOGLE_REDIRECT_URI=$(GOOGLE_REDIRECT_URI) \
		$(IMAGE_NAME):latest
	@echo "✓ App running at http://localhost:8080"

logs:
	docker logs -f $(CONTAINER_NAME)

app-shell:
	docker exec -it $(CONTAINER_NAME) sh

# --- Testing Commands ---

go-test:
	@echo "Running backend tests..."
	cd backend && go test -v ./...

# --- Utility Commands ---

clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -rf frontend/dist

docker-clean:
	@echo "Stopping and removing Docker images/containers..."
	-docker stop $(CONTAINER_NAME) 2>/dev/null
	-docker rm $(CONTAINER_NAME) 2>/dev/null
	-docker rmi $(IMAGE_NAME):latest 2>/dev/null

help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Development Commands:"
	@echo "  dev            - Start dev environment (db + node builder)"
	@echo "  dev-down       - Stop dev environment and remove volumes"
	@echo "  dev-logs       - Follow dev environment logs"
	@echo "  dev-shell      - Open PSQL shell in database"
	@echo ""
	@echo "Build Commands:"
	@echo "  docker         - Full clean build and docker image creation"
	@echo "  build-all      - Build backend and frontend locally"
	@echo "  frontend-install - Install frontend dependencies"
	@echo "  frontend-audit-fix - Fix frontend dependency vulnerabilities"
	@echo ""
	@echo "Runtime Commands:"
	@echo "  run            - Run the app container locally"
	@echo "  logs           - Follow app container logs"
	@echo "  app-shell      - Open shell inside the app container"
	@echo ""
	@echo "Testing Commands:"
	@echo "  go-test        - Run Go backend tests"
	@echo ""
	@echo "Utility Commands:"
	@echo "  clean          - Remove local build artifacts"
	@echo "  docker-clean   - Remove app container and image"
