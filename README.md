**English** | [中文](README.zh-CN.md)

# RoundTable

> **Build AI Teams, not AI Agents.**

*One problem. Many minds. One decision.*

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![CI](https://github.com/wiidz/round_table/actions/workflows/ci.yml/badge.svg)](https://github.com/wiidz/round_table/actions/workflows/ci.yml)

RoundTable is a **Multi-Agent Meeting Engine** — structured meetings where multiple AI experts debate and reach consensus, instead of stacking one stronger monolithic agent.

---

## What it is / is not

| | |
|---|---|
| ❌ Another AI agent | ✅ A Meeting Engine coordinating experts |
| ❌ Agent runtime (LangGraph / AutoGen / CrewAI) | ✅ Domain engine independent of runtime |
| ❌ Workflow engine | ✅ Structured discussion, not a task DAG |
| ❌ Chatbot | ✅ Chat is UI; discussion is architecture |

---

## Quick start

Pick either path. **Just want to explore the UI?** Run `make seed-demo` first — no API key required to browse a sample meeting in the Web app.

### Option A · Local dev (Go + Node)

**Requires:** Go 1.25+, Node.js 22+, Make (see `apps/server/go.mod`)

```bash
git clone https://github.com/wiidz/round_table.git
cd round_table

cp deploy/.env.example deploy/.env   # optional for demo-only browsing

make seed-demo      # import sample meeting (no API key)
make server-dev     # terminal 1: API → http://localhost:7777
make web-dev        # terminal 2: Web → http://localhost:5173
```

Open the Web app → **Meetings** → `mtg-demo-001` for overview, documents, flow, and replay.

**Run a real meeting on Discord** (set `DEEPSEEK_API_KEY` + `DISCORD_BOT_TOKEN`):

```bash
make run-discord    # separate terminal; see Makefile for proxy notes in CN
```

Step-by-step guide: [docs/getting-started.md](./docs/getting-started.md).

### Option B · Docker one-command deploy (servers)

**Requires:** Docker 24+, Docker Compose v2 (Linux recommended for production; Mac Docker Desktop host networking differs)

```bash
git clone https://github.com/wiidz/round_table.git
cd round_table

cp deploy/.env.example deploy/.env   # set DEEPSEEK_API_KEY, DISCORD_BOT_TOKEN
sh deploy/init-data-dirs.sh

make docker-up      # build & start Web + API + Discord Supervisor
```

Default: <http://127.0.0.1:7777>. Logs: `make docker-logs` / `make docker-logs-discord`.

See [deploy/README.md](./deploy/README.md).

---

## Core concepts (read this first)

| Term | Who | Role |
|------|-----|------|
| **Meeting** | The unit of discussion | One **Topic** from pre-meeting through closure; produces minutes, deliverables, and token usage; state is **event-sourced** and auditable |
| **Principal** | Human decision-maker | Starts meetings, sets scope, holds final **acceptance**; may approve or reject in the confirmation gate |
| **Moderator** | Facilitator agent | **Runs the room** — speaking order, round summaries, readiness for consensus/synthesis; expertise lives with Participants, not the Moderator |
| **Participant** | Domain expert agent | Role-bound specialists (e.g. design, engineering, ops); **speak only when invited**, no direct peer channels |
| **Meeting flow** | Engine standard path | Pre-meeting (Round 0) → debate/deliberation rounds (Round 1+) → optional free dialogue → Moderator summary → consensus / synthesis → optional **Principal confirmation** → closure & artifacts |

**Deliberation** mode focuses on a design draft; **decision** mode on actionable consensus. Rejection triggers more rounds and re-submission for confirmation.

```
Principal sets topic → Moderator schedules → Participants debate → consensus → [confirmation] → minutes / artifacts
                                              ↑                         │
                                              └──── reject → more rounds ┘
```

Details: [docs/domain/](./docs/domain/README.md) · Constitution: [CONSTITUTION.md](./docs/CONSTITUTION.md)

---

## Current support & roadmap

| Capability | Status |
|------------|--------|
| **LLM** | ✅ [DeepSeek](https://platform.deepseek.com/) (`DEEPSEEK_API_KEY`) |
| **Transport** | ✅ **Discord** (bot meetings, confirmation gate, free dialogue) |
| **Workbench** | ✅ Web UI (meetings, documents, flow, round-table replay) |
| **Planned** | 🔜 More model providers and transports (Slack, enterprise IM, …) — Engine core is adapter-decoupled; see **Architecture independence** below |

CLI local runs: `apps/server/cmd/meet` (developer tooling, not the primary end-user path).

---

## Project status

🚧 **Phase 1** — Engine runs end-to-end locally.  
🚧 **Phase 1.5** — Discord transport runs full meetings (Principal bind → brief wizard → presets → confirmation → deliverables).

Engine: event sourcing, pre-meeting, multi-round debate/deliberation, free dialogue, Moderator summaries, consensus/synthesis, confirmation (rejections & cycle limits), workspace projection, token accounting.  
Discord: natural language / `!rt` launch, three-step brief wizard, confirmation gate, runtime interventions, executive recap, multi-bot, Chinese i18n, etc. See [docs/adapters/discord-transport.md](./docs/adapters/discord-transport.md).

```
apps/server/cmd/discord/         # Discord transport
apps/server/cmd/roundtable/      # HTTP API
apps/server/internal/engine/     # Meeting Engine
apps/web/                        # React workbench
```

Roadmap: [docs/roadmap.md](./docs/roadmap.md)

---

## Documentation

| Doc | Description |
|-----|-------------|
| [getting-started.md](./docs/getting-started.md) | Hands-on setup (demo data, Discord, dev details) |
| [CONTRIBUTING.md](./CONTRIBUTING.md) | How to contribute |
| [SECURITY.md](./SECURITY.md) | Security reporting |
| [CONSTITUTION.md](./docs/CONSTITUTION.md) | Architecture constitution |
| [domain/](./docs/domain/README.md) | Domain model |
| [deploy/README.md](./deploy/README.md) | Docker deployment |
| [apps/web/README.md](./apps/web/README.md) | Web UI |
| [architecture/](./docs/architecture/README.md) | ADR index |

---

## Development & testing (contributors)

Common commands (full list in root `Makefile`):

| Command | Purpose |
|---------|---------|
| `make seed-demo` | Import demo meeting & profiles |
| `make server-dev` | API with hot reload |
| `make web-dev` | Frontend dev server |
| `make run-discord` | Local Discord bot |
| `make docker-up` | Docker Compose up |
| `make sync-data-pull` | Pull `data/` from deploy host (`deploy/sync-data.sh`) |

**Tests & CI:**

```bash
make test                    # Go unit/integration tests
cd apps/web && npm test      # Frontend Vitest
cd apps/web && npm run build # Frontend production build
```

CI: `.github/workflows/ci.yml`. Before contributing, read [CONTRIBUTING.md](./CONTRIBUTING.md) and [COMMITS.md](./docs/COMMITS.md).

---

## Design principles

1. **Everything is a Meeting** — not a workflow, prompt, or agent  
2. **Moderator controls the discussion** — Participants cannot interrupt  
3. **Participants own expertise** — stay within role boundaries  
4. **Consensus over Completion** — team alignment, not solo execution  
5. **Discussion is structured** — Round / Agenda / Minutes / Consensus / Confirmation  
6. **Principal owns the decision** — final authority stays with the Principal  

See [PRINCIPLES.md](./docs/PRINCIPLES.md).

---

## Architecture independence

The Meeting Engine **core domain does not depend** on a specific runtime, model, or transport — all plug in via adapters. See [CONSTITUTION.md](./docs/CONSTITUTION.md).

---

## License

[Apache 2.0](LICENSE)
