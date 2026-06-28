# RoundTable Web

Principal UI — React + Vite + TypeScript + Tailwind v4 + shadcn/ui。

设计规范见 [DESIGN.md](./DESIGN.md)（亮色 / 暗色双主题）。

## 要求

- Node **22+**（见 `.nvmrc`）

```bash
nvm use
npm install
```

## 开发

```bash
# 在 apps/web 目录
npm run dev

# 或仓库根目录
make web-dev
```

默认 **http://localhost:5173**（端口由 `ROUND_TABLE_WEB_PORT` 控制，见 `deploy/.env` 或 `apps/web/.env.local`）。

`/api` 代理到 `http://127.0.0.1:${ROUND_TABLE_HTTP_PORT:-7777}`；也可用 `ROUND_TABLE_API_PROXY` 指定完整地址。

```bash
# deploy/.env 示例
ROUND_TABLE_WEB_PORT=5173
ROUND_TABLE_HTTP_PORT=7777
```

## 预览生产构建

```bash
make web-preview   # build + vite preview，端口同 ROUND_TABLE_WEB_PORT
```

## 构建

```bash
npm run build
# 或仓库根目录
make web-build
```

产物目录：`apps/web/dist/`（静态文件，供 `ROUND_TABLE_WEB_ROOT` 或 Docker 镜像使用）。

### 生产部署（推荐）

**不要在服务器上 `make web-build`**，用 Docker 在干净 Node 镜像里构建：

```bash
docker compose up -d --build   # Dockerfile 内 npm ci && npm run build
```

### Linux 上 `make web-build` 报 rolldown native binding

Vite 8 依赖 `@rolldown/binding-*` 平台包，npm 偶发漏装 optional 依赖。处理：

```bash
make web-reinstall   # rm -rf node_modules && npm ci
make web-build
```

仍失败时手动安装当前平台 binding（x64 GNU 示例）：

```bash
cd apps/web && npm install @rolldown/binding-linux-x64-gnu@1.1.3
```

## 目录

```
src/
  api/           # fetch 薄封装
  components/    # UI 与 layout
  hooks/         # use-theme 等
  pages/         # 路由页面
  router/
  styles/        # theme.css 设计 token
```
