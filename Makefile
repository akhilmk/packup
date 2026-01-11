COMPOSE_FILE ?= compose.postgres-dev.yml
DOCKER_COMPOSE ?= docker compose

.PHONY: db-up db-down db-logs db-shell

db-up:
	$(DOCKER_COMPOSE) -f $(COMPOSE_FILE) up -d

db-down:
	$(DOCKER_COMPOSE) -f $(COMPOSE_FILE) down -v

db-logs:
	$(DOCKER_COMPOSE) -f $(COMPOSE_FILE) logs -f

db-shell:
	# open an interactive psql shell in the postgres service
	$(DOCKER_COMPOSE) -f $(COMPOSE_FILE) exec postgres psql -U $${DB_USER:-postgres} -d $${DB_NAME:-todos}
