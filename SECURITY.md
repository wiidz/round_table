# Security Policy

## 支持的版本

| 版本 | 支持 |
|------|------|
| `main` 分支最新提交 | ✅ |

## 报告漏洞

**请勿在公开 Issue 中披露可利用的安全问题。**

请通过仓库维护者私下联系（GitHub Security Advisory 或维护者邮箱），并包含：

- 问题描述与影响范围
- 复现步骤或 PoC
- 受影响组件（server / web / discord transport）

我们会在确认后尽快回复。

## 敏感信息

以下内容 **不得** 提交到 Git：

- `deploy/.env`、`apps/server/.env` 及任何 API Key / Bot Token
- `data/workspaces/`、`data/logs/` 等运行时目录中的真实会议内容
- 含客户/内部业务机密的 workspace 导出

`.gitignore` 已忽略常见路径；提交前请运行 `git status` 确认。

## 部署建议

- 生产环境通过环境变量或密钥管理服务注入 `DEEPSEEK_API_KEY`、`DISCORD_BOT_TOKEN`
- 不要将 HTTP API 无鉴权暴露到公网（当前版本面向受信网络 / 本地开发）
- 定期轮换 Discord Bot Token
