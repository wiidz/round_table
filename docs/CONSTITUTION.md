# RoundTable Constitution v0.2

> Build AI Teams, not AI Agents.

---

# Project Identity

RoundTable is a **Multi-Agent Meeting Engine**.

It is **NOT** another AI Agent.

It is **NOT** an Agent Runtime.

It is **NOT** a Workflow Engine.

It is a collaborative reasoning platform where multiple AI participants discuss, debate, and reach consensus under the control of a Moderator, on behalf of a Principal.

The goal is not autonomous execution.

The goal is collaborative decision making.

---

# Vision

One problem.

Many minds.

One decision.

RoundTable models how expert teams solve complex problems.

Every complex task is treated as a structured meeting.

---

# Design Philosophy

Everything is a Meeting.

Everything else is built around this concept.

The Meeting is the highest-level abstraction in the system.

Agents are implementation details.

---

# Core Concepts

The following concepts are considered part of the core domain.

They should not be renamed or replaced without an architecture discussion.

## Meeting

Represents an entire discussion around a topic.

Contains:

* Topic
* Agenda
* Principal
* Moderator
* Participants
* Rounds
* Consensus
* Confirmation
* Minutes
* Artifacts
* Action Items

---

## Principal

The Principal is NOT an AI expert.

The Principal is NOT the Moderator.

The Principal is the **decision owner** — the person who initiates the Meeting and accepts or rejects the outcome.

Responsibilities:

* defines Topic and Agenda
* creates the Meeting
* reviews Confirmation Brief
* approves or rejects the final conclusion
* holds ultimate control (veto, force consensus, pause, abort)

The Principal owns the problem.

The Principal does not provide domain expertise.

One Meeting has exactly one Principal.

In the domain layer and everywhere else in RoundTable, use **Principal** only.

See [principal.md](./domain/principal.md).

---

## Moderator

The Moderator is NOT an AI expert.

The Moderator is a scheduler.

Responsibilities:

* controls speaking order
* distributes context
* summarizes discussions
* detects consensus
* starts and ends rounds
* assigns participants
* manages meeting state
* prepares confirmation brief
* presents confirmation to the Principal

The Moderator owns orchestration.

The Moderator does not own expertise.

---

## Participant

Participants are domain experts.

Examples:

* Designer
* Programmer
* Researcher
* Architect
* Tester
* Artist
* Balance Designer

Participants never schedule themselves.

Participants only respond when invited by the Moderator.

Participants never communicate directly.

Participants never communicate directly with the Principal.

All communication passes through the Moderator.

---

## Round

A Meeting consists of multiple Rounds.

Each Round contains:

* speaking order
* participant responses
* moderator summary
* updated meeting state

---

## Agenda

Represents the discussion objective.

A Meeting may contain one or multiple Agenda items.

---

## Opinion

A participant's current viewpoint.

Opinions may evolve during the Meeting.

---

## Consensus

Represents whether sufficient agreement has been reached.

Consensus is a Meeting property.

Not a Participant property.

---

## Confirmation

The optional gate before a Meeting produces its final conclusion.

After Consensus among Participants, the Moderator prepares a **Confirmation Brief** — numbered items for the Principal to review.

The Principal approves or rejects. Rejection returns the Meeting to discussion.

Confirmation can be skipped via configuration (`confirmation_mode: skip`).

Consensus answers: do the experts agree?

Confirmation answers: does the Principal accept?

See [confirmation.md](./domain/confirmation.md).

---

## Minutes

The structured summary of a Meeting.

Minutes are not chat history.

Minutes are curated knowledge.

---

## Knowledge

Persistent information shared between Meetings.

Knowledge is long-term.

Minutes are Meeting-specific.

---

## Artifact

An output produced by a Meeting.

Examples:

* Markdown document
* Design proposal
* Code
* Architecture
* TODO list

---

## Action Item

Concrete follow-up work produced by the Meeting.

---

# Architecture Principles

Meeting First.

Not Agent First.

RoundTable models discussions rather than conversations.

Discussion drives the architecture.

Chat is only the interface layer.

---

# Communication Model

All communication flows through the Moderator.

Correct:

Principal

↓

Moderator

↓

Participant

↓

Moderator

↓

Participant

↓

Moderator

↓

Consensus

↓

Confirmation（optional）

↓

Decision

Incorrect:

Participant A

↓

Participant B

Participants never directly exchange messages.

Participant

↓

Principal

is also forbidden.

---

# Runtime Independence

RoundTable does not depend on any specific AI runtime.

The core domain must remain independent from:

* OpenClaw
* Claude Code
* Codex
* LangGraph
* AutoGen
* CrewAI

Runtime integrations belong to adapter layers.

---

# Platform Independence

The Meeting Engine must not depend on:

* Discord
* Slack
* Telegram
* Web UI

These are transport layers.

The Meeting Engine is platform agnostic.

---

# Model Independence

The Meeting Engine never depends directly on a model provider.

Model providers are adapters.

Examples:

* DeepSeek
* OpenAI
* Anthropic
* Gemini

The domain layer must never reference a specific provider.

---

# Development Principles

Architecture before implementation.

Concepts before code.

Discussion before optimization.

Small iterations.

Simple abstractions.

Composition over inheritance.

Interfaces over concrete implementations.

No unnecessary frameworks.

Keep the domain pure.

---

# Project Structure

The project is organized around the domain.

Never around frameworks.

Recommended structure (Monorepo):

apps/

  server/     — Go Meeting Engine（全部 Go 代码）

  web/        — React Principal UI

  android/ / ios/ — 移动端（规划）

pkg/          — 跨 app 共享库（Go SDK 等，v0.1 留空）

docs/

scripts/

仓库根 `go.work` 引用 `apps/server` module。

---

# Naming Rules

Use domain language. Full guide: [NAMING.md](./NAMING.md).

Preferred:

Meeting

Principal

Moderator

Participant

Round

Agenda

Consensus

Confirmation

Minutes

Artifact

Knowledge

Avoid:

MainAgent

SubAgent

MasterAgent

BossAgent

TaskManager

WorkflowManager

ConversationManager

---

# AI Collaboration Rules

AI assistants implement.

Humans design.

Architecture decisions must never be invented automatically.

If architecture is unclear:

STOP.

Ask for clarification.

Never redesign the system autonomously.

Never introduce new core concepts without discussion.

---

# Forbidden Designs

Do not couple the Meeting Engine to Discord.

Do not couple the domain to any LLM provider.

Do not allow Participants to communicate directly.

Do not allow Participants to communicate directly with the Principal.

Do not store business logic inside UI.

Do not let infrastructure leak into the domain.

Do not build around prompts.

Build around concepts.

---

# Implementation Order

The project evolves in the following order:

1. Vision

2. Core Concepts

3. Domain Model

4. Moderator Scheduler

5. Meeting Engine

6. Runtime Adapter

7. Model Adapter

8. Memory

9. Transport Layer

10. Web UI

Infrastructure follows architecture.

Never the opposite.

---

# Long-Term Goal

RoundTable aims to become a reusable Meeting Engine capable of coordinating multiple AI experts through structured collaboration.

The Meeting is the product.

The Principal owns the problem.

The Moderator is the orchestrator.

Participants provide expertise.

Consensus produces collective agreement.

Confirmation produces the Principal's decision.

Everything begins with a Meeting.
