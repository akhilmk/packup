# Development Commands

This project uses a `Makefile` to simplify development tasks. Environment variables are automatically loaded from `docker/.env.dev`.

## Development Environment Commands

- `make dev-build-up`: Starts the development PostgreSQL database and persistent Node.js builder in the background.
- `make dev-build-down`: Stops the dev environment and removes the containers and volumes (data is reset).
- `make dev-build-logs`: Follows the output logs of the dev environment containers.
- `make dev-build-db-shell`: Opens an interactive PSQL shell within the database container.

## Build Commands

- `make frontend-install`: **[First Time]** Installs the required NPM dependencies for the frontend. (Runs `npm install` inside node container)
- `make frontend-audit-fix`: If audit error occurs during `make frontend-install`, run this command to fix it.
- `make build-frontend`: **[Develop Time]** Build only frontend and copy to bin folder, copy UI changes to running container. (Runs `frontend-build`)
- `make docker`: **[App Docker Image]** Removes old containers, images, and build files, then builds both the frontend and backend, and finally **creates a new local Docker image**. (Runs `docker-stop docker-clean clean build-all`)


## App Run Commands

- `make run`: Starts the application container locally. It automatically handles container removal if one is already running. (Runs `docker rm` then `docker run`)
- `make logs`: Follows the application container logs.
- `make app-shell`: Opens an interactive shell inside the running application container for debugging.

## Stop and Clean Commands

- `make clean`: Deletes local build artifacts (`bin/` and `frontend/dist`).
- `make docker-stop`: Stops the application container.
- `make docker-clean`: Removes the application container and deletes the local Docker image. (Runs `docker rm` and `docker rmi`)
- `make help`: Displays a summary of all available commands.

## Testing Commands

- `make go-test`: Runs all Go backend unit tests with verbose output.




## API Documentation Commands (Swagger)

If you make changes to the API (handlers, models, or global info), you need to update the Swagger documentation:

1.  **Update Annotations**: Modify the `@Summary`, `@Description`, `@Param`, or `@Success` comments in the backend handler files.
2.  **Generate Documentation**: Run the following command from the project root:
    - `make swagger`: Updates the OpenAPI specification and Swagger UI files.
3.  **Verify**: Start the app (`make run`) and visit [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html).

---
