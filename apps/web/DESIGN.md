---
version: 1.0
name: RoundTable Web — Principal UI
description: Multi-Agent Meeting Engine 的 Web 入口。Principal 在此审阅 Confirmation、阅读 Minutes 与 Artifacts。设计语言继承 CRM Console 的克制与精确，并针对「会议阅读 + 决策确认」场景裁剪；支持亮色 / 暗色双主题。
---

# RoundTable Web 1.0 · Principal UI

## North Star

RoundTable 不是聊天窗口。

Principal 来到 Web 端，是为了：

* **审阅** Confirmation Brief
* **批准或驳回** 会议结论
* **阅读** Minutes 与 Artifacts
* **追溯** 多轮讨论与共识过程

软件负责整理会议产出，Principal 负责最终决策。

---

## Philosophy

Principal 在 Web 端主要有两类工作：

### Review Work — 阅读与决策

Confirmation Brief、Minutes、Consensus、Participant 发言摘要、Artifacts。

目标：快速理解、逐项确认、明确驳回理由。

### Browse Work — 浏览与检索

会议列表、按状态筛选、打开历史 Meeting、下载 workspace 文件。

目标：低密度操作、高可读性；**不是** CRM 式高密度表格后台。

> AI 发言、Moderator 总结、Suggested Action 是**会议内容的一部分**，用 AI 紫统一标识，而非独立「助手页」。

---

# Design Principles

## 1. Meeting First

一切 UI 围绕 **Meeting** 组织：Topic → Rounds → Consensus → Confirmation → Artifacts。

不要做成通用 Admin Dashboard。

---

## 2. Decision First

Confirmation 页中，**待确认项 / 请示 / 驳回入口** 必须比元数据（时间、Channel ID）更醒目。

Principal 的操作（批准、驳回、补充 ItemNotes）使用 **品牌色（Principal Orange）**。

---

## 3. Calm by Default

默认安静：低饱和、留白、清晰层级。

强调色克制；状态与 AI 产出需要时才出现颜色。

---

## 4. Content is the Interface

Minutes、Artifact 正文采用 **Document Layout**，而非 Form Layout。

最大阅读宽度 720–760px，Body 15px，line-height 1.85。

---

## 5. Trust & Transparency

AI / Participant 产出必须可识别：

* AI 紫 + ✦ 标识
* 标注来源（Participant 名称、Round）
* 不冒充 Principal 人工输入

---

## 6. Dual Theme（亮色 / 暗色）

**必须支持** Light 与 Dark，且语义 token 一一对应。

* 切换方式：Header 主题按钮；偏好写入 `localStorage`（`roundtable-theme`）
* 首屏：`index.html` 内联脚本防闪烁（FOUC）
* 默认：跟随 `prefers-color-scheme`，用户选择后覆盖

实现：`src/styles/theme.css` + `html.dark` class。

---

# Foundations

## Color

### Strategy

Neutral-first。彩色只用于：Principal 操作、AI 产出、语义状态。

---

### Neutral（Light / Dark）

| Token | Light | Dark |
|---|---|---|
| canvas | #FBFBFC | #0B0C0E |
| surface | #FFFFFF | #16181D |
| surface-raised | #FFFFFF | #1C1F26 |
| border-subtle | #EDEDF0 | #23262E |
| border | #E2E2E6 | #2C2F38 |
| text-primary | #16181D | #F4F5F7 |
| text-secondary | #5B5F66 | #A6ABB5 |
| text-tertiary | #8A8F98 | #6E747F |

Tailwind 用法：`bg-canvas`、`bg-surface`、`text-text-secondary`、`border-border-subtle`。

---

### Brand — Principal Orange

Primary `#E85D04`

用途：Principal 主操作——批准、驳回、提交 ItemNotes、关键 CTA。

Soft 背景：`rgba(232, 93, 4, 0.10)`（暗色约 15%）

