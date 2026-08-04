---
rfc: "0003"
title: "Storage Engine & State Rebuilds"
status: Draft
authors:
  - Lith Maintainers <maintainers@lith.dev>
created: 2026-08-04
updated: 2026-08-04
discussion: https://github.com/lith-project/lith/discussions
supersedes:
superseded_by:
---

# RFC-0003: Storage Engine & State Rebuilds

## Summary
Defines the disposable storage engine architecture, transactional state management, and complete vault re-indexing mechanisms.

## Motivation
In accordance with [PROJECT_PRINCIPLES.md](../PROJECT_PRINCIPLES.md), SQLite and index databases are disposable derived state. This RFC specifies how state is persisted, queried, and completely reconstructed from raw Markdown files.

## Goals
- Specify local database table schemas and transactional write boundaries.
- Provide fast full-rebuild mechanisms from Markdown sources.

## Non-Goals
- Replacing Markdown files as the canonical source of truth.

## Background
See [PROJECT_PRINCIPLES.md](../PROJECT_PRINCIPLES.md) (Principle #2: SQLite is Disposable).

## Proposed Design
*To be elaborated during M0 architecture phase.*

## Alternatives Considered
1. Persistent database with manual sync.
2. In-memory only state without local persistence.

## Risks
TBD.

## Migration
None.

## Open Questions
- [ ] Schema versioning and auto-rebuild triggers on version mismatch.

## Future Work
- Database driver and migration engine.

## References
- [PROJECT_PRINCIPLES.md](../PROJECT_PRINCIPLES.md)
- [RFC-0001](0001-project-vision.md)
