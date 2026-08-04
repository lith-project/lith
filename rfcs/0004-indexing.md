---
rfc: "0004"
title: "Indexing & Link Graph Engine"
status: Draft
authors:
  - Lith Maintainers <maintainers@lith.dev>
created: 2026-08-04
updated: 2026-08-04
discussion: https://github.com/lith-project/lith/discussions
supersedes:
superseded_by:
---

# RFC-0004: Indexing & Link Graph Engine

## Summary
Specifies the link graph construction algorithms, entity relation extraction, and graph traversal capabilities of Lith.

## Motivation
Markdown vaults contain rich interconnected structures (wiki links, tags, block embeds). This RFC defines how Lith extracts, indexes, and queries these connections.

## Goals
- Build bidirectional link graphs across vault notes.
- Support graph query primitives and entity resolution.

## Non-Goals
- Mandatory vector embedding generation (plugins handle vector search).

## Background
See [RFC-0001](0001-project-vision.md).

## Proposed Design
*To be elaborated during M0 architecture phase.*

## Alternatives Considered
1. Primitive grep/text search.
2. Direct Neo4j / external graph database dependency.

## Risks
TBD.

## Migration
None.

## Open Questions
- [ ] Resolving uncreated / dangling wiki links (`[[non-existent-note]]`).

## Future Work
- Query capabilities for AI interfaces.

## References
- [RFC-0001](0001-project-vision.md)
