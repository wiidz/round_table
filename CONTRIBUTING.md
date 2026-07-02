# Contributing to RoundTable

感谢你对 RoundTable 的兴趣。本项目是 **Multi-Agent Meeting Engine**，不是通用 Agent 框架；贡献前请先理解领域语言与架构边界。

## 开始前必读

| 文档 | 用途 |
|------|------|
| [docs/CONSTITUTION.md](./docs/CONSTITUTION.md) | 架构宪法（**最高优先级**） |
| [docs/NAMING.md](./docs/NAMING.md) | Meeting / Moderator / Participant 等命名 |
| [docs/COMMITS.md](./docs/COMMITS.md) | Commit message 格式 |
| [docs/getting-started.md](./docs/getting-started.md) | 本地跑通 |

若实现与 Constitution 冲突，**以 Constitution 为准**；有歧义请在 Issue 讨论后再改代码。

## 开发环境

```bash
cp deploy/.env.example deploy/.env   # 填入 DEEPSEEK_API_KEY（跑真实 LLM 会议时）
make test
make seed-demo                       # 导入演示数据（无需 API key 可浏览 Web）
make server-dev                      # 终端 1：API :7777
make web-dev                         # 终端 2：Web :5177
```

中国大陆网络可设置：

```bash
export GOPROXY=https://goproxy.cn,direct
```

## 提交流程

1. Fork 仓库，从 `main` 拉分支（`feat/…`、`fix/…`）。
2. 小步提交，**每个 PR 只做一件事**。
3. 运行 `make test` 与 `cd apps/web && npm test && npm run build`。
4. Commit 遵循 [docs/COMMITS.md](./docs/COMMITS.md) 结构化格式。
5. 打开 Pull Request，说明 Context / 测试方式。

## 代码原则

- **通用方案**：不要为单个 workspace 或场景加特殊 case 启发式（见 `.cursorrules` → Generic Solutions）。
- **Go**：优先标准库；新依赖需有明确理由。
- **领域语言**：代码与文档使用 Moderator / Participant / Meeting，不要用 Main Agent / Sub Agent。
- **测试**：行为变更请补测试；纯 UI 微调可说明手动验证步骤。

## 报告问题

使用 [Bug report](./.github/ISSUE_TEMPLATE/bug_report.yml) 模板，附上复现步骤、环境与日志（**勿贴 Token / API Key**）。

## 许可证

贡献即表示你同意 [Apache-2.0](./LICENSE) 许可你的贡献。
