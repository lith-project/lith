---
rfc: "0005"
title: "Background Worker & Job Engine"
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
  - Jobs
supersedes: []
superseded_by: []
---

# RFC-0005: Background Worker & Job Engine

> [!NOTE]
> **Placeholder RFC**: Detailed technical specification will be finalized following completion of [RFC-0002: Domain Model](0002-domain-model.md).

## Summary
Specifies the non-blocking background job pipeline for vault parsing, graph indexing, cache invalidation, and asynchronous plugin tasks.

## Motivation
In accordance with [PROJECT_PRINCIPLES.md](../PROJECT_PRINCIPLES.md) (Principle #8: Background before Synchronous), heavy processing must occur asynchronously to maintain fast client response times.

## Goals
- Provide concurrent, background job queues for vault ingestion.
- Support job status monitoring and graceful task cancellation.

## Non-Goals
- Synchronous blocking file processing on client query paths.

## References

### Internal
- [PROJECT_PRINCIPLES.md](../PROJECT_PRINCIPLES.md)
- [RFC-0001](0001-project-vision.md)
- [RFC-0002](0002-domain-model.md)

### External
- [Go Concurrency Patterns](https://go.dev/blog/pipelines)
