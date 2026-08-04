---
rfc: "0003"
title: "Storage Engine & State Rebuilds"
status: Draft
milestone: M1
authors:
  - Lith Maintainers <maintainers@lith.dev>
created: 2026-08-04
updated: 2026-08-04
discussion: https://github.com/lith-project/lith/discussions
requires:
  - "0001"
  - "0002"
capability:
  - Storage
supersedes: []
superseded_by: []
---

# RFC-0003: Storage Engine & State Rebuilds

> [!NOTE]
> **Placeholder RFC**: Detailed technical specification will be finalized following completion of [RFC-0002: Domain Model](0002-domain-model.md).

## Summary
Defines the disposable storage engine architecture, transactional state management, and complete vault re-indexing mechanisms.

## Motivation
In accordance with [PROJECT_PRINCIPLES.md](../PROJECT_PRINCIPLES.md), SQLite and index databases are disposable derived state. This RFC specifies how state is persisted, queried, and completely reconstructed from raw Markdown files.

## Goals
- Specify local database table schemas and transactional write boundaries.
- Provide fast full-rebuild mechanisms from Markdown sources.

## Non-Goals
- Replacing Markdown files as the canonical source of truth.

## References

### Internal
- [PROJECT_PRINCIPLES.md](../PROJECT_PRINCIPLES.md)
- [RFC-0001](0001-project-vision.md)
- [RFC-0002](0002-domain-model.md)

### External
- [SQLite Documentation](https://www.sqlite.org/docs.html)
