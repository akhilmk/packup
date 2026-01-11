# Build System Summary

## What Was Created

### 1. Version Management
- **VERSION** - Tracks current version (starts at 0.0.1)
- **scripts/increment-version.sh** - Bash script to auto-increment patch version

### 2. Docker Support
- **Dockerfile** - Multi-stage build using Alpine Linux
- **.dockerignore** - Optimizes Docker build context

### 3. Enhanced Makefile
New commands added:
- `make build-backend` - Build Go binary to bin/server
- `make build-frontend` - Build frontend and copy to bin/frontend/dist
- `make build-all` - Build both backend and frontend
- `make version` - Increment version number
- `make docker-build` - Create Docker image from pre-built binaries
- `make release` - Complete release workflow
- `make clean` - Remove build artifacts
- `make help` - Show all available commands

### 4. Updated Backend
- Modified `main.go` to support multiple static file paths
- Works in development and production environments

### 5. Documentation
- **BUILD.md** - Comprehensive build and deployment guide

## Usage

### Quick Release
```bash
make release
```

This will:
1. Increment version (0.0.1 → 0.0.2)
2. Build Go binary → bin/server
3. Build frontend → bin/frontend/dist
4. Create Docker image → go-todo:dev-0.0.2 and go-todo:latest

### Manual Steps
```bash
# Step by step
make version              # Increment version
make build-all           # Build everything
make docker-build        # Create Docker image
```

## Docker Image Tags

Each build creates two tags:
- `go-todo:dev-X.Y.Z` - Versioned tag (e.g., go-todo:dev-0.0.2)
- `go-todo:latest` - Always points to latest build

## Directory Structure After Build

```
bin/
├── server                 # Go binary (Linux executable)
└── frontend/
    └── dist/             # Static frontend files
        ├── index.html
        ├── assets/
        └── vite.svg
```

## Running the Application

### Local (from bin/)
```bash
cd bin
./server
```

### Docker
```bash
docker run -p 8080:8080 \
  -e DB_HOST=localhost \
  -e DB_PASSWORD=yourpass \
  go-todo:latest
```

## Notes

- Binary is built for Linux (suitable for Docker/production)
- Frontend is copied to bin/ so the binary can serve it
- Docker image uses pre-built binaries (no compilation in Docker)
- Version auto-increments on each `make release`
