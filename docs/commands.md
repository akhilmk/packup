# Development Commands

This project uses a `Makefile` to simplify development tasks. Environment variables are automatically loaded from `docker/.env.dev`.

## Database Commands

- `make db`: Starts the development PostgreSQL database in the background.
- `make db-down`: Stops the database and removes the containers and volumes (data is reset).
- `make db-logs`: Follows the output logs of the database container.
- `make db-shell`: Opens an interactive PSQL shell within the database container.

## Build Commands

- `make docker`: Performs a "deep clean" (removes old containers, images, and build files), then builds both the frontend and backend, and finally creates a new local Docker image.
- `make build-all`: Builds the backend and frontend artifacts locally in the `bin/` directory.
- `make frontend-install`: Installs the required NPM dependencies for the frontend.

## Runtime Commands

- `make run`: Starts the application container locally. It automatically handles container removal if one is already running.
- `make logs`: Follows the application container logs.
- `make app-shell`: Opens an interactive shell inside the running application container for debugging.

## Testing Commands

- `make go-test`: Runs all Go backend unit tests with verbose output.

## Utility Commands

- `make clean`: Deletes local build artifacts (`bin/` and `frontend/dist`).
- `make docker-clean`: Stops and removes the application container and deletes the local Docker image.
- `make help`: Displays a summary of all available commands.
