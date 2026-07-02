# RoundTable 演示数据

脱敏的静态示例，供 `make seed-demo` 复制到运行时 `data/` 目录。**不含真实 Token、客户议题或 Discord 日志。**

## 内容

| 路径 | 说明 |
|------|------|
| `workspaces/mtg-demo-001/` | 已结束的裁决型示例会议（Auth Service 拆分议题） |
| `profiles/principals/demo/USER.md` | 演示用委托人档案 |
| `briefs/decision-review/BRIEF.yaml` | 与 `_templates/briefs` 相同的简报模板副本 |

专家档案 `skeptic` / `pragmatist` 在 seed 时从 `scenarios/3-round-debate` 复制。

## 用途

- 开源用户 **无需 API Key** 即可浏览 Web 会议详情、流程、文档目录
- 文档截图与 CI 快照测试的固定样本（未来可扩展）

## 注意

`data/workspaces/` 为 gitignore；修改演示内容请编辑本目录后重新 `make seed-demo`。
