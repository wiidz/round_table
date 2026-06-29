# Discord Transport

Discord 是 RoundTable v0.2 的 **Principal 入口 Transport**：绑定身份、发起会议、确认关、运行期干预与交付物拉取。

实现：`apps/server/cmd/discord/`、`apps/server/internal/adapter/transport/discord/`。

---

## 启动

```bash
# apps/server/.env 或 deploy/.env
DISCORD_BOT_TOKEN=...
DISCORD_BOT_TOKEN_DESIGNER=...   # 可选，多 Bot 发言

make run-discord          # 独立启动 Discord Transport
make server-dev           # 热重载 HTTP；自动拉起 Discord 子进程（discord 包变更时会 rebuild tmp/roundtable-discord）
make stop-discord         # 清理孤儿 Discord 进程
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

发起 `新会议` 后，主持人逐步引导：

| 步骤 | 内容 | 跳过 |
|------|------|------|
| 1/3 目标 | 本场要交付什么 | 发送 `-` |
| 2/3 讨论议题 | 子议题列表（每行一条） | 发送 `-` |
| 3/3 边界与完成标准 | 讨论范围 / 不在范围 / 完成标准 | 发送 `-` |

简报确认后进入预设菜单：

- 研讨型预设 **1–6**，裁决型 **J1–J5**，**0** 进入自定义
- 自定义步骤中 **0** 返回上一级

**简报展示**：Discord 中议题以 `1）2）3）…` 纯文本编号（避免 markdown 列表吞编号）；完整字段写入 `MEETING.md`。

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
| `会议状态` / `状态` | 查看当前频道输入态与可接受指令 |
| `!rt status` | 同上 |

结束时会自动推送各交付物**短节选**（纪要 / 草案 / 待决等，视会议模式而定）；**按需拉取指令仅在最后一条节选末尾出现一次**，避免重复刷屏。

---

## 输入态（Input Phase）

Principal 在 Discord 上**只有文本**一种交互方式。Transport 用 **Input Phase** 描述「当前频道期望 Principal 发什么」。

| Phase | 含义 | 典型可接受输入 |
|-------|------|----------------|
| `idle` | 无进行中的 setup/会议 | `新会议` · `!rt principal bind` |
| `setup_topic` | 等待会议主题 | 主题文字 · `取消会议` |
| `setup_menu` | 等待预设编号 | `1`–`6` / `J1`–`J5` · `0` 自定义 · `取消会议` |
| `setup_custom` | 自定义向导某一步 | 当前步骤菜单数字 · `0` 返回 · `取消会议` |
| `meeting_running` | 会议进行中 | 运行期干预 · 自由问答阶段可 `提问 …` |
| `meeting_paused` | 已暂停 | `恢复会议` · `终止会议` |
| `meeting_free_dialogue` | 自由问答 | `提问 [participant] …` · 运行期干预 |
| `meeting_confirmation` | 确认关 | `批准` / `驳回 …` / `1: …` · 触顶 `1/2/3` |
| `post_meeting` | 已结束 | `获取纪要/草案/待决/结论` · `新会议` |

### 查询当前输入态

| 触发 | 说明 |
|------|------|
| `会议状态` / `状态` / `status` | 自然语言，无需前缀 |
| `!rt status` | 前缀指令 |

返回：**当前 Phase 名称** + **本阶段可接受指令清单**（含 meeting ID，若适用）。

### 误输入提示

- **Setup / 确认关**：无效选项会返回解析错误（含正确格式说明）
- **暂停 / 会议结束后**：Principal 发送无法识别的文本时，Bot 返回简短 hint，并提示发送 `会议状态`

实现：`input_state.go` 中 `MeetRunner.InputPhase` 聚合 setup 会话、活跃会议、`ChannelPrincipal`（确认/暂停/自由问答）与 `lastByChannel`。

### 正在输入（Typing Indicator）

Discord 原生 API：`POST /channels/{channel_id}/typing`（discordgo：`Session.ChannelTyping`）。客户端显示 **「Bot 名称 正在输入…」**，单次约 10 秒；LLM 流式期间每 **7 秒** 自动续期。

| 角色 | Bot | 频道表现 |
|------|-----|----------|
| Participant（有独立 Bot） | `DISCORD_BOT_TOKEN_<ID>` | 该 Bot 名称「正在输入」，**不再**发 `🎤 策划 · 研讨` 头行 |
| Participant（无独立 Bot） | 主 Bot | 主 Bot「正在输入」+ 保留 `🎤 **策划** · …` 头行（名称与 Bot 账号不一致时用头行标识说话者） |
| Moderator（合成/就绪等） | 主 Bot | 主 Bot「正在输入」 |

配置独立 Bot：`.env` 中 `DISCORD_BOT_TOKEN_DESIGNER` 等，见 `server.yaml` → `participant_bots`。

---

## 频道行为

| 内容 | 推送方式 |
|------|----------|
| 进度（轮次开始/结束、自由问答开始、轮到谁回答） | 主 Bot |
| Participant / Moderator LLM 发言 | 各 Participant Bot（或主 Bot 回退） |
| 自由问答 Q&A 正文 | Participant stream（不重复 progress 正文） |
| Principal `提问` ack | 主 Bot |
| 确认关 Brief | 主 Bot |
| 结束交付物 | 主 Bot（各 artifact 短节选；**拉取指令仅在最后一条末尾**） |
| LLM 流式生成 | 对应 Bot 触发 Discord「正在输入」（`ChannelTyping`） |

网络：发送失败自动重试 3 次；网关重连后向活跃会议频道发送恢复提示。

---

## 关联

- [Principal](../domain/principal.md)
- [Confirmation](../domain/confirmation.md)
- [ADR-0004](../architecture/ADR-0004-principal-confirmation.md)
- [ADR-0011](../architecture/ADR-0011-meeting-mode.md) — Meeting Mode、Principal 自由对话代问
- [apps/server/README.md](../../apps/server/README.md) — 本地开发与配置
