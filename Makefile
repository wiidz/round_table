SERVER   := ./apps/server
APP      := roundtable
CMD      := $(SERVER)/cmd/roundtable
MEET_CMD := $(SERVER)/cmd/meet
DISCORD_CMD := $(SERVER)/cmd/discord
BIN_DIR  := ./bin

# China-friendly module proxy for local dev (see apps/server/README.md)
export GOPROXY := https://goproxy.cn,direct

SCENARIO_GAME_CLASS := data/_templates/scenarios/game-class-design
TOPIC_GAME_CLASS    := 设计新职业「影舞者」的核心技能与定位

WEB     := ./apps/web

.PHONY: run build test clean migrate tidy meet seed-scenario-3round meet-3round seed-scenario-game-class meet-game-class run-discord docker-build docker-up docker-down docker-logs docker-logs-discord server-dev server-build web-install web-reinstall web-dev web-preview web-build

SCENARIO_3ROUND := data/_templates/scenarios/3-round-debate
TOPIC_3ROUND    := 是否将用户认证拆为独立 Auth Service（JWT + Redis 撤销）并批准进入开发？

## run: start the server
run:
	go run $(CMD)/main.go

## run-discord: start Discord transport bot (requires DISCORD_BOT_TOKEN in deploy/.env)
run-discord:
	https_proxy=http://127.0.0.1:7897 http_proxy=http://127.0.0.1:7897 all_proxy=socks5://127.0.0.1:7897 \
	go run $(DISCORD_CMD)/main.go

## meet: run a meeting with DeepSeek (requires DEEPSEEK_API_KEY in deploy/.env)
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

## seed-scenario-game-class: copy deliberation scenario profiles
seed-scenario-game-class:
	@for p in designer ops player tech_lead; do \
		mkdir -p data/profiles/participants/$$p; \
		cp $(SCENARIO_GAME_CLASS)/profiles/$$p/SOUL.md data/profiles/participants/$$p/; \
		cp $(SCENARIO_GAME_CLASS)/profiles/$$p/AGENTS.md data/profiles/participants/$$p/; \
	done

## meet-game-class: game class deliberation scenario (see data/_templates/scenarios/game-class-design/README.md)
meet-game-class: seed-scenario-game-class
	go run $(MEET_CMD)/main.go -mode deliberation -topic "$(TOPIC_GAME_CLASS)" -max-rounds 2 \
		-agenda "skills:核心技能与资源机制,positioning:职业定位与差异化,monetization:商业化与活动联动,engineering:工程实现与平衡约束" \
		-participants "designer:游戏策划:gameplay,ops:运营:monetization,player:玩家代表:experience,tech_lead:主程:engineering"

## build: compile for the current OS
build: server-build

## server-dev: start HTTP server with hot reload (installs air if missing)
server-dev:
	@command -v air >/dev/null 2>&1 || (echo 'installing air...' && go install github.com/air-verse/air@latest)
	cd $(SERVER) && PATH="$$(go env GOPATH)/bin:$$PATH" air

## server-build: compile HTTP server binary
server-build:
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

## docker-build: build image (Web UI + API + Discord binaries)
docker-build:
	docker compose build server

## docker-up: start Web UI + API + Discord (Supervisor, host network)
docker-up:
	docker compose up -d --build

## docker-down: stop compose services
docker-down:
	docker compose down

## docker-logs: follow server container logs
docker-logs:
	docker compose logs -f server

## docker-logs-discord: follow Discord transport log file (Supervisor child)
docker-logs-discord:
	tail -f data/logs/discord-transport.log

## web-install: install web dependencies (npm ci — lockfile exact)
web-install:
	cd $(WEB) && npm ci

## web-reinstall: clean reinstall (fixes Vite 8 rolldown native binding on Linux)
web-reinstall:
	cd $(WEB) && rm -rf node_modules && npm ci

## web-dev: start Vite dev server (ROUND_TABLE_WEB_PORT / ROUND_TABLE_HTTP_PORT in deploy/.env)
web-dev: web-install
	cd $(WEB) && npm run dev

## web-preview: serve production build (same ports as web-dev)
web-preview: web-build
	cd $(WEB) && npm run preview

## web-build: production build for web (Vite → apps/web/dist)
web-build: web-install
	cd $(WEB) && npm run build
