# 结论纪要

**会议：** mtg-demo-001（演示数据）

## 已决

1. 批准将用户认证拆为独立 Auth Service，会话采用 JWT + Redis 撤销。
2. 启动 4 周 PoC，验收含安全评审、渗透测试与回滚演练。
3. 全量流量切换在 PoC 评审通过后再排期。

## 行动项

| 负责人 | 事项 | 时限 |
|--------|------|------|
| pragmatist | Auth Service 骨架与双写方案 | 2 周 |
| skeptic | 威胁模型与渗透测试清单 | 2 周 |
| demo (Principal) | PoC 评审会 | 4 周 |
