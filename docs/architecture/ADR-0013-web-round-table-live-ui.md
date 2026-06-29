# ADR-0013: Web 圆桌 Live 视图（围坐发言 + Drawer 历史）

**状态**: Draft  
**日期**: 2026-06-29  
**关联**: [CONSTITUTION.md](../CONSTITUTION.md), [ADR-0012-transport-reception-agent.md](./ADR-0012-transport-reception-agent.md), commit `3561e2e`（WebChat IM 基线）

---

## 背景

RoundTable 产品名与核心隐喻是 **圆桌会议**：多 Agent 围坐讨论，Moderator 调度，Principal 在场决策。

浏览器 WebChat Transport（ADR 前置实现，commit `3561e2e`）已具备：

- WebSocket 入站/出站，无需 Principal 绑定即可发起与进行会议；
- 出站帧携带 `role`、`author_id`、`author_name`、`at`；
- 前端线性 IM 聊天窗（头像、尖角气泡、时间戳、可滚动历史）。

Principal 反馈：线性堆叠聊天窗 **不像「几个人在桌上开会」**；希望：

1. 气泡从 **各参与者座位** 伸出，而非共用一列时间轴；
2. **保留完整聊天历史**，但默认 UI 轻量（小窗口 + 摘要）；
3. 点击摘要 → **Drawer** 展开 Markdown 全文；
4. Live 区每人仅 **最新 1 条** 气泡；**当前发言者高亮**，其余 **暗沉**；
5. 全局 **严格递增发言序号**（#1、#2、#3…），标明发言顺序。

本 ADR 仅约束 **Web 前端呈现层** 与 **客户端状态模型**；不修改 Engine、合成 pipeline 或 Transport 协议（v1 不新增 WS 字段）。

---

## 决策

### 1. 视图模式：双区 + 模式切换

| 区域 | 职责 | 默认可见性 |
|------|------|------------|
| **RoundTableStage**（上） | 席位布局、Live 气泡、议题/轮次、当前发言高亮 | 会议进行中为主 |
| **TranscriptStrip**（下） | 浓缩历史列表、发言序号、点击开 Drawer | 始终（高度受限） |
| **TranscriptDrawer**（侧） | 单条消息 Markdown 全文 | 按需 |
| **ChatComposer**（底） | 输入框（复用现有） | 始终 |

**模式规则**：

| 会话阶段 | 主视图 |
|----------|--------|
| 空闲 / Setup / Reception 确认 | **IM 基线**（`3561e2e` 线性聊天窗）或简化版 Stage（仅司仪 + 我） |
| 会议进行中（`InputPhaseMeeting*`） | **圆桌 Live** |
| 会议结束（`InputPhasePostMeeting`） | 圆桌最后一帧 + TranscriptStrip；Drawer 便于拉交付物 |

窄屏（`< md`）降级：Stage 缩略或隐藏，TranscriptStrip + Drawer 为主（见 §5）。

### 2. Live 区语义（核心交互）

**数据**：客户端维护 `messages: ChatMessage[]`（全量历史，不变）；由 WS  append。

每条 **非系统、非用户** 的会议发言（`moderator` / `participant`）在 append 时分配：

```text
turn: number   // 全局严格递增，从 1 开始；用户消息、系统消息不占 turn（或占独立序列，见计划）
speakerId: string  // authorId ?? role 映射
```

**Live 投影**（派生状态，不另存服务端）：

```text
latestBySeat: Record<speakerId, ChatMessage & { turn }>
activeSpeakerId: speakerId | null   // 最近一次占 turn 的发言者
```

**渲染规则**：

1. 每个 roster 席位最多 **1 条** Live 气泡 = `latestBySeat[seatId]`（无则空）。
2. `activeSpeakerId === seatId` → 该席 **高亮**（opacity 1、ring、可选轻微 scale）。
3. 其他有内容的席位 → **暗沉**（opacity ~0.4–0.5，去饱和）。
4. 气泡角标显示 **`#turn`**，与 TranscriptStrip 中序号 **一致**。
5. 新 turn 到达 → `activeSpeakerId` 切换；旧席气泡保留但变暗。

**用户（Principal）**：固定席位（建议 6 点钟方向）；用户消息不占 `turn`，不参与「专家轮次」序号，但可在 Strip 中显示「我」标签。

