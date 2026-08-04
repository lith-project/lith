---
rfc: "0004"
title: "Indexing & Link Graph Engine"
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
  - Graph
  - Indexing
supersedes: []
superseded_by: []
---

# RFC-0004: Indexing & Link Graph Engine

> [!NOTE]
> **Placeholder RFC**: Detailed technical specification will be finalized following completion of [RFC-0002: Domain Model](0002-domain-model.md).

## Summary
Specifies the link graph construction algorithms, entity relation extraction, and graph traversal capabilities of Lith.

## Motivation
Markdown vaults contain rich interconnected structures (wiki links, tags, block embeds). This RFC defines how Lith extracts, indexes, and queries these connections.

## Goals
- Build bidirectional link graphs across vault notes.
- Support graph query primitives and entity resolution.

## Non-Goals
- Mandatory vector embedding generation (plugins handle vector search).

## References

### Internal
- [PROJECT_PRINCIPLES.md](../PROJECT_PRINCIPLES.md)
- [RFC-0001](0001-project-vision.md)
- [RFC-0002](0002-domain-model.md)

### External
- [Obsidian Wiki-link Syntax](https://help.obsidian.md/Editing+and+formatting/Internal+links)
