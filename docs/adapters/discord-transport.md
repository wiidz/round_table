# Discord Transport

Discord 是 RoundTable v0.2 的 **Principal 入口 Transport**：绑定身份、发起会议、确认关、运行期干预与交付物拉取。

实现：`apps/server/cmd/discord/`、`apps/server/internal/adapter/transport/discord/`。

---

## 启动

```bash
# apps/server/.env
DISCORD_BOT_TOKEN=...
DISCORD_BOT_TOKEN_DESIGNER=...   # 可选，多 Bot 发言

make run-discord
```

配置：`apps/server/configs/server.yaml` → `transport.discord`（locale、预设、participant_bots 等）。

Principal 绑定持久化：`data/transport/discord-principal.json`。

---

## 指令一览

### 前缀指令（默认 `!rt`）

| 指令 | 说明 |
|------|------|
| `!rt help` | 帮助 |
| `!rt principal bind` | 绑定本服务器/私信 Principal |
| `!rt principal whoami` | 查看绑定 |
| `!rt principal unbind` | 解除绑定 |
| `!rt meet [topic]` | 带主题发起（仍走主持人引导） |

### 自然语言（无需前缀）

| 触发 | 说明 |
|------|------|
| `新会议` / `开始会议` / `会议开始` | 发起会议，主持人逐步引导 |
| `取消会议` | 取消待确认的会议配置 |

### 会议配置（Setup 阶段）

- 研讨型预设 **1–6**，裁决型 **J1–J5**，**0** 进入自定义
- 自定义步骤中 **0** 返回上一级

### 确认关（`confirmation_mode: required`）

| 指令 | 说明 |
|------|------|
| `批准` / `1` | 通过并归档 |
| `驳回 …` / `2` | 追加 1 轮研讨（可附整体意见） |
| `1: …  2: …` | 逐项附注（ItemNotes） |
| 触顶后 `1`/`2`/`3` | 强制批准 / 继续研讨 / 中止（ADR-0004 §6） |

### 运行期干预（Turn boundary）

| 指令 | 说明 |
|------|------|
| `暂停会议` | 当前发言结束后暂停 |
| `恢复会议` | 暂停后恢复 |
| `终止会议` | 中止并输出部分纪要 |
| `立即合成` | 研讨型：强制进入草案合成 |
| `强制共识` | 裁决型：强制宣布共识 |

### 自由问答（Round 1 后，预设含 free dialogue 时）

| 指令 | 说明 |
|------|------|
| `提问 …` | Principal 代问（经 Domain Moderator relay，频道仅 ack + Participant 回答） |
| `提问 designer …` | 指定回答者 |

### 会议结束后

| 指令 | 说明 |
|------|------|
| `获取纪要` | 推送 `MINUTES.md` 节选 |
| `获取草案` | 研讨型：`design-draft.md` |
| `获取待决` | 研讨型：`open-questions.md` |
| `获取结论` | 裁决型：`artifacts/minutes.md` |

结束时会自动推送短节选，并提示上述按需拉取指令。

---

## 频道行为

| 内容 | 推送方式 |
|------|----------|
| 进度（轮次开始/结束、自由问答开始、轮到谁回答） | 主 Bot |
| Participant / Moderator LLM 发言 | 各 Participant Bot（或主 Bot 回退） |
| 自由问答 Q&A 正文 | Participant stream（不重复 progress 正文） |
| Principal `提问` ack | 主 Bot |
| 确认关 Brief | 主 Bot |
| 结束交付物 | 主 Bot（短节选 + 按需拉取） |

网络：发送失败自动重试 3 次；网关重连后向活跃会议频道发送恢复提示。

---

## 关联

- [Principal](../domain/principal.md)
- [Confirmation](../domain/confirmation.md)
- [ADR-0004](../architecture/ADR-0004-principal-confirmation.md)
- [ADR-0011](../architecture/ADR-0011-meeting-mode.md) — Meeting Mode、Principal 自由对话代问
- [apps/server/README.md](../../apps/server/README.md) — 本地开发与配置
