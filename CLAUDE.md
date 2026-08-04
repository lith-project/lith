# CLAUDE.md - Assistant & Developer Guidelines

This document provides context, conventions, and operational rules for AI assistants (including Claude, Gemini, etc.) and human developers working in the **Lith** repository.

---

## Project Context

**Lith** (`lith-project/lith`) is a local-first semantic knowledge engine for Markdown and Obsidian vaults. It builds structured, transactional understanding of Markdown vaults for AI agents and client interfaces.

---

## Core Rules & Principles

1. **Executive Vision & Scope**: Read [VISION.md](VISION.md) to understand project boundaries and intentional non-goals.
2. **Constitutional Adherence**: Always respect [PROJECT_PRINCIPLES.md](PROJECT_PRINCIPLES.md). Amending principles requires an RFC.
   - Markdown is the canonical source of truth.
   - All SQLite / database indexes are disposable derived state.
   - AI agents never perform raw, unvalidated edits directly on Markdown files.
   - Everything must be transactional.
   - Interfaces (CLI, REST, MCP, SDK) are peers.
   - Vector embeddings are optional plugins, not hardcoded core assumptions.
3. **RFC-Driven Architectural Changes**: Major architectural additions must be created and submitted as an RFC under `rfcs/NNNN-title.md` following [rfcs/templates/rfc-template.md](rfcs/templates/rfc-template.md).
4. **Canonical Glossary**: Consult [docs/glossary.md](docs/glossary.md) for canonical definitions of domain terms (*Vault*, *Workspace*, *Capability*, *Job*, *Block*, etc.).

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

- `VISION.md`: 5-minute executive vision and non-goals.
- `PROJECT_PRINCIPLES.md`: Core project constitution.
- `ARCHITECTURE.md`: High-level map of components and accepted RFCs.
- `docs/glossary.md`: Canonical domain terminology.
- `rfcs/`: Specifications, index ([rfcs/index.md](rfcs/index.md)), template ([rfcs/templates/rfc-template.md](rfcs/templates/rfc-template.md)), and documentation ([rfcs/README.md](rfcs/README.md)).
- `ROADMAP.md`: Milestone progress (M0 - M4).
- `.github/`: Issue templates, PR template, CODEOWNERS.
