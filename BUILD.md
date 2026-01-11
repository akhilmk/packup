# Go Todo App - Build & Deployment Guide

## Quick Start

```bash
# See all available commands
make help

# Full release (increment version, build, create Docker image)
make release
```

## Build Commands

### Development Build
```bash
# Build backend binary only
make build-backend

# Build frontend only
make build-frontend

# Build both (backend + frontend copied to bin/)
make build-all
```

### Production Release
```bash
# Complete release workflow:
# 1. Increment version (0.0.1 -> 0.0.2)
# 2. Build backend binary
# 3. Build and copy frontend to bin/
# 4. Create Docker image
make release
```

## Version Management

The project uses semantic versioning stored in the `VERSION` file.

```bash
# Manually increment version
make version

# Current version is read automatically by make commands
cat VERSION
```

Version format: `MAJOR.MINOR.PATCH` (e.g., `0.0.1`)

## Docker

### Building Docker Image

```bash
# Build with current version
make docker-build

# This creates:
# - go-todo:dev-X.Y.Z (versioned tag)
# - go-todo:latest (latest tag)
```

### Running Docker Container

```bash
# Run the container
docker run -d \
  -p 8080:8080 \
  -e DB_HOST=your-db-host \
  -e DB_PORT=5432 \
  -e DB_USER=postgres \
  -e DB_PASSWORD=yourpassword \
  -e DB_NAME=todos \
  --name go-todo \
  go-todo:latest

# View logs
docker logs -f go-todo

# Stop container
docker stop go-todo
docker rm go-todo
```

### Docker Image Details

- **Base Image**: Alpine Linux (minimal size)
- **Binary Location**: `/app/server`
- **Frontend Location**: `/app/frontend/dist`
- **Exposed Port**: 8080
- **Size**: ~15-20MB (optimized)

## Directory Structure

```
.
├── bin/                      # Build output
│   ├── server               # Go binary
│   └── frontend/
│       └── dist/            # Frontend static files
├── backend/                 # Go source code
│   ├── cmd/server/
│   ├── internal/
│   └── ...
├── frontend/                # Svelte source code
│   ├── src/
│   ├── dist/               # Build output (temporary)
│   └── ...
├── scripts/
│   └── increment-version.sh
├── Dockerfile
├── Makefile
└── VERSION                  # Current version
```

## Database

```bash
# Start PostgreSQL (development)
make db-up

# Stop PostgreSQL
make db-down

# View logs
make db-logs

# Open psql shell
make db-shell
```

## Frontend Development

```bash
# Install dependencies
make frontend-install

# Build frontend
make frontend-build

# Clean build artifacts
make frontend-clean
```

## Complete Workflow Example

### Development
```bash
# 1. Start database
make db-up

# 2. Install frontend dependencies (first time)
make frontend-install

# 3. Build everything
make build-all

# 4. Run the server
./bin/server

# Visit http://localhost:8080
```

### Production Release
```bash
# 1. Make your code changes
# 2. Test locally
# 3. Create release
make release

# 4. Push Docker image (if using registry)
docker tag go-todo:latest your-registry/go-todo:latest
docker push your-registry/go-todo:latest
```

## Environment Variables

The application supports the following environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | HTTP server port | `8080` |
| `DB_HOST` | PostgreSQL host | `localhost` |
| `DB_PORT` | PostgreSQL port | `5432` |
| `DB_USER` | Database user | `postgres` |
| `DB_PASSWORD` | Database password | `postgres` |
| `DB_NAME` | Database name | `todos` |

## Cleaning Up

```bash
# Remove all build artifacts
make clean

# This removes:
# - bin/ directory
# - frontend/dist directory
```

## Troubleshooting

### Frontend not loading
- Check if `bin/frontend/dist` exists after build
- Verify server logs show: "Serving static files from: frontend/dist"

### Database connection issues
- Ensure PostgreSQL is running: `make db-up`
- Check environment variables are set correctly
- Verify database exists: `make db-shell`

### Docker build fails
- Ensure `make build-all` completes successfully first
- Check that `bin/server` and `bin/frontend/dist` exist
- Verify Docker daemon is running

## CI/CD Integration

Example GitHub Actions workflow:

```yaml
name: Build and Release

on:
  push:
    branches: [main]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.22'
      
      - name: Set up Node
        uses: actions/setup-node@v3
        with:
          node-version: '20'
      
      - name: Build release
        run: make release
      
      - name: Push Docker image
        run: |
          docker login -u ${{ secrets.DOCKER_USERNAME }} -p ${{ secrets.DOCKER_PASSWORD }}
          docker push go-todo:latest
```

## License

[Your License Here]
