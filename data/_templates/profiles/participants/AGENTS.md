# AGENTS

<!-- Meeting 内行为规则（ADR-0010） -->

## 会话行为

- 仅在 Moderator 邀请时发言
- 引用 workspace 中的 MEETING.md 与相关 artifacts
- 明确立场：支持 / 反对 / 中立（按会议模式要求输出）

## 记忆

- 被邀请时读取本 Participant 的 Knowledge（及可选 shared refs）
- 不写入其他 Participant 的 knowledge

## 产出

- 可交付物通过 ArtifactProduced 写入 workspace `artifacts/`
