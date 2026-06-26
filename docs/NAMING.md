# 命名约定

> 命名决定设计。

RoundTable 用**会议隐喻**组织多智能体协作。错误的命名会把系统拉回「单主 Agent + 子任务分发」的老路。

权威来源：[CONSTITUTION.md](./CONSTITUTION.md) § Naming Rules

---

## 统一称谓

| 不要说 | 统一叫 | 说明 |
|--------|--------|------|
| Main Agent / Boss Agent | **Moderator** | 司仪，负责调度，不是专家 |
| Sub Agent | **Participant** | 领域专家，只响应邀请 |
| Task / Workflow | **Meeting** | 最高层抽象，一切从 Meeting 出发 |
| Chat History | **Minutes** | 结构化纪要，不是原始聊天记录 |
| Memory | **Knowledge** | 跨 Meeting 的持久知识 |
| — | **Principal** | 委托人，发起 Meeting、验收结论；全项目唯一称谓 |

---

## 为什么这样命名

### Main Agent → Moderator

如果一直叫 Main Agent，实现 inevitably 会滑向 OpenClaw / LangGraph 式的「主 Agent 分发子任务」模型。  
Moderator 强调的是**控场**，不是**执行**。

### Sub Agent → Participant

Participant 是会议中的**专家角色**，有明确的 Role 边界，不是可被任意调度的子进程。

### Task → Meeting

复杂问题不是「任务队列里的一项」，而是一场需要讨论、辩论、达成共识的**会议**。

### Principal

Principal 强调**决策归属**——Meeting 因 Principal 的问题而召开，结论须符合 Principal 的预期。  
Domain、Auth、UI 全栈统一使用 Principal。

---

## 完整对照表

| 使用 | 避免 |
|------|------|
| Meeting | Task, Workflow |
| Principal | Main Agent, Boss Agent |
| Moderator | Main Agent, Boss Agent, TaskManager, WorkflowManager |
| Participant | Sub Agent, Worker Agent |
| Round | Step, Node |
| Agenda | Prompt chain |
| Consensus | Auto-complete |
| Confirmation | Manual approval（无 Brief 结构时） |
| Minutes | Chat History, Log |
| Artifact | Output file（无 Meeting 语义时） |
| Knowledge | Memory, Context window |
| Event | Implicit state mutation |

---

## 关联

- [CONSTITUTION.md](./CONSTITUTION.md)
- [PRINCIPLES.md](./PRINCIPLES.md)
- [domain/README.md](./domain/README.md)
