.PHONY: up up-build down stop build test

# Path to the application directory
APP_DIR=src/todo_app

# Get the current timestamp (YYYYMMDD_HHMMSS)
# Using PowerShell for cross-platform compatibility
TIMESTAMP=$(shell powershell -NoProfile -Command "Get-Date -Format 'yyyyMMdd_HHmmss'")
LATEST_DUMP=$(shell powershell -NoProfile -Command "Get-ChildItem -Filter '*_dump.sql' $(APP_DIR) | Sort-Object LastWriteTime -Descending | Select-Object -First 1 -ExpandProperty Name")

# Starts the docker containers in detached mode with a fresh build and restores latest DB dump if exists
up-build:
	cd $(APP_DIR) && docker compose up -d --build
ifneq ($(LATEST_DUMP),)
	@echo "Restoring database from latest dump: $(LATEST_DUMP)..."
	-docker compose -f $(APP_DIR)/docker-compose.yml exec -T app sqlite3 /app/data/webapp.sql "DROP TABLE IF EXISTS users; DROP TABLE IF EXISTS todos; DROP TABLE IF EXISTS sessions;"
	-docker compose -f $(APP_DIR)/docker-compose.yml exec -T app sqlite3 /app/data/webapp.sql < $(APP_DIR)/$(LATEST_DUMP)
endif

# Starts the docker containers in detached mode and restores latest DB dump if exists
up:
	cd $(APP_DIR) && docker compose up -d
ifneq ($(LATEST_DUMP),)
	@echo "Restoring database from latest dump: $(LATEST_DUMP)..."
	-docker compose -f $(APP_DIR)/docker-compose.yml exec -T app sqlite3 /app/data/webapp.sql "DROP TABLE IF EXISTS users; DROP TABLE IF EXISTS todos; DROP TABLE IF EXISTS sessions;"
	-docker compose -f $(APP_DIR)/docker-compose.yml exec -T app sqlite3 /app/data/webapp.sql < $(APP_DIR)/$(LATEST_DUMP)
endif

# Stops containers after dumping the database to a timestamped file
down:
	@echo "Dumping database to $(TIMESTAMP)_dump.sql..."
	-docker compose -f $(APP_DIR)/docker-compose.yml exec app sh -c "sqlite3 /app/data/webapp.sql .dump | grep -v sqlite_sequence" > $(APP_DIR)/$(TIMESTAMP)_dump.sql
	cd $(APP_DIR) && docker compose down

# Stops running docker containers without removing them
stop:
	cd $(APP_DIR) && docker compose stop

# Build the Go application locally
build:
	cd $(APP_DIR) && go build -o ../../bin/todo_app main.go

# Run tests
test:
	cd $(APP_DIR) && go test ./...
