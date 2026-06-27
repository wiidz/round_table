# Ubuntu Docker 部署

在 Linux 服务器上以 Docker 运行 RoundTable **Discord Transport**（可选 HTTP `/health`）。

## 前置条件

- Ubuntu 22.04+（或其他 Linux）
- [Docker Engine](https://docs.docker.com/engine/install/) 24+
- [Docker Compose](https://docs.docker.com/compose/install/) v2
- Discord Bot（[开发者门户](https://discord.com/developers/applications)）已开启 **Message Content Intent**
- DeepSeek（或其他 OpenAI 兼容）API Key

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
# docker-compose.yml — discord 服务下追加
volumes:
  - ./apps/server/configs/server.yaml:/app/apps/server/configs/server.yaml:ro
```

## 3. 构建并启动（Web + API + Discord）

```bash
# 若 ShellCrash mixed-port=4567，拉基础镜像可能也需要代理：
export HTTP_PROXY=http://127.0.0.1:4567 HTTPS_PROXY=http://127.0.0.1:4567

docker compose up -d --build
# 或：make docker-up
docker compose logs -f
```

- **Web UI + REST API**：`http://127.0.0.1:${ROUND_TABLE_HTTP_PORT:-7777}`（同一端口，静态页 + `/api`）
- **Discord Bot**：`discord` 容器，`network_mode: host`（Linux 代理用 `127.0.0.1`）

`deploy/.env` 端口：

```bash
ROUND_TABLE_HTTP_PORT=7777   # 宿主机访问 Web/API
ROUND_TABLE_WEB_PORT=5173    # 仅本地 make web-dev（Vite），Docker 不用
```

> **go mod / npm 慢**：Dockerfile 使用 `GOPROXY=https://goproxy.cn,direct`；构建时经 `HTTP_PROXY` 拉基础镜像。

成功日志示例：

```text
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

示例：`/mnt/data1/projects/round_table/data/workspaces/mtg-xxx/`

> 此前若用过 **Docker 命名卷**，旧数据在 `docker volume inspect roundtable_workspaces`，需手动拷贝到 `data/workspaces/` 后删除旧卷。

备份：

```bash
tar czf workspaces-$(date +%F).tar.gz -C data workspaces
```

## 可选：仅 Discord（不启 HTTP/Web）

```bash
docker compose up -d --build discord
```

## 常用命令

```bash
# 重建镜像并滚动重启（Web + API + Discord）
docker compose up -d --build

# 停止（不删 data/ 下文件）
docker compose down

# 进入容器排查
docker compose exec server sh
docker compose exec discord sh
```

## 代理 / ShellCrash（mixed-port 4567）

`discord` 服务使用 **`network_mode: host`**（Linux 专用）：容器与宿主机共享网络栈，`.env` 里可直接写 **`127.0.0.1:4567`**，无需 `host.docker.internal` / `172.17.0.1`。

ShellCrash 默认只监听本机时，桥接网络会报 `172.17.0.1:4567: connection refused`；host 网络可绕过。

| 场景 | 配置 |
|------|------|
| **容器运行时** | `.env`：`http_proxy` + **`https_proxy`** 均指向 `http://USER:PASS@127.0.0.1:4567` |
| **docker build** | 宿主机 `export HTTP_PROXY=http://127.0.0.1:4567 HTTPS_PROXY=http://127.0.0.1:4567` |

验证：

```bash
docker compose up -d --force-recreate discord
docker compose logs -f discord
```

应出现 `discord bot connected`。

## 防火墙

## 自定义 Participant Profile

将本地 profile 挂进卷（首次需先 `docker compose up` 创建卷，或 bind mount）：

```yaml
volumes:
  - ./data/profiles/participants/designer:/app/data/profiles/participants/designer:ro
```

或在宿主机直接编辑卷内文件（路径见 `docker volume inspect`）。

## 故障排查

| 现象 | 检查 |
|------|------|
| `172.17.0.1:4567: connection refused` | ShellCrash 只监听 127.0.0.1；`git pull` 后用 host 网络 + `.env` 改回 `127.0.0.1:4567` |
| `TLS handshake timeout` / `open gateway` | 缺 `https_proxy`；或代理地址/认证错误 |
| `apk add` / `could not connect` | 构建前 `export HTTP_PROXY=http://127.0.0.1:4567 HTTPS_PROXY=...` |
| `golang:1.25-alpine` pull 慢/失败 | 构建前 `export HTTP_PROXY=http://127.0.0.1:4567` |
| `DISCORD_BOT_TOKEN required` | 根目录 `.env` 是否存在且 compose 能读到 |
| Bot 在线但不回消息 | Message Content Intent；`guild_id` 是否匹配 |
| LLM 报错 | `DEEPSEEK_API_KEY`；容器内 `wget -O- https://api.deepseek.com` |
| 重启后 Principal 要重新 bind | `transport` 卷是否挂载；是否用了 `docker compose down -v` |
| 会议跑完但 `data/workspaces` 空 | 是否仍用旧命名卷；`git pull` 后 bind mount + `sh deploy/init-data-dirs.sh` 重建 |
| entrypoint `cannot write to /app/data/workspaces` | 宿主机 `data/workspaces` 权限；`sudo chown -R 1000:1000 data/` |

## 数据卷权限（entrypoint）

启动时会 `chown` bind mount 目录为 uid **1000** 并做写入自检。升级后：

```bash
git pull
sh deploy/init-data-dirs.sh
docker compose up -d --build --force-recreate discord
ls -la data/workspaces
```

无需 `chmod 777`。

## 文件说明

| 文件 | 作用 |
|------|------|
| `Dockerfile` | 多阶段构建 `roundtable-discord` / `roundtable-server` |
| `deploy/docker-entrypoint.sh` | 启动时 chown 数据卷并降权为 uid 1000 |
| `docker-compose.yml` | 服务定义与持久化卷 |
| `deploy/.env.example` | 环境变量模板 |
