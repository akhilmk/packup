# Development Commands

This project uses a `Makefile` to simplify development tasks. Environment variables are automatically loaded from `docker/.env.dev`.

## Development Environment Commands

- `make dev-up`: Starts the development PostgreSQL database and persistent Node.js builder in the background.
- `make dev-down`: Stops the dev environment and removes the containers and volumes (data is reset).
- `make dev-logs`: Follows the output logs of the dev environment containers.
- `make db-shell`: Opens an interactive PSQL shell within the database container.

## Build Commands

- `make frontend-install`: **[First Time]** Installs the required NPM dependencies for the frontend.
- `make frontend-audit-fix`: If audit error occurs during `make frontend-install`, run this command to fix it.
- `make build-frontend`: **[Develop Time]** Build only frontend and copy to bin folder, copy UI changes to running container (using volume path mount).
- `make docker`: **[App DockerImage]** Removes old containers, images, and build files, then builds both the frontend and backend, and finally **creates a new local Docker image**.


## Runtime Commands

- `make run`: Starts the application container locally. It automatically handles container removal if one is already running.
- `make logs`: Follows the application container logs.
- `make app-shell`: Opens an interactive shell inside the running application container for debugging.

## Stop and Clean Commands

- `make clean`: Deletes local build artifacts (`bin/` and `frontend/dist`).
- `make docker-clean`: Stops and removes the application container and deletes the local Docker image.
- `make help`: Displays a summary of all available commands.

## Testing Commands

- `make go-test`: Runs all Go backend unit tests with verbose output.

## Production Commands (via docker/Makefile)

These commands should be run using the `-f` flag from the project root:

- `make -f docker/Makefile docker-release`: Builds the entire project, creates a Docker image, and pushes it to Docker Hub using the `DOCKER_HUB_USER` defined in `.env.prod`.
- `make -f docker/Makefile prod-up`: Starts the production stack (App + DB + Traefik) using the production compose file and environment.
- `make -f docker/Makefile prod-down`: Stops the production stack.


