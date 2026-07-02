# 快速上手

约 **15 分钟**在本地浏览 Web UI；配置 API Key 后可在 Discord 跑真实会议。

> 总览与一键命令见仓库根目录 [README.md](../README.md)。

## 前置条件

| 工具 | 版本 |
|------|------|
| Go | 1.25+（见 `apps/server/go.mod`） |
| Node.js | 22+ |
| Make | 任意 |

跑真实会议还需：

- [DeepSeek](https://platform.deepseek.com/) API Key → `DEEPSEEK_API_KEY`
- Discord Bot Token → `DISCORD_BOT_TOKEN`（见 [discord-transport.md](./adapters/discord-transport.md)）

## 1. 克隆与配置

```bash
git clone <your-fork-url> round_table
cd round_table

cp deploy/.env.example deploy/.env
# 仅浏览演示数据可暂不填 Key
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

## 4. Discord 跑会（可选）

在 `deploy/.env` 配置 `DEEPSEEK_API_KEY` 与 `DISCORD_BOT_TOKEN` 后：

```bash
make run-discord
```

频道内绑定委托人 → 发起会议 → 按向导或预设菜单配置 → 会中确认/驳回。

详见 [adapters/discord-transport.md](./adapters/discord-transport.md)。

## 5. Docker 部署（服务器）

```bash
cp deploy/.env.example deploy/.env
sh deploy/init-data-dirs.sh
make docker-up
```

详见 [deploy/README.md](../deploy/README.md)。

## 6. 开发与测试

```bash
make test
cd apps/web && npm test && npm run build
```

CI 在 `.github/workflows/ci.yml` 中运行相同检查。

本地 CLI 调试（需 `DEEPSEEK_API_KEY`）：

```bash
make meet TOPIC="你的议题" MEET_FLAGS='-max-rounds 2 -participants "a:Role:x,b:Role:y"'
```

## 目录速查

| 路径 | 说明 |
|------|------|
| `data/_templates/` | 入库模板与场景 |
| `data/workspaces/` | 会议产出（运行时，gitignore） |
| `apps/server/internal/engine/` | Meeting Engine |
| `apps/web/` | React 工作台 |
| `docs/CONSTITUTION.md` | 架构宪法 |

## 下一步

- [CONTRIBUTING.md](../CONTRIBUTING.md) — 贡献流程
- [data/README.md](../data/README.md) — 数据层与 Workspace 布局
- [roadmap.md](./roadmap.md) — 产品路线图
