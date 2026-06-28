# Ubuntu Docker 部署

在 Linux 服务器上以 Docker 运行 RoundTable：**Web UI + REST API + Discord Transport**（单容器、`network_mode: host`）。

## 前置条件

- Ubuntu 22.04+（或其他 Linux）
- [Docker Engine](https://docs.docker.com/engine/install/) 24+
- [Docker Compose](https://docs.docker.com/compose/install/) v2
- Discord Bot（[开发者门户](https://discord.com/developers/applications)）已开启 **Message Content Intent**
- DeepSeek（或其他 OpenAI 兼容）API Key

> **Mac Docker Desktop**：`network_mode: host` 行为与 Linux 不同，生产部署请用 Linux 服务器。

## 1. 拉代码

```bash
git clone <your-repo-url> round_table
cd round_table
```

## 2. 配置密钥与数据目录

```bash
cp deploy/.env.example deploy/.env
nano deploy/.env   # 填入 DEEPSEEK_API_KEY、DISCORD_BOT_TOKEN 等

sh deploy/init-data-dirs.sh   # 创建 ./data/workspaces 等目录
```

非敏感选项（参与者列表、locale、预设默认值）在镜像内 `apps/server/configs/server.yaml`；如需自定义，可挂载覆盖：

```yaml
# docker-compose.yml — server 服务下追加
volumes:
  - ./apps/server/configs/server.yaml:/app/apps/server/configs/server.yaml:ro
```

## 3. 构建并启动

```bash
# 若 ShellCrash mixed-port=4567，拉基础镜像可能也需要代理：
export HTTP_PROXY=http://127.0.0.1:4567 HTTPS_PROXY=http://127.0.0.1:4567

# 从双容器升级到单容器时务必清理旧 discord 容器（否则会重复回复）
sh deploy/ensure-clean-discord.sh

docker compose up -d --build
# 或：make docker-up
docker compose logs -f server
```

- **Web UI + REST API**：`http://127.0.0.1:<ROUND_TABLE_ADDR 端口>`（host 网络，静态页 + `/api`）
- **Discord Bot**：由 `roundtable-server` 内 **Supervisor** 自动拉起子进程 `roundtable-discord`（`ROUND_TABLE_DISCORD_AUTO_START=true`）

`deploy/.env` 端口示例：

```bash
ROUND_TABLE_ADDR=:7777          # host 网络下 HTTP 监听（Web + API）
# 或留空 ROUND_TABLE_ADDR，由 ROUND_TABLE_HTTP_PORT / ROUND_TABLE_WEB_PORT 推导
ROUND_TABLE_HTTP_PORT=7777
ROUND_TABLE_WEB_PORT=5173       # 仅本地 make web-dev
```

Web 设置 → IM → Discord 可查看 Transport 状态、启停与日志（与 Supervisor 同一进程树）。

> **go mod / npm 慢**：Dockerfile 使用 `GOPROXY=https://goproxy.cn,direct`；构建时经 `HTTP_PROXY` 拉基础镜像。

成功日志示例（`docker compose logs server` 或 `data/logs/discord-transport.log`）：

```text
discord transport auto-started
discord bot connected — prefix="!rt "
discord participant bots: 4/4 connected
```

## 4. Discord 内验证

1. 邀请 Bot 进服务器（`applications → OAuth2 → URL Generator`，权限：Send Messages、Read Message History）
2. 频道发送：`!rt principal bind`
3. 发送：`新会议` → 按主持人引导完成一场会

## 数据持久化（bind mount → 项目 `data/`）

会议产出直接写在宿主机 **`./data/`** 下（与本地开发路径一致）：

| 宿主机路径 | 内容 |
|------------|------|
| `data/workspaces/` | 会议产出（MINUTES、artifacts） |
| `data/profiles/` | Participant / Principal Profile |
| `data/knowledge/` | 长期记忆 |
| `data/transport/` | Principal 绑定 `discord-principal.json` |
| `data/logs/` | Discord Transport 日志（Supervisor 子进程） |

示例：`/mnt/data1/projects/round_table/data/workspaces/mtg-xxx/`

> 此前若用过 **Docker 命名卷**或独立的 `roundtable-discord` 容器，升级后只需 `docker compose up -d --build` 单服务即可。

备份：

```bash
tar czf workspaces-$(date +%F).tar.gz -C data workspaces
```

## 常用命令

```bash
# 重建镜像并滚动重启
docker compose up -d --build

# 停止（不删 data/ 下文件）
docker compose down

# 进入容器排查
docker compose exec server sh

# Discord 日志（Supervisor 写入 bind mount）
tail -f data/logs/discord-transport.log
# 或：make docker-logs-discord
```

## 代理 / ShellCrash（mixed-port 4567）

`server` 服务使用 **`network_mode: host`**（Linux 专用）：容器与宿主机共享网络栈，`.env` 里可直接写 **`127.0.0.1:4567`**，Supervisor 拉起的 Discord 子进程继承同一网络与代理环境。

| 场景 | 配置 |
|------|------|
| **容器运行时** | `.env`：`http_proxy` + **`https_proxy`** 均指向 `http://USER:PASS@127.0.0.1:4567` |
| **docker build** | 宿主机 `export HTTP_PROXY=http://127.0.0.1:4567 HTTPS_PROXY=http://127.0.0.1:4567` |

验证：

```bash
docker compose up -d --force-recreate server
docker compose logs -f server
# 或
tail -f data/logs/discord-transport.log
```

应出现 `discord bot connected`。

## 防火墙

放行 `ROUND_TABLE_ADDR` 对应端口（默认 7777），供 Web UI / API 访问。

## 自定义 Participant Profile

```yaml
volumes:
  - ./data/profiles/participants/designer:/app/data/profiles/participants/designer:ro
```

或在宿主机直接编辑 `data/profiles/` 下文件。

## 故障排查

| 现象 | 检查 |
|------|------|
| `172.17.0.1:4567: connection refused` | 是否仍用旧桥接 + 独立 discord 容器；升级后应为 host 网络 + `127.0.0.1:4567` |
| `TLS handshake timeout` / `open gateway` | 缺 `https_proxy`；或代理地址/认证错误 |
| `apk add` / `could not connect` | 构建前 `export HTTP_PROXY=http://127.0.0.1:4567 HTTPS_PROXY=...` |
| `golang:1.25-alpine` pull 慢/失败 | 构建前 `export HTTP_PROXY=http://127.0.0.1:4567` |
| `DISCORD_BOT_TOKEN required` | `deploy/.env` 是否存在且 compose 能读到 |
| Bot 在线但不回消息 | Message Content Intent；`guild_id` 是否匹配 |
| LLM 报错 | `DEEPSEEK_API_KEY`；容器内 `wget -O- https://api.deepseek.com` |
| 重启后 Principal 要重新 bind | `transport` 卷是否挂载；是否用了 `docker compose down -v` |
| 会议跑完但 `data/workspaces` 空 | bind mount 是否指向 `./data`；`sh deploy/init-data-dirs.sh` |
| entrypoint `cannot write to /app/data/workspaces` | 宿主机 `data/workspaces` 权限；`sudo chown -R 1000:1000 data/` |
| Web 显示 Discord 已停但 Bot 在线 | 旧版双容器残留；`sh deploy/ensure-clean-discord.sh` |
| **主持人消息成对重复**（如两次「请输入会议主题」） | **两个 Discord Transport 同时在线**（旧 `roundtable-discord` 容器 + Supervisor 子进程）；见下方 |

## 数据卷权限（entrypoint）

启动时会 `chown` bind mount 目录为 uid **1000** 并做写入自检。升级后：

```bash
git pull
sh deploy/init-data-dirs.sh
docker compose up -d --build --force-recreate server
ls -la data/workspaces
```

无需 `chmod 777`。

## 主持人消息重复（双 Transport）

升级单容器后，若 **未删除** 旧 `roundtable-discord` 容器，会出现两个进程用同一 Bot Token 连 Discord，同一条「新会议」会触发 **两次**「请输入会议主题」。

```bash
docker ps -a | grep roundtable
pgrep -af roundtable-discord   # 正常应只有 1 条（Supervisor 子进程）

sh deploy/ensure-clean-discord.sh
docker compose up -d --build --force-recreate server
```

确认日志只有一条 `discord transport auto-started`，且 `data/logs/discord-transport.log` 里只有一次 `discord bot connected`。

## 文件说明

| 文件 | 作用 |
|------|------|
| `Dockerfile` | 多阶段构建 `roundtable-discord` + `roundtable-server`（Discord 由 Supervisor 拉起） |
| `deploy/docker-entrypoint.sh` | chown 数据卷、解析 `ROUND_TABLE_ADDR`、降权 uid 1000 |
| `docker-compose.yml` | 单服务 `server`（host 网络） |
| `deploy/.env.example` | 环境变量模板 |
