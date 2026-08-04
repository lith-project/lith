# AGENTS.md - Agent Operational Guidelines

This document outlines mandatory guidelines for AI agents operating within the **Lith** codebase.

---

## 1. Operating Rules

1. **Check Core Documents First**:
   - Read [PROJECT_PRINCIPLES.md](PROJECT_PRINCIPLES.md) before proposing architecture or modifying core code.
   - Read [ARCHITECTURE.md](ARCHITECTURE.md) and linked RFC issues on GitHub.
2. **Never Violate Constitutional Tenets**:
   - Do not make assumptions that SQLite or vector indexes are permanent sources of truth.
   - Do not bypass transactional mechanisms.
   - Do not hardcode vector embeddings as a mandatory dependency of the core engine.
3. **Respect Architectural Boundaries**:
   - Do not introduce arbitrary Go directory structures (`pkg/...`, `internal/...`) unless agreed upon in an active RFC.
   - Maintain strict separation between core domain logic and external adapters (CLI, REST, MCP, SDK).

---

## 2. Workflows for Agents

### Proposing Architectural Changes
1. Create an RFC as a GitHub Issue labeled `rfc` and `architecture`.
2. Ensure the proposal explicitly details the problem, motivation, domain boundaries, and alignment with `PROJECT_PRINCIPLES.md`.
3. Update `ARCHITECTURE.md` to reference the RFC issue number.

### Code Editing & Refactoring
1. Verify signature changes across all calling sites.
2. Maintain clear, wrapped error messages.
3. Run tests and static analysis (`go test ./...`) after edits.

---

## 3. Communication & Summaries

- Provide concise, structured pull request descriptions referencing corresponding issues or RFCs.
- Always include relative links or GitHub Issue references when referring to project documents or RFC proposals.
