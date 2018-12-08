# App
PKG_NAME = $(shell head -n 1 go.mod|sed 's/^module //g')
APP_NAME = $(shell basename $(PKG_NAME))

# Database
DB_MIGRATION_PATH = db/migration
DB_MIGRATE = @migrate -database 'postgres://localhost:5432/${APP_NAME}?sslmode=disable' -path $(DB_MIGRATION_PATH)
PSQL = @psql -h localhost -p 5432

default: help

help:
	@grep -E '^[\. a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

db.create: ## Create the database
	$(PSQL) -c 'create database ${APP_NAME};'

db.drop: ## Drop the database
	$(PSQL) -c 'drop database ${APP_NAME};'

db.generate: ## Generate migration with NAME=<migration_name>
	$(DB_MIGRATE) create -ext sql -dir $(DB_MIGRATION_PATH) $(NAME)

db.force: ## Set migration version V=<v> without running migration
	$(DB_MIGRATE) force $(V)

db.migrate: ## Apply all or N=<n> database migrations
	$(DB_MIGRATE) up $(N)

db.rollback: ## Rollback all or N=<n> database migrations
	$(DB_MIGRATE) down $(N)

db.seed: ## Seed the database with db/seed.sql
	$(PSQL) -f db/seed.sql

db.version: ## Show current migration version
	$(DB_MIGRATE) version

test: ## Run unit tests
	@echo $(APP_NAME)
