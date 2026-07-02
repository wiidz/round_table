# 实施计划：Web 圆桌 Live 视图

> **ADR**: [ADR-0013](../architecture/ADR-0013-web-round-table-live-ui.md)  
> **基线回滚点**: `3561e2e`（`feat(web): 浏览器 WebChat Transport 与 IM 聊天窗`）  
> **状态**: Accepted（M0–M5 已完成，2026-06-29）  
> **最后更新**: 2026-06-29

---

## 目标

在浏览器 Chat 页实现 **围坐圆桌 Live 视图**，同时保留 **浓缩历史 + Drawer 全文**；会议 Setup 阶段可继续使用 IM 基线。

**不在本计划内**：Engine 改动、WS 协议扩展、移动端原生 App、3D 桌台。

---

## 里程碑

| 阶段 | 交付 | 验收 |
|------|------|------|
| **M0** | 文档 + ADR（本文） | ADR-0013 Draft，roadmap 更新 | ✅ |
| **M1** | 状态层 + 摘要 + Drawer | turn 序号、condense、Drawer 可读 Markdown | ✅ |
| **M2** | RoundTableStage 静态布局 | 3/5/6 人 roster 席位正确、主持人/我固定方位 | ✅ |
| **M3** | Live 气泡 + 高亮/暗沉 | 新消息切换 activeSpeaker；每席 1 条；#turn 一致 | ✅ |
| **M4** | TranscriptStrip + 模式切换 | 小历史条、点击 Drawer；会议中进行切圆桌 | ✅ |
| **M5** | 降级与 polish | `< md` IM 降级、新消息浮钮、空状态 | ✅ |

---

## 组件结构

```text
apps/web/src/
  components/
    chat/
      chat-window.tsx          # 保留：IM 基线（Setup / 降级）
      chat-composer.tsx        # 从 chat-window 抽出输入区（可选）
    round-table/
      round-table-stage.tsx    # 椭圆席位 + 中心议题 + Live 气泡
      seat-anchor.tsx          # 单席：头像 + 名称 + 可选 #turn 角标
      live-bubble.tsx          # 单条 Live 气泡（高亮/暗沉 props）
      transcript-strip.tsx     # 底部浓缩列表
      transcript-drawer.tsx    # 侧栏全文
      round-table-view.tsx     # Stage + Strip + Drawer + Composer 组合
  hooks/
    use-chat-socket.ts         # 现有 WS
    use-meeting-transcript.ts  # messages → turn、latestBySeat、activeSpeaker
  lib/
    condense-message.ts        # 摘要纯函数
    round-table-layout.ts      # roster → 角度/坐标 %
  pages/
    chat-page.tsx              # 按 phase 切换 IM | RoundTable
  types/
    chat.ts                    # 扩展 ChatMessage.turn?: number
```

---

## 数据模型（前端）

### ChatMessage 扩展

```typescript
export interface ChatMessage {
  id: string
  role: ChatRole
  content: string
  authorId?: string
  authorName?: string
  createdAt: number
  /** 会议发言全局序号；user/system 为 undefined */
  turn?: number
  pending?: boolean
  error?: boolean
}
```

### useMeetingTranscript

输入：`messages: ChatMessage[]`

输出：

```typescript
{
  turns: ChatMessage[]           // 仅含 turn 的消息，按 turn 排序
  latestBySeat: Map<string, ChatMessage>
  activeSpeakerId: string | null
  nextTurn: number               // 下一条 moderator/participant 的序号
}
```

**turn 分配规则**（在 append 时）：

- `role === 'moderator' | 'participant'` → `turn = nextTurn++`
- `role === 'user' | 'system'` → 无 turn

**activeSpeakerId**：最后一次分配 turn 的 `authorId ?? role`。

### Roster 来源（M2–M4）

| 阶段 | 来源 |
|------|------|
| v1 | `GET /participants` 中 `in_roster !== false` 的专家列表 |
| v1.1 | 会议进行中从「会议状态」解析已选 participant ids（若 API 暴露） |
| 降级 | 仅显示消息中出现过的 `authorId` + 固定 moderator + user |

---

## 任务分解

### M1 — 状态层 + Drawer（1–2 天） ✅

