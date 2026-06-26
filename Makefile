SERVER   := ./apps/server
APP      := roundtable
CMD      := $(SERVER)/cmd/roundtable
MEET_CMD := $(SERVER)/cmd/meet
BIN_DIR  := ./bin

# China-friendly module proxy for local dev (see apps/server/README.md)
export GOPROXY := https://goproxy.cn,direct

.PHONY: run build test clean migrate tidy meet seed-scenario-3round meet-3round

SCENARIO_3ROUND := data/_templates/scenarios/3-round-debate
TOPIC_3ROUND    := 是否将用户认证拆为独立 Auth Service（JWT + Redis 撤销）并批准进入开发？

## run: start the server
run:
	go run $(CMD)/main.go

## meet: run a meeting with DeepSeek (requires DEEPSEEK_API_KEY in apps/server/.env)
meet:
	@test -n "$(TOPIC)" || (echo 'usage: make meet TOPIC="your topic"'; exit 1)
	go run $(MEET_CMD)/main.go -topic "$(TOPIC)" $(MEET_FLAGS)

## seed-scenario-3round: copy scenario profiles for skeptic + pragmatist
seed-scenario-3round:
	@mkdir -p data/profiles/participants/skeptic data/profiles/participants/pragmatist
	cp $(SCENARIO_3ROUND)/profiles/skeptic/SOUL.md data/profiles/participants/skeptic/
	cp $(SCENARIO_3ROUND)/profiles/skeptic/AGENTS.md data/profiles/participants/skeptic/
	cp $(SCENARIO_3ROUND)/profiles/pragmatist/SOUL.md data/profiles/participants/pragmatist/
	cp $(SCENARIO_3ROUND)/profiles/pragmatist/AGENTS.md data/profiles/participants/pragmatist/

## meet-3round: 3-round debate scenario (see data/_templates/scenarios/3-round-debate/README.md)
meet-3round: seed-scenario-3round
	go run $(MEET_CMD)/main.go -topic "$(TOPIC_3ROUND)" -max-rounds 3 \
		-participants "skeptic:Security Architect:security,pragmatist:Tech Lead:delivery"

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