**禁止**：大面积铺满；与 AI 紫混用于同一控件。

---

### Intelligence — AI Accent

AI `#6E56F8`

Gradient `linear-gradient(135deg, #6E56F8 0%, #9B6CF6 100%)`

用途：Participant 发言、Moderator 总结、AI 摘要、流式生成态。

Soft 背景：`rgba(110, 86, 248, 0.08)`（暗色略加深）

> **铁律**：品牌橙 = Principal 人的操作；AI 紫 = 机器产出。两者不混用。

---

### Semantic

| 含义 | 色值 | 用途 |
|---|---|---|
| Success | #2FB67C | 已批准、共识达成 |
| Info | #3DA8F5 | 提示、链接 |
| Warning | #F5A623 | 待确认、进行中 |
| Danger | #F2545B | 驳回、终止、错误 |

呈现：文字 + Soft 背景（约 10% 透明度），禁止大色块。

---

## Typography

字体：Inter / SF Pro / 苹方，`tabular-nums` 用于时间戳。

| 角色 | Size | Weight | 用途 |
|---|---|---|---|
| Display | 28px | 600–700 | Artifact / Minutes 主标题 |
| Page Title | 22px | 600 | 页面标题 |
| Section Title | 16px | 600 | Brief 区块 |
| Body | 15px | 400 | 阅读正文 |
| UI Text | 14px | 400 | 按钮、导航 |
| Meta | 13px | 400 | 时间、Meeting ID |
| Label | 12px | 500 | 字段标签 |

---

## Radius & Spacing

Radius：sm 8 · md 12 · lg 16 · xl 20 · 2xl 28

Spacing：8pt 栅格；section-gap 32px

---

## Elevation

| Level | 用途 |
|---|---|
| flat | 嵌入区块 |
| raised | 卡片、Brief 条目 |
| overlay | Dialog、Sheet |
| ai-glow | AI 流式生成态 |

暗色主题阴影减弱白边，提高深度对比而非描边。

---

## Motion

| 场景 | 时长 |
|---|---|
| Hover | 120ms |
| 展开 | 200ms |
| Dialog | 280ms |

尊重 `prefers-reduced-motion`。

---

## Iconography

Lucide，1.5px 线宽。AI 相关：Sparkles / Wand + AI 紫。

---

# Workspace Types

## Meeting List（Browse）

Layout：PageHeader · 筛选 · 会议卡片列表

优先级：Topic > 状态 > 时间 > Channel

规则：卡片式列表，**不用 AG Grid**（初版）。

---

## Confirmation（Review + Decision）

Layout：Brief 摘要 · 逐项 Agenda Item · ItemNotes 输入 · 批准 / 驳回

### Confirmation Item Block

* 默认：flat 列表 + 间距
* 待决策：左侧 **品牌橙** 强调线 + brand-soft 背景
* 已批准：Success 绿强调线
* 已驳回：Danger 红强调线

Principal 操作区固定在条目底部或页脚 sticky bar。

---

## Artifact / Minutes（Review）

Layout：Document 单列 · 可选目录锚点 · AI Summary（若有）

### AI Summary Block

AI 紫 soft 背景 + ✦ + 「AI 生成」标签。

### Follow Content

最大宽度 720–760px，Document Layout，禁止字段 Grid 堆砌。

---

# Components

## Button

| 变体 | 用途 |
|---|---|
| Primary（橙） | 批准、确认、提交 |
| Destructive | 驳回、终止 |
| Outline | 次级操作 |
| Ghost | 导航、主题切换 |
| AI（紫渐变） | 查看 AI 摘要（只读类） |

---

## Badge

状态：Running / Confirmation / Completed / Aborted

AI 标签：「✦ AI」紫色 Soft Pill

---

## Theme Toggle

Header 右侧 Sun / Moon 图标按钮；状态持久化。

---

# Do