**司仪（Moderator）**：独立席位（建议 12 点钟）；占 turn，与专家同等规则。

### 3. TranscriptStrip（浓缩历史）

- 固定高度 **120–160px**（约 2–3 条），`overflow-y: auto`。
- 每行：`#turn` · 显示名 · `HH:mm` · **摘要**（见 §4）。
- 过长内容：`line-clamp-1` 或首句截取；右侧「详情」或整行可点。
- 新消息：默认滚到底；用户上翻时显示「↓ 新消息」浮钮，不强制 scroll。
- 点击行 → 打开 **TranscriptDrawer**，展示该条完整 Markdown + 元信息。

### 4. 摘要策略（客户端纯函数）

`condenseMessage(content: string): string`

1. 去 Markdown 装饰（`#`、`**`、列表符）仅用于摘要，不修改原文。
2. 取 **第一句**（`。！？\n` 或 `.!?` 断句）或 **前 60 个字符**（Unicode 码点），取较短展示需求者。
3. 若全文 ≤ 80 字 → 不截断，不显示「展开」。
4. 进度类/系统短句：原样或 `[系统]` 前缀。

**禁止**：按 Topic/workspace 关键词做领域特化摘要（遵守 Constitution 通用 pipeline 原则）。

### 5. 席位布局

- 根据 **当前会议 roster**（来自配置 `meet_participants` 或 setup 已选 subset）动态排布。
- 3–8 人：椭圆/圆形等分角度；**Principal** 固定底部中心；**Moderator** 固定顶部中心。
- 气泡复用现有 `.chat-bubble` 尖角样式，尾巴指向 **圆心或头像**（与 IM 基线一致）。
- 中心：**议题标题** + 可选 meeting id / 轮次（只读，来自 status 或首条司仪进度）。

**不采用** React Flow 作为 v1 布局引擎（见「拒绝的选项」）。

### 6. 协议与后端

**v1 不修改 WebSocket Frame**；`turn` 与 `latestBySeat` 纯前端从消息流推导。

可选 **v2**（非本 ADR 范围）：服务端推送 `turn` 字段，与 Engine Event 序号对齐，便于回放一致。

### 7. 回滚策略

- IM 基线已提交：`3561e2e`。
- 实现时保留 `ChatWindow`（或 `ChatViewMode = 'im' | 'roundtable'`），Feature flag / 输入态切换可一键回 IM。
- 圆桌组件独立目录 `apps/web/src/components/round-table/`，不删除 IM 代码直至 Live 稳定。

---

## 拒绝的选项

| 选项 | 原因 |
|------|------|
| **纯 React Flow 圆桌** | 过重；v1 只需固定角度布局 + CSS，Flow 留作关系图/编排扩展 |
| **每席 Live 堆叠多条** | 与「围坐感」冲突；Principal 明确要求 Live 仅 1 条 |
| **条内展开全文** | Principal 倾向 **Drawer**；长 Markdown / deliberation 侧栏更合适 |
| **去掉完整历史** | Principal 要求保留；Strip + Drawer 满足「轻 + 可查」 |
| **服务端 turn 字段（v1）** | 增加协议与 Engine 耦合；客户端递增可满足 Live/Strip 一致序号 |
| **Discard IM 基线** | 需可回滚；Setup/窄屏仍需要线性聊天 |

---

## 后果

### 正面

- Web 体验与产品名、Meeting 隐喻一致，差异化明显。
- Transport 协议不变，Engine 零改动，迭代风险可控。
- Drawer + Strip 兼顾「氛围」与「查纪要」。

### 负面 / 风险

- 多席位 + 动画对 **窄屏、无障碍** 不友好 → 必须提供 IM 降级。
- 前端 turn 与 Engine Event 序号可能 **短暂不一致**（仅 UI）；v2 可对齐。
- Roster 动态变化（setup 中途改人）需重新计算席位；需测试边界。

### 待实现

见 [plans/web-round-table-live-ui.md](../plans/web-round-table-live-ui.md)。

---

## 参考

- `apps/web/src/pages/chat-page.tsx`
- `apps/web/src/components/chat/chat-window.tsx`
- `apps/web/src/hooks/use-chat-socket.ts`
- `apps/server/internal/adapter/transport/web/message.go`
- commit `3561e2e` — WebChat IM 基线
