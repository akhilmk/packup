# Build System Summary

### 1. Makefile
New commands added:
- `make frontend-install` - First time Node build resources.
- `make db-up` - Run db from docker-compose.
- `make db-down` - Stop db from docker-compose.
- `make docker-build` - Build Go binary and HTLM resources to a docker image.
- `make docker-run` - Run docker image with latest tag
- `make docker-stop` - Stop docker image with latest tag
- `make docker-clean` - Remove docker image with latest tag
