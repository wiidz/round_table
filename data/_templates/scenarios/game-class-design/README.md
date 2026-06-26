# 游戏职业研讨场景（deliberation）

## 模式

`meeting_mode: deliberation` — **不投票 OK/KO**，多角色贡献后 Moderator 合成 `artifacts/design-draft.md`。

## 参与者

| ID | 角色 | 视角 |
|----|------|------|
| `designer` | 策划 | 玩法 loop、技能树 |
| `ops` | 运营 | 商业化、活动联动 |
| `player` | 玩家代表 | 爽点、差异化 |
| `tech_lead` | 主程 | 实现成本、同步与平衡 |

## 运行

```bash
make meet-game-class
```

或手动：

```bash
make seed-scenario-game-class
go run ./apps/server/cmd/meet/main.go \
  -mode deliberation \
  -topic "设计新职业「影舞者」的核心技能与定位" \
  -max-rounds 2 \
  -max-free-dialogue-questions 1 \
  -participants "designer:游戏策划:gameplay,ops:运营:monetization,player:玩家代表:experience,tech_lead:主程:engineering"
```

## 成功标准

- 日志：`synthesis: resolved_by=readiness|synthesis|max_rounds`
- `artifacts/design-draft.md` 含各轮贡献汇总
- 辩论轮发言 `stance=none`（无 agree/object 投票）
- `MEETING.md` 显示会议模式为「研讨型」

## 与 3-round-debate 的区别

| | 3-round-debate | game-class-design |
|--|----------------|-------------------|
| mode | decision（默认） | deliberation |
| 目标 | 是否批准方案 | 合成职业设计草案 |
| Stance | agree/object | none（贡献型） |
| 终止 | Consensus / Moderator | SynthesisCompleted |
