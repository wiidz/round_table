# RoundTable 路线图

> 与 [CONSTITUTION.md](./CONSTITUTION.md) 一致；架构变更需 ADR。

最后更新：2026-07-03（v0.2 Web 工作台与 Brief 模板对齐）

---

## 已完成（v0.2）

### Discord 切片

| 能力 | 说明 |
|------|------|
| Principal 绑定 | `!rt principal bind`，每服务器/DM 一位 |
| 会议发起 | 自然语言 / `!rt meet`，预设 1–6 / J1–J5 / 自定义 |
| 确认关 | 批准/驳回、ItemNotes、触顶三选一 |
| 运行期干预 | 暂停/恢复/终止、立即合成/强制共识 |
| 自由问答 | Principal `提问 [participant] …`，Round 1 后 |
| 交付物 | 结束短节选 + `获取纪要/草案/待决/结论` |
| 稳定性 | 发送重试、网关重连提示、多 Bot 发言、中文 i18n |
| 测试 | Engine 集成测试（Principal 代问、ItemNotes、limit-continue、artifacts） |

详见 [adapters/discord-transport.md](./adapters/discord-transport.md)。

### Web 工作台（v0.2 增量）

| 能力 | 说明 |
|------|------|
| 会议列表与详情 | 概览统计、三栏布局、流程时间线（含驳回步骤） |
| Artifact 浏览 | 文档标签 + Markdown 目录（TOC）、本地化文件标签 |
| 圆桌 Live / 回放 | 围坐席位、Live 高亮、#turn、Strip + Drawer；历史会议回放页 |
| Brief 模板 | 模板列表/编辑、`POST /api/meetings/clone-brief`；Discord `-template` 套用 |
| Principal 画像 | 多 Persona、`USER.md` 结构化编辑（Web） |
| 会议管理 | Workspace 打包下载（`GET /api/meetings/{id}/archive`）、删除 |
| 全站 i18n | `ROUND_TABLE_LOCALE` zh/en（设置页切换） |
| WebChat Transport | WS `/api/chat/ws`，无需 Principal 绑定即可发起会议 |

相关 ADR：[ADR-0013](./architecture/ADR-0013-web-round-table-live-ui.md) · [ADR-0014](./architecture/ADR-0014-meeting-brief-template.md)

---

## P0 — 持久化与可恢复

| 项 | 目标 | 依赖 |
|----|------|------|
| SQLite Event Store | Meeting Event 落库，重启可回放 | ADR-0003 |
| app_settings / ConfigService | 非密钥运行时配置 Web 可改、热生效 | SQLite |
| 会议列表 API | `GET /meetings`、按 channel/principal 查询 | Event Store |
| Discord 历史 | 频道 `获取*` 不依赖内存 `lastByChannel` | API + workspace 索引 |

---

## P1 — Discord 输入态完善（基本完成）

Principal 在 Discord 的唯一交互是**文本**。需要明确的**频道输入态**（Input Phase），让 Principal 随时知道「现在该发什么」。

| 项 | 状态 | 说明 |
|----|------|------|
| 输入态文档 | ✅ | [discord-transport.md §输入态](./adapters/discord-transport.md#输入态-input-phase) |
| `会议状态` / `!rt status` | ✅ | 查询当前频道可接受的指令 |
| 误输入友好提示 | ✅ | 暂停/结束后 hint；setup/确认关解析错误 |
| Discord Typing 指示 | ✅ | LLM 流式期间 `ChannelTyping`，多 Bot 显示对应账号「正在输入」 |
| Slash Commands / 按钮 | 🔲 | 确认关、预设菜单（降低误输入） |

---

## P2 — Web 与非 Discord 入口

| 项 | 说明 | 状态 |
|----|------|------|
| WebChat Transport | WS `/api/chat/ws`，无需 Principal，可发起/进行会议 | ✅ |
| WebChat IM 基线 | 线性聊天窗、头像、尖角气泡、时间戳 | ✅ |
| Web 圆桌 Live 视图 | 围坐席位、Live 高亮/暗沉、#turn、Strip + Drawer | ✅ [ADR-0013](./architecture/ADR-0013-web-round-table-live-ui.md) · [计划](./plans/web-round-table-live-ui.md) |
| Brief 模板 Web | 列表/编辑、历史克隆、开会预填 | ✅ [ADR-0014](./architecture/ADR-0014-meeting-brief-template.md) |
| Artifact 浏览 | Minutes / design-draft 在线阅读 + TOC | ✅ |
| 会议下载/删除 | Workspace zip 下载、会议记录删除 | ✅ |
| Principal 画像 Web | Persona 管理、`USER.md` 编辑 | ✅ |
| Web 全站 i18n | zh/en 切换 | ✅ |
| Web Confirmation UI | Brief 可视化、逐项 ItemNotes **编辑** | 🔲 流程时间线已展示驳回；逐项编辑仍待做 |
| Web 发起会议 parity | Web 侧完整开会向导（对标 Discord） | 🔲 |
| REST/WS | 与 Engine 同源，Transport 可切换 | 部分（Chat WS + 浏览/管理 API） |

---

## P3 — 研讨质量

| 项 | 说明 |
|----|------|
| Deliberation 预设调优 | readiness / synthesis prompt |
| Principal Veto | Consensus 后、Confirmation 前 |
| 确认关长 Brief | Discord 分页/折叠 Item |
| Principal `USER.md` Engine 注入 | ADR-0010 待办；与 Brief 模板独立 |

---

## P4 — 运维与扩展

| 项 | 说明 |
|----|------|
| Event 回放 CLI | 审计与调试 |
| Docker 生产部署 | ✅ [deploy/README.md](../deploy/README.md) |
| Co-principal | D-PR01 |
| Slack Transport | 复用 Principal Port 模式 |

---

## 关联

- [discord-transport.md](./adapters/discord-transport.md)
- [architecture/README.md](./architecture/README.md)
- [plans/web-round-table-live-ui.md](./plans/web-round-table-live-ui.md)
- [releases/v0.2.0.md](./releases/v0.2.0.md)
