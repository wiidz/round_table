# Pre-meeting (Round 0)

## skeptic (Security Architect)

独立视角：拆分可降低单体攻击面，但 JWT 撤销与密钥轮换必须集中治理；Redis 单点需高可用。建议 PoC 阶段明确威胁模型与审计日志。

## pragmatist (Tech Lead)

独立视角：团队已有 OAuth 经验，拆分可并行开发；首阶段仅迁移登录与会话，授权仍走现有网关。关注迁移窗口与双写期复杂度。
