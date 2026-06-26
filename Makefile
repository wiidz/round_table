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

.PHONY: run build test clean migrate tidy meet seed-scenario-3round meet-3round seed-scenario-game-class meet-game-class run-discord

SCENARIO_3ROUND := data/_templates/scenarios/3-round-debate
TOPIC_3ROUND    := 是否将用户认证拆为独立 Auth Service（JWT + Redis 撤销）并批准进入开发？

## run: start the server
run:
	go run $(CMD)/main.go

## run-discord: start Discord transport bot (requires DISCORD_BOT_TOKEN in apps/server/.env)
run-discord:
	go run $(DISCORD_CMD)/main.go

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
