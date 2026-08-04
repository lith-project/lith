---
rfc: "0005"
title: "Background Worker & Job Engine"
status: Draft
authors:
  - Lith Maintainers <maintainers@lith.dev>
created: 2026-08-04
updated: 2026-08-04
discussion: https://github.com/lith-project/lith/discussions
supersedes:
superseded_by:
---

# RFC-0005: Background Worker & Job Engine

## Summary
Specifies the non-blocking background job pipeline for vault parsing, graph indexing, cache invalidation, and asynchronous plugin tasks.

## Motivation
In accordance with [PROJECT_PRINCIPLES.md](../PROJECT_PRINCIPLES.md) (Principle #8: Background before Synchronous), heavy processing must occur asynchronously to maintain fast client response times.

## Goals
- Provide concurrent, background job queues for vault ingestion.
- Support job status monitoring and graceful task cancellation.

## Non-Goals
- Synchronous blocking file processing on client query paths.

## Background
See [PROJECT_PRINCIPLES.md](../PROJECT_PRINCIPLES.md).

## Proposed Design
*To be elaborated during M0 architecture phase.*

## Alternatives Considered
1. Synchronous single-threaded parsing.
2. Unbounded goroutine spawns without task queues.

## Risks
TBD.

## Migration
None.

## Open Questions
- [ ] Memory limits for worker queues on large vaults (100k+ notes).

## Future Work
- Job engine implementation in Go.

## References
- [PROJECT_PRINCIPLES.md](../PROJECT_PRINCIPLES.md)
- [RFC-0001](0001-project-vision.md)
