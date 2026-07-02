# 快速上手

约 **30 分钟**在本地浏览 RoundTable Web UI，或 **1 小时**跑一场真实 LLM 会议。

## 前置条件

| 工具 | 版本 |
|------|------|
| Go | 1.25+（见 `apps/server/go.mod`） |
| Node.js | 22+ |
| Make | 任意 |

可选（真实 LLM / Discord）：

- [DeepSeek](https://platform.deepseek.com/) API Key → `DEEPSEEK_API_KEY`
- Discord Bot Token → `DISCORD_BOT_TOKEN`（见 [discord-transport.md](./adapters/discord-transport.md)）

## 1. 克隆与配置

```bash
git clone <your-fork-url> round_table
cd round_table

cp deploy/.env.example deploy/.env
# 仅浏览演示数据可暂不填 Key；跑 make meet-3round 需填入 DEEPSEEK_API_KEY
```

中国大陆建议：

```bash
export GOPROXY=https://goproxy.cn,direct
```

## 2. 导入演示数据（推荐）

无需 API Key，即可在 Web 中查看一场已结束的示例会议：

```bash
make seed-demo
```

将复制到 `data/`（gitignore）：

- `workspaces/mtg-demo-001/` — 完整 Workspace 产出
- `profiles/participants/{skeptic,pragmatist}/` — 专家档案
- `profiles/principals/demo/` — 委托人档案
- `briefs/decision-review/` — 简报模板样例

## 3. 启动 Web 工作台

**终端 1 — API**

```bash
make server-dev
# 或：make run
```

默认 <http://localhost:7777/health>

**终端 2 — 前端**

```bash
make web-dev
```

默认 <http://localhost:5173>（端口见终端输出）。

打开 **会议** → `mtg-demo-001`，可体验概览、文档三栏、流程结局与 Markdown 目录。

## 4. 跑一场真实会议（可选）

```bash
make seed-scenario-3round   # 仅复制 3 轮辩论场景专家档案
make meet-3round            # 需 deploy/.env 中 DEEPSEEK_API_KEY
```

议题：是否将用户认证拆为独立 Auth Service。完成后在 Web **会议** 列表刷新即可看到新 `mtg-*`。

其他场景：

```bash
make meet-game-class        # 研讨型 · 游戏职业设计
make meet TOPIC="你的议题" MEET_FLAGS='-max-rounds 2 -participants "a:Role:x,b:Role:y"'
```

## 5. Discord（可选）

```bash
# deploy/.env 配置 DISCORD_BOT_TOKEN
make run-discord
```

详见 [adapters/discord-transport.md](./adapters/discord-transport.md)。

## 6. 测试

```bash
make test
cd apps/web && npm test && npm run build
```

CI 在 `.github/workflows/ci.yml` 中运行相同检查。

## 目录速查

| 路径 | 说明 |
|------|------|
| `data/_templates/` | 入库模板与场景（`make seed-*` 来源） |
| `data/workspaces/` | 会议产出（运行时，gitignore） |
| `apps/server/internal/engine/` | Meeting Engine |
| `apps/web/` | React 工作台 |
| `docs/CONSTITUTION.md` | 架构宪法 |

## 下一步

- [CONTRIBUTING.md](../CONTRIBUTING.md) — 贡献流程
- [data/README.md](../data/README.md) — 数据层与 Workspace 布局
- [roadmap.md](./roadmap.md) — 产品路线图
