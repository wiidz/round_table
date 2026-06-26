SERVER   := ./apps/server
APP      := roundtable
CMD      := $(SERVER)/cmd/roundtable
BIN_DIR  := ./bin

.PHONY: run build test clean migrate tidy

## run: start the server
run:
	go run $(CMD)/main.go

## build: compile for the current OS
build:
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(APP) $(CMD)/main.go

## test: run all server tests
test:
	go test $(SERVER)/...

## migrate: run database migrations
migrate:
	go run $(SERVER)/cmd/migrate/main.go

## tidy: tidy go modules
tidy:
	cd $(SERVER) && go mod tidy

## clean: remove compiled binaries
clean:
	rm -rf $(BIN_DIR)
