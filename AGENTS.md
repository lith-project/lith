# AGENTS.md - Agent Operational Guidelines

This document outlines mandatory guidelines for AI agents operating within the **Lith** codebase.

---

## 1. Operating Rules

1. **Check Core Documents First**:
   - Read [VISION.md](VISION.md) for executive scope and non-goals.
   - Read [PROJECT_PRINCIPLES.md](PROJECT_PRINCIPLES.md) before proposing architecture or modifying core code.
   - Consult [docs/glossary.md](docs/glossary.md) for canonical domain terminology.
   - Read [ARCHITECTURE.md](ARCHITECTURE.md) and [rfcs/index.md](rfcs/index.md).
2. **Never Violate Constitutional Tenets**:
   - Do not make assumptions that SQLite or vector indexes are permanent sources of truth.
   - Do not bypass transactional mechanisms.
   - Do not hardcode vector embeddings as a mandatory dependency of the core engine.
   - Do not modify `PROJECT_PRINCIPLES.md` without an approved RFC.
3. **Respect Architectural Boundaries**:
   - Do not introduce arbitrary Go directory structures (`pkg/...`, `internal/...`) unless agreed upon in an active RFC.
   - Maintain strict separation between core domain logic and external adapters (CLI, REST, MCP, SDK).

---

## 2. Workflows for Agents

### Proposing Architectural Changes
1. Draft an RFC file in `rfcs/NNNN-title.md` using [rfcs/templates/rfc-template.md](rfcs/templates/rfc-template.md).
2. Ensure the proposal includes machine-readable frontmatter (`status: Draft`, `milestone`, `requires`, `capability`).
3. Update [rfcs/index.md](rfcs/index.md) and [ARCHITECTURE.md](ARCHITECTURE.md) to reference the new RFC.

### Code Editing & Refactoring
1. Verify signature changes across all calling sites.
2. Maintain clear, wrapped error messages.
3. Run tests and static analysis (`go test ./...`) after edits.

---

## 3. Communication & Summaries

- Provide concise, structured pull request descriptions referencing corresponding issues or RFCs.
- Always use repository-relative links (`VISION.md`, `PROJECT_PRINCIPLES.md`, `rfcs/index.md`) when referring to project documents.
