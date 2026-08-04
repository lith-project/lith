# CLAUDE.md - Assistant & Developer Guidelines

This document provides context, conventions, and operational rules for AI assistants (including Claude, Gemini, etc.) and human developers working in the **Lith** repository.

---

## Project Context

**Lith** (`lith-project/lith`) is a local-first semantic knowledge engine for Markdown and Obsidian vaults. It builds structured, transactional understanding of Markdown vaults for AI agents and client interfaces.

---

## Core Rules & Principles

1. **Constitutional Adherence**: Always respect [PROJECT_PRINCIPLES.md](PROJECT_PRINCIPLES.md).
   - Markdown is the canonical source of truth.
   - All SQLite / database indexes are disposable derived state.
   - AI agents never perform raw, unvalidated edits directly on Markdown files.
   - Everything must be transactional.
   - Interfaces (CLI, REST, MCP, SDK) are peers.
   - Vector embeddings are optional plugins, not hardcoded core assumptions.
2. **RFC-Driven Architectural Changes**: Major architectural additions must be created and discussed as GitHub Issues with the `rfc` label before implementation.
3. **No Premature Implementation Locking**: Keep architecture and domain models decoupled from specific parser libraries or hardcoded directory layouts until approved via RFC.

---

## Go Development & Conventions (Post-M0 Implementation)

When Go codebase development begins:
* **Tooling Commands**:
  - Build: `go build ./...`
  - Test: `go test -v ./...`
  - Lint: `golangci-lint run`
  - Format: `gofmt -s -w .` / `goimports -w .`
* **Error Handling**: Always return explicit, wrapped errors (`fmt.Errorf("failed to parse file %s: %w", path, err)`).
* **Concurrency**: Use context cancellation (`context.Context`) for all long-running or worker pool operations.
* **Testing**: Write table-driven unit tests. Keep mock interfaces small and localized.

---

## Repository Structure

- `docs/`: Technical documentation (`docs/architecture/`, `docs/diagrams/`, `docs/adr/`).
- `PROJECT_PRINCIPLES.md`: Core project constitution.
- `ARCHITECTURE.md`: High-level map of components and indexed RFC GitHub Issues.
- `ROADMAP.md`: Milestone progress (M0 - M4).
- `.github/`: Issue templates, PR template, CODEOWNERS.
