COMPOSE_FILE ?= compose.postgres-dev.yml
DOCKER_COMPOSE ?= docker compose
VERSION := $(shell cat version 2>/dev/null || echo "0.0.1")
IMAGE_NAME := go-todo
IMAGE_TAG := dev-$(VERSION)

.PHONY: db-up db-down db-logs db-shell frontend-install frontend-build frontend-clean \
        build-backend build-frontend build-all docker-build release clean version help

# Database commands
db-up:
	$(DOCKER_COMPOSE) -f $(COMPOSE_FILE) up -d

db-down:
	$(DOCKER_COMPOSE) -f $(COMPOSE_FILE) down -v

db-logs:
	$(DOCKER_COMPOSE) -f $(COMPOSE_FILE) logs -f

db-shell:
	# open an interactive psql shell in the postgres service
	$(DOCKER_COMPOSE) -f $(COMPOSE_FILE) exec postgres psql -U $${DB_USER:-postgres} -d $${DB_NAME:-todos}

# Frontend commands
frontend-install:
	cd frontend && npm install

frontend-build:
	cd frontend && npm run build

# Backend commands
build-backend:
	@echo "Building Go backend..."
	cd backend && go build -o ../bin/server ./cmd/server
	@echo "✓ Backend binary created at bin/server"

# Build frontend and copy to bin folder
build-frontend: frontend-build
	@echo "Copying frontend to bin/frontend/dist..."
	mkdir -p bin/frontend
	rm -rf bin/frontend/dist
	cp -r frontend/dist bin/frontend/
	@echo "✓ Frontend copied to bin/frontend/dist"

# Build everything (backend + frontend)
build-all: build-backend build-frontend
	@echo "✓ Build complete! Binary and frontend are in bin/"

# Increment version
version:
	@chmod +x scripts/increment-version.sh
	@NEW_VERSION=$$(bash scripts/increment-version.sh); \
	echo "Version incremented to $$NEW_VERSION"

# Build Docker image using pre-built binaries
docker-build: build-all
	@echo "Building Docker image $(IMAGE_NAME):$(IMAGE_TAG)..."
	docker build -t $(IMAGE_NAME):$(IMAGE_TAG) .
	docker tag $(IMAGE_NAME):$(IMAGE_TAG) $(IMAGE_NAME):latest
	@echo "✓ Docker image built: $(IMAGE_NAME):$(IMAGE_TAG)"
	@echo "✓ Tagged as: $(IMAGE_NAME):latest"

# Complete release: increment version, build everything, create Docker image
release: version build-all docker-build
	@echo "✓ Release $(VERSION) complete!"
	@echo "  - Binary: bin/server"
	@echo "  - Frontend: bin/frontend/dist"
	@echo "  - Docker: $(IMAGE_NAME):$(IMAGE_TAG)"

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
	@echo "  make frontend-install - Install npm dependencies"
	@echo "  make frontend-build   - Build frontend"
	@echo "  make build-backend    - Build Go binary"
	@echo "  make build-frontend   - Build frontend and copy to bin/"
	@echo "  make build-all        - Build backend + frontend"
	@echo ""
	@echo "  make version          - Increment version number"
	@echo "  make docker-build     - Build Docker image (after build-all)"
	@echo "  make release          - Full release: version + build + docker"
	@echo ""
	@echo "  make clean            - Remove build artifacts"
	@echo "  make help             - Show this help"
	@echo ""
	@echo "Current version: $(VERSION)"

