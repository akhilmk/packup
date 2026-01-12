COMPOSE_FILE ?= compose.postgres-dev.yml
DOCKER_COMPOSE ?= docker compose
IMAGE_NAME := itinera

.PHONY: db-up db-down db-logs db-shell frontend-install frontend-build \
        build-backend build-frontend build-all docker-build docker-run docker-stop \
        docker-logs docker-clean release clean help test

# Test commands
test:
	@echo "Running backend tests..."
	cd backend && go test -v ./...

test-coverage:
	@echo "Running backend tests with coverage..."
	cd backend && go test -coverprofile=coverage.out ./...
	cd backend && go tool cover -html=coverage.out -o coverage.html
	@echo "✓ Coverage report: backend/coverage.html"

# Database commands
db-up:
	$(DOCKER_COMPOSE) -f $(COMPOSE_FILE) up -d

db-down:
	$(DOCKER_COMPOSE) -f $(COMPOSE_FILE) down -v

db-logs:
	$(DOCKER_COMPOSE) -f $(COMPOSE_FILE) logs -f

db-shell:
	# open an interactive psql shell in the postgres service
	$(DOCKER_COMPOSE) -f $(COMPOSE_FILE) exec postgres psql -U $${DB_USER:-postgres} -d $${DB_NAME:-itinera}

# Frontend commands
frontend-install:
	cd frontend && npm install

frontend-build:
	cd frontend && npm run build

build-frontend: frontend-build
	@echo "Copying frontend to bin/frontend/dist..."
	mkdir -p bin/frontend
	rm -rf bin/frontend/dist
	cp -r frontend/dist bin/frontend/
	@echo "✓ Frontend copied to bin/frontend/dist"

# Backend commands
build-backend:
	@echo "Building Go backend..."
	cd backend && CGO_ENABLED=0 GOOS=linux go build -o ../bin/itinera ./cmd/server
	@echo "Copying migrations to bin/migrations..."
	mkdir -p bin/migrations
	cp -r backend/migrations/* bin/migrations/
	@echo "✓ Backend binary created at bin/itinera"
	@echo "✓ Migrations copied to bin/migrations"

# Build everything (backend + frontend)
build-all: build-backend build-frontend
	@echo "✓ Build complete! Binary and frontend are in bin/"

# Build Docker image using pre-built binaries
docker-build: build-all
	@echo "Building Docker image $(IMAGE_NAME):latest..."
	docker build -t $(IMAGE_NAME):latest .
	@echo "✓ Docker image built: $(IMAGE_NAME):latest"

# Run Docker container
docker-run:
	@echo "Starting container $(IMAGE_NAME)..."
	-docker rm -f itinera 2>/dev/null
	docker run -d \
		--name itinera \
		-p 8080:8080 \
		--add-host=host.docker.internal:host-gateway \
		--env-file .env.dev \
		-e DB_HOST=host.docker.internal \
		$(IMAGE_NAME):latest
	@echo "✓ Container started: itinera"
	@echo "  Access at: http://localhost:8080"
	@echo "  View logs: make docker-logs"

# Stop and remove Docker container
docker-stop:
	@echo "Stopping container..."
	-docker stop itinera
	-docker rm itinera
	@echo "✓ Container stopped and removed"

# Clean Docker images
docker-clean:
	@echo "Removing Docker images..."
	-docker rmi $(IMAGE_NAME):latest
	@echo "✓ Docker images removed"

# View Docker container logs
docker-logs:
	docker logs -f itinera

# Complete release: build everything and create Docker image
release: build-all docker-build
	@echo "✓ Release complete!"
	@echo "  - Binary: bin/itinera"
	@echo "  - Frontend: bin/frontend/dist"
	@echo "  - Docker: $(IMAGE_NAME):latest"

# Clean build artifacts
clean:
	rm -rf bin/
	rm -rf frontend/dist
	@echo "✓ Cleaned build artifacts"

# Help command
help:
	@echo "Available commands:"
	@echo "  make db-up           - Start PostgreSQL database"
	@echo "  make db-down         - Stop PostgreSQL database"
	@echo "  make db-logs         - View database logs"
	@echo "  make db-shell        - Open psql shell"
	@echo ""
	@echo "  make test             - Run backend tests"
	@echo "  make test-coverage    - Run tests with coverage report"
	@echo ""
	@echo "  make frontend-install - Install npm dependencies"
	@echo "  make frontend-build   - Build frontend"
	@echo "  make build-backend    - Build Go binary"
	@echo "  make build-frontend   - Build frontend and copy to bin/"
	@echo "  make build-all        - Build backend + frontend"
	@echo ""
	@echo "  make docker-build     - Build Docker image"
	@echo "  make docker-run       - Run Docker container"
	@echo "  make docker-stop      - Stop and remove container"
	@echo "  make docker-logs      - View container logs"
	@echo "  make docker-clean     - Remove Docker images"
	@echo "  make release          - Full release: build + docker"
	@echo ""
	@echo "  make clean            - Remove build artifacts"
	@echo "  make help             - Show this help"