* 支持 Light / Dark，用 CSS 变量而非硬编码色值
* Principal 操作用橙，AI 产出用紫
* Confirmation Decision First
* Minutes 阅读优先、Document Layout
* 与 Discord Transport 语义一致（同一领域词汇）

---

# Don't

* 不要引入 CRM 式 AG Grid / 重型 Dashboard（初版）
* 不要让 Channel ID 比 Brief 内容更醒目
* 不要 AI 紫与品牌橙混在同一按钮
* 不要仅支持单主题
* 不要把 Meeting Engine 业务逻辑写进前端

---

# Golden Rule

页面目标是「列出会议」→ Meeting List。

页面目标是「阅读并做决定」→ Confirmation / Artifact Review。

Principal 橙管「人的决定」，AI 紫管「机器的产出」—— 这是 RoundTable Web 1.0 的色彩宪法。

---

# High-End Flat（Profile 管理页）

Principal / Participant 列表与档案编辑采用 CRM Console 4.0 **High-End Flat** 变体：

| 元素 | 规范 |
|------|------|
| 页面壳 | `bg-surface` + `ring-[var(--panel-shell-ring)]`，无灰边；圆角 `1.75rem` |
| 字段编辑区 | `bg-canvas` 嵌入底 + `--field-inset-shadow` + inset ring |
| **Input / Textarea（可编辑）** | `bg-surface` + `rounded-xs` + inset ring + field inset shadow |
| **Input / Textarea（只读 / disabled）** | `bg-black/5%` 灰底 + `rounded-lg`（大于可编辑）+ inset ring `black/6%`；文字 `text-tertiary` |
| **委托人档案偏好区** | 白底 `SideTabWorkspacePanel`，**不要**再套 `heFormEmbed` 灰底卡片 |
| **字段说明 ?** | 控件右侧 `FieldHintPopover`；hover / focus 显示 tooltip；设置表单用 `SettingsFieldRow` 的 `hint` |
| **表单 label 列宽** | 左列固定约 **8 个汉字**（`6rem` / `heFormFieldRowGrid`） |
| **表单 label 字间距** | `tracking-[0.06em]`（`heFieldLabel` / `briefFieldCaptionClass`） |
| 交互 focus | **品牌橙** ring（人的编辑操作） |
| Principal Eyebrow | 品牌橙 pill · `Decision Owner` |
| Participant Eyebrow | AI 紫 pill · `Expert Profile`（标识专家身份，非表单 focus） |
| 文件 Pill | 选中：品牌橙 soft + inset ring；未选中：浅灰底 |
| 动效 | `500ms` 弹簧曲线；尊重 `prefers-reduced-motion` |

实现：`src/lib/highend-styles.ts`（`heInputEditable` / `heInputReadonly` / `heFormFieldRowGrid`）、`src/components/ui/input.tsx`、`src/components/ui/textarea.tsx`、`src/components/settings/field-hint-popover.tsx`

---

# Implementation

| 文件 | 说明 |
|---|---|
| `src/styles/theme.css` | Light / Dark CSS 变量 |
| `src/index.css` | Tailwind v4 + `@theme inline` 映射 |
| `src/hooks/use-theme.ts` | 主题切换与持久化 |
| `src/lib/highend-styles.ts` | High-End Flat 类名 token（含 `heInputEditable` / `heInputReadonly`） |
| `src/components/ui/input.tsx` | 全站单行输入；自动只读/可编辑样式；可选 `hint` |
| `src/components/ui/textarea.tsx` | 全站多行输入；同上 |
| `src/components/settings/field-hint-popover.tsx` | `SettingsFieldRow` + 字段右侧 `?` tooltip |
| `src/components/profile/` | Profile 列表/页头组件 |
| `index.html` | 首屏防 FOUC |

与 CRM Console 4.0 的关系：继承 Neutral-first、双主题、橙/紫分工；裁剪 Data Workspace 高密度表格与 CRM 侧栏壳层。
