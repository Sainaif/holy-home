# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Holy Home is a self-hosted household bill and finance management app for shared living situations. It tracks utility bills, splits costs based on usage, records meter readings, and manages loans between residents.

## Tech Stack

- **Backend**: Go 1.24 + Fiber v2 + MongoDB
- **Frontend**: Vue 3 + Vite + Tailwind CSS + Pinia
- **Auth**: JWT tokens (15min access / 30 day refresh), Argon2id passwords, WebAuthn/passkeys, optional TOTP 2FA

## Development Commands

### Backend (from `/backend`)
```bash
go run ./cmd/api           # Start dev server
go test -v -race ./...     # Run tests with race detection
go vet ./...               # Lint
gofmt -s -w .              # Format code
go build -o api ./cmd/api  # Build binary
```

### Frontend (from `/frontend`)
```bash
npm install                # Install dependencies
npm run dev                # Start Vite dev server
npm test -- --run          # Run tests once
npm run test:coverage      # Run tests with coverage
npm run build              # Production build
```

### Full Stack (from `/deploy`)
```bash
docker-compose up -d                      # Start all services
docker-compose build && docker-compose up -d  # Rebuild after changes
```

**Ports**: Frontend on 16161, API on 16162, MongoDB on 27017

## Architecture

```
backend/
├── cmd/api/main.go           # Entry point
└── internal/
    ├── config/               # Environment config loading
    ├── database/             # MongoDB connection
    ├── models/               # Data structures
    ├── handlers/             # HTTP route handlers
    ├── services/             # Business logic layer
    ├── middleware/           # Auth, rate limiting, request ID
    └── utils/                # JWT, password hashing, TOTP, WebAuthn

frontend/
└── src/
    ├── views/                # Page components (Dashboard, Bills, Balance, etc.)
    ├── components/           # Reusable UI components
    ├── composables/          # Vue 3 composition functions
    ├── stores/               # Pinia state (auth, notification)
    ├── api/                  # Axios client with JWT interceptors
    └── locales/pl.json       # Polish translations (UI is in Polish)
```

**Pattern**: Handlers call services for business logic. Services interact with MongoDB.

## Testing Requirements

Backend tests require a MongoDB replica set:
```bash
docker run -d -p 27017:27017 --name test-mongo mongo:7 --replSet rs0
docker exec test-mongo mongosh --eval "rs.initiate()"
```

CI runs `go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...`

## Key Domain Concepts

- **Bills**: Utility bills (electricity, gas, internet, custom) with allocation types
- **Consumptions**: Meter readings per user/group
- **Allocations**: Calculated cost splits - personal usage charged individually, common areas split equally
- **Groups**: Household groups with cost-splitting weights
- **Loans**: Money lent/borrowed between residents

## MongoDB Collections

users, groups, bills, consumptions, allocations, payments, loans, events, chores, supplies

## Environment

Copy `.env.example` to `.env`. Key variables:
- `JWT_SECRET` / `JWT_REFRESH_SECRET` - generate with `openssl rand -base64 32`
- `ADMIN_EMAIL` / `ADMIN_PASSWORD` - bootstrap admin on first run
- `APP_DOMAIN` - critical for WebAuthn/passkey functionality
- `MONGO_URI` - defaults to `mongodb://mongo:27017`

## [AGENT TOOLBOX]

You have access to multiple AI agents - USE THEM AGGRESSIVELY. Don't try to do everything yourself.
Delegate early and often. Parallelise where possible.

### Internal Subagents (Claude family)

- **Haiku** (fast, cheap)
  - File/directory scanning and inventory
  - Grep-like searches, pattern matching
  - Quick summaries and categorisation
  - Bulk processing of repetitive tasks

- **Sonnet** (balanced)
  - Code review and logic analysis
  - Implementation of well-defined tasks
  - Test generation and documentation

- **Opus** (deep reasoning)
  - Complex architectural decisions
  - Ambiguous or cross-cutting problems
  - High-stakes design choices

### External AI Agents

#### Gemini
Google's AI with massive context window (1M tokens). Good for processing large codebases and alternative perspectives.

**Interactive mode:**
```bash
gemini
# Then type your prompt in the REPL interface
```

**Non-interactive mode:**
```bash
gemini -p "your prompt here"
```

**Use for:** sanity-checking designs, processing huge files/codebases, alternative implementations, cross-validation of reasoning.

---

#### Codex
OpenAI's coding agent with strong code understanding and file/command execution capabilities.

**Interactive mode:**
```bash
codex
# Then type your task in the TUI interface
```

**Non-interactive mode:**
```bash
codex exec "your prompt here"
```

**Use for:** code review, static analysis, different coding perspectives, verification of tricky logic, automated refactoring.

---

### Usage Philosophy

1. **Default to delegation** - If a subtask can be handled by a subagent, delegate it
2. **Use external AI for verification** - Cross-check important decisions with different models
3. **Parallelise aggressively** - Spawn multiple agents for independent subtasks
4. **Combine perspectives** - Use model diversity to catch blind spots
5. **Escalate appropriately** - Use heavier models only when lighter ones fail

### When to call external AI

- Before finalising architectural decisions → ask Gemini for alternative approach
- After writing complex logic → send to Codex for review
- When stuck on a problem → get fresh perspective from different model
- For critical code paths → get consensus from multiple models
