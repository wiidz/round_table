# 3 轮辩论场景

## 流程

```
Round 0 (Pre-meeting)  各 Participant 独立提交初始观点（互不可见）
        ↓ 汇总
Round 1                辩论
        ↓
Free dialogue          Round 1 后互相提问/回答（默认每人 1 问，可配置）
        ↓
Moderator 总结 Round 1
        ↓
Round 2                辩论 + Moderator 总结
        ↓
Round 3                辩论 → 共识
```

## 设计目标

| 轮次 | skeptic | pragmatist | 结果 |
|------|---------|------------|------|
| 0 | 独立视角 | 独立视角 | 汇总 → Round 1 |
| 1 | object（安全缺口） | agree | 未共识 → 自由对话 → Moderator 总结 → 第 2 轮 |
| 2 | object（审计/回滚） | agree | 未共识 → Moderator 总结 → 第 3 轮 |
| 3 | agree | agree | 共识达成 → Completed |

`max-rounds=3` 指辩论轮 1–3，不含 Round 0。

关闭 Round 1 后自由对话：

```bash
-max-free-dialogue-questions 0
# 或 server.yaml: free_dialogue_max_questions: 0
# 或 ROUND_TABLE_FREE_DIALOGUE_MAX_QUESTIONS=0
```

## 议题

将 monolith 用户模块拆为独立 Auth Service，采用 JWT + Redis 会话撤销，是否批准进入开发？

## 参与者

| ID | 角色 | 立场 |
|----|------|------|
| `skeptic` | Security Architect | R0 独立视角；R1–2 object；R3 agree |
| `pragmatist` | Tech Lead | R0 独立视角；R1–3 agree |

## 运行

```bash
make meet-3round
```

## 成功标准

- 日志：`debate_rounds=3`（`max-rounds` 不含 Round 0），`pre_meeting=1`
- `pre-meeting/perspectives.md` 含两人独立观点
- `free-dialogue/after-round-001.md` 含 Round 1 后 Q&A
- `moderator/round-001-summary.md`、`round-002-summary.md` 存在
- 第 2、3 轮发言引用 Pre-meeting 与 Moderator 总结

Role 含空格时，`-participants` 的值必须用引号包住。
