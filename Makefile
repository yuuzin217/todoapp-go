.PHONY: up down stop build test

# Path to the application directory
APP_DIR=src/todo_app

# Starts the docker containers in detached mode
up:
	cd $(APP_DIR) && docker compose up -d

# Stops and removes the docker containers, networks, and images created by up
down:
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
