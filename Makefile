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

.PHONY: db db-down db-logs db-shell \
        frontend-install frontend-build build-backend build-frontend build-all \
        docker run logs app-shell go-test \
        clean docker-clean help

# --- Database Commands ---

db:
	@echo "Starting dev database..."
	$(DOCKER_COMPOSE) -f $(COMPOSE_FILE) up -d

db-down:
	@echo "Stopping dev database..."
	$(DOCKER_COMPOSE) -f $(COMPOSE_FILE) down -v

db-logs:
	$(DOCKER_COMPOSE) -f $(COMPOSE_FILE) logs -f

db-shell:
	$(DOCKER_COMPOSE) -f $(COMPOSE_FILE) exec postgres psql -U $${DB_USER:-postgres} -d $${DB_NAME:-packup}

# --- Build Commands ---

frontend-install:
	cd frontend && npm install

frontend-build:
	cd frontend && npm run build

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

run:
	@echo "Starting $(CONTAINER_NAME) container..."
	-docker rm -f $(CONTAINER_NAME) 2>/dev/null
	docker run -d \
		--name $(CONTAINER_NAME) \
		-p 8080:8080 \
		--add-host=host.docker.internal:host-gateway \
		--env-file docker/.env.dev \
		-e DB_HOST=host.docker.internal \
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
	@echo "Database Commands:"
	@echo "  db             - Start dev database"
	@echo "  db-down        - Stop dev database and remove volumes"
	@echo "  db-logs        - Follow database logs"
	@echo "  db-shell       - Open PSQL shell in database"
	@echo ""
	@echo "Build Commands:"
	@echo "  docker         - Full clean build and docker image creation"
	@echo "  build-all      - Build backend and frontend locally"
	@echo "  frontend-install - Install frontend dependencies"
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
