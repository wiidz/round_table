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

## 2. 配置密钥

```bash
cp deploy/.env.example .env
nano .env   # 填入 DEEPSEEK_API_KEY、DISCORD_BOT_TOKEN 等
```

非敏感选项（参与者列表、locale、预设默认值）在镜像内 `apps/server/configs/server.yaml`；如需自定义，可挂载覆盖：

```yaml
# docker-compose.yml — discord 服务下追加
volumes:
  - ./apps/server/configs/server.yaml:/app/apps/server/configs/server.yaml:ro
```

## 3. 构建并启动 Discord Bot

```bash
# 若 ShellCrash mixed-port=4567，拉 golang 基础镜像可能也需要代理：
export HTTP_PROXY=http://127.0.0.1:4567 HTTPS_PROXY=http://127.0.0.1:4567

docker compose up -d --build discord
docker compose logs -f discord
```

> **apk / 镜像拉取失败**：Dockerfile 已切换 Alpine 源到 `mirrors.aliyun.com`。若仍卡在 `golang:1.25-alpine`，用上面的 `HTTP_PROXY` 再 build 一次。

成功日志示例：

```text
discord bot connected — prefix="!rt "
discord participant bots: 4/4 connected
```

## 4. Discord 内验证

1. 邀请 Bot 进服务器（`applications → OAuth2 → URL Generator`，权限：Send Messages、Read Message History）
2. 频道发送：`!rt principal bind`
3. 发送：`新会议` → 按主持人引导完成一场会

## 数据持久化

Compose 使用命名卷，重启/升级镜像后保留：

| 卷 | 内容 |
|----|------|
| `workspaces` | 会议产出（MINUTES、artifacts） |
| `profiles` | Participant / Principal Profile |
| `knowledge` | 长期记忆 |
| `transport` | Principal 绑定 `discord-principal.json` |

查看卷位置：

```bash
docker volume inspect roundtable_workspaces
```

备份示例：

```bash
docker run --rm -v roundtable_workspaces:/data -v $(pwd):/backup alpine \
  tar czf /backup/workspaces-$(date +%F).tar.gz -C /data .
```

## 可选：HTTP 健康检查服务

```bash
docker compose --profile http up -d
curl http://127.0.0.1:7777/health
```

## 常用命令

```bash
# 重建镜像并滚动重启
docker compose up -d --build discord

# 停止
docker compose down

# 停止并删除数据卷（慎用）
docker compose down -v

# 进入容器排查
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
| `apk add` / `no such package` | 多为 Alpine CDN 超时；`git pull` 最新 Dockerfile（阿里云 mirror）后 `--no-cache` 重建 |
| `golang:1.25-alpine` pull 慢/失败 | 构建前 `export HTTP_PROXY=http://127.0.0.1:4567` |
| `DISCORD_BOT_TOKEN required` | 根目录 `.env` 是否存在且 compose 能读到 |
| Bot 在线但不回消息 | Message Content Intent；`guild_id` 是否匹配 |
| LLM 报错 | `DEEPSEEK_API_KEY`；容器内 `wget -O- https://api.deepseek.com` |
| 重启后 Principal 要重新 bind | `transport` 卷是否挂载；是否用了 `docker compose down -v` |

## 文件说明

| 文件 | 作用 |
|------|------|
| `Dockerfile` | 多阶段构建 `roundtable-discord` / `roundtable-server` |
| `docker-compose.yml` | 服务定义与持久化卷 |
| `deploy/.env.example` | 环境变量模板 |