- [x] `condense-message.ts` + 单元测试（中英文、Markdown、短文本）— 纯函数已实现，Vitest 待补
- [x] `ChatMessage.turn` 在 `use-chat-socket` append 时赋值
- [x] `use-meeting-transcript.ts`
- [x] `transcript-drawer.tsx`（侧栏全文 + MarkdownDocument）
- [x] 临时挂在 IM 视图：Strip 点行开 Drawer（验证链路）

### M2 — 席位布局（1 天） ✅

- [x] `round-table-layout.ts`：`computeRoundTableSeats` + `participantAngles`
- [x] `seat-anchor.tsx`：绝对定位 `%` + transform
- [x] `round-table-stage.tsx`：中心议题占位、席位高亮/已发言态
- [x] `use-roster-seats.ts`：`GET /participants` roster，消息 author 降级

### M3 — Live 气泡（1–2 天） ✅

- [x] `live-bubble.tsx`：复用 `index.css` chat-bubble；`highlighted` / `dimmed` / `#turn`
- [x] 绑定 `latestBySeat`；200ms transition 切换 active
- [x] `#turn` 角标与 TranscriptStrip 一致；点击 Live 气泡开 Drawer

### M4 — 整合 + 模式切换（1 天） ✅

- [x] `round-table-view.tsx` 组合 Stage + Strip + Drawer（Composer 在 ChatWindow）
- [x] `use-chat-view-mode` + `chat-meeting-phase`：状态回复解析；running/post 自动圆桌
- [x] 手动切换「圆桌 / 列表」；列表模式 `ImTranscriptView` 恢复 IM 气泡
- [x] TranscriptStrip 固定 `h-36`、scroll、新消息浮钮

### M5 — 降级与 polish（0.5–1 天）

- [x] `@media (max-width: 768px)` → 强制 IM 或 Strip-only
- [x] 点击 Strip 行 → 可选高亮对应 Seat
- [x] 空 roster / 仅主持人时的占位 UI
- [x] 更新 `chat-page` 描述文案

---

## 测试

| 类型 | 内容 |
|------|------|
| **Vitest** | `condenseMessage`、`assignTurn`、`computeSeats` |
| **手动** | 开会 → 多专家轮流发言 → turn 连续、高亮切换、Drawer 全文 |
| **回归** | Setup 阶段 IM 仍可用；`git checkout 3561e2e -- apps/web/src/components/chat` 可恢复 |

---

## 风险与缓解

| 风险 | 缓解 |
|------|------|
| 无法从 WS 获知 meeting phase | M4 先用手动切换；后续 `会议状态` 回复解析或 REST phase API |
| 长 Markdown 撑破 Live 气泡 | Live 仅 condense 摘要；全文只在 Drawer |
| StrictMode 双连接 | 已修复于 `3561e2e` 后 hook；圆桌不引入新 WS |
| 领域特化摘要 | 仅结构断句 + 长度，不加 Topic 关键词 |

---

## 完成定义（Definition of Done）

1. 会议进行中，Chat 页展示圆桌 Live：每席 ≤1 气泡，当前发言高亮，其余暗沉，带 `#turn`。
2. 底部 TranscriptStrip 显示全部历史摘要，序号与 Live 一致。
3. 点击 Strip 任一行，Drawer 展示完整 Markdown。
4. `< md` 宽度下降级为 IM 列表，功能不缺失。
5. ADR-0013 状态可升为 Accepted；IM 基线代码仍存在于 repo。

---

## 后续（Phase 2）

| 项 | 状态 | 说明 |
|----|------|------|
| Vitest 单测 | ✅ | `condenseMessage`、`assignTurn`、`computeSeats`、meeting id 解析 |
| Strip 专家筛选 chips | ✅ | TranscriptStrip 按 speaker 过滤 |
| 中心议题 API 绑定 | ✅ | 从状态回复解析 `meeting_id`，`GET /meetings/:id` 拉 topic |
| WS 权威 `turn` 对齐 Engine | ✅ | Hub 会话递增 + 帧 `turn`；客户端优先服务端 |
| 会议回放 turn scrub | ✅ | TranscriptScrubBar + Live/回放投影 |
| ADR 修订 IM 并存策略 | ✅ | ADR-0013 §7 视图并存 |

---

## 后续（非 v1，原 backlog）

---

## 关联提交建议

实现完成后单次或分批 commit，Module `web`，引用 ADR-0013。
