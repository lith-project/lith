---
rfc: "0001"
title: "Project Vision & Strategic Architecture"
status: Draft
milestone: M0
authors:
  - Lith Maintainers <maintainers@lith.dev>
created: 2026-08-04
updated: 2026-08-04
discussion: https://github.com/lith-project/lith/discussions
requires: []
capability:
  - Core Engine
  - Architecture
supersedes: []
superseded_by: []
---

# RFC-0001: Project Vision & Strategic Architecture

## Summary

**Lith** is a local-first semantic knowledge engine designed for Markdown knowledge bases and Obsidian vaults. 

Unlike file-system scripts or raw context dumpers, Lith continuously ingests, extracts, and maintains structured semantic knowledge from Markdown source files. It provides AI agents, local workflows, and client applications with transactional, capability-based access to connected knowledge rather than raw, unindexed files.

## Motivation

Markdown has become the de facto format for personal and organizational knowledge bases (e.g., Obsidian, Dendron, logseq, plain directory vaults). However, current interactions between AI agents and Markdown vaults suffer from critical limitations:

1. **Context Fragmentation**: Tools rely on crude file-path listing, glob searching, or static line chunking, ignoring rich wiki links (`[[link]]`), frontmatter metadata, block embeds, and link graph relationships.
2. **Brittle File Mutations**: AI agents often attempt direct regex edits or file rewrites, leading to corrupt links, missing frontmatter, or lost context.
3. **Lack of Transactional Integrity**: Index updates and queries are ad-hoc and unverified, causing stale state and inconsistent query answers.
4. **Tight Coupling to Specific Interfaces**: Tools are often built as single-purpose MCP wrappers or editor plugins rather than robust, standalone knowledge engines.

Lith solves these challenges by serving as an independent, local-first engine that bridges the gap between raw file-based Markdown vaults and structured semantic AI capabilities.

## Goals

- Establish Markdown as the canonical, immutable source of truth.
- Maintain disposable, transactional local indexes for fast graph and semantic queries.
- Expose peer interfaces (CLI, REST API, MCP Protocol, SDK) wrapping a unified knowledge engine.
- Define a capability-oriented model for AI agents rather than primitive file CRUD operations.
- Support swappable extension plugins (such as vector embeddings) without coupling the core engine.

## Non-Goals

Lith explicitly does **NOT** attempt to be:
- A text/Markdown editor or note-taking UI (Obsidian, VS Code, and Neovim fulfill this role).
- A cloud synchronization protocol (Obsidian Sync, Syncthing, and Git handle file transport).
- A generic web scraper or unstructured document reader.

## Background

Lith was conceived to provide AI agents with structured domain knowledge from personal vaults. Existing tools treat Markdown either as raw strings or depend on monolithic cloud vector databases. Lith prioritizes local-first privacy, transactional index rebuilds, and graph-first semantics.

## Proposed Design

### Domain Boundaries & System Topology

Lith decomposes knowledge engine responsibilities into clear domain boundaries:
* **Vault Observation & Ingestion**: Tracking file updates and building structured ASTs.
* **Knowledge Engine**: Extracting entities, links, structure, and semantic relationships.
* **State & Derived Cache Store**: Maintaining transactional local graph state and query indexes.
* **Capability Interfaces**: Exposing peer endpoints (CLI, REST API, MCP Protocol, SDK) for external consumers.
* **Extension Hooks**: Allowing optional plugins (e.g., vector embeddings, custom entity extractors) without coupling the core engine.

```
+-------------------------------------------------------------------+
|                        Client Applications                        |
|             (AI Agents, CLI Tools, Web UIs, Extensions)           |
+-------------------------------------------------------------------+
                                  |
                                  v
+-------------------------------------------------------------------+
|                           AI Interfaces                           |
|                  (CLI  |  REST  |  MCP  |  SDK)                   |
+-------------------------------------------------------------------+
                                  |
                                  v
+-------------------------------------------------------------------+
|                         Knowledge Engine                          |
|    - Graph Relations              - Entity Context                |
|    - Transaction Management       - Background Worker Pipeline    |
+-------------------------------------------------------------------+
      |                                                |
      v                                                v
+-----------------------+                    +----------------------+
|  Derived State Store  |                    |  Extension Plugins   |
|   (Disposable Index)  |                    | (Embeddings, Vector) |
+-----------------------+                    +----------------------+
      ^
      | (Rebuilt from)
+-------------------------------------------------------------------+
|                        Local Filesystem                           |
|                 (Markdown Vault / Source of Truth)                 |
+-------------------------------------------------------------------+
```

### Capability-Driven Interaction Model
AI agents interact with Lith via semantic capabilities (e.g., `query_graph`, `get_entity_context`, `resolve_relations`, `propose_transactional_change`) rather than direct raw file read/write operations.

## Alternatives Considered

1. **Direct MCP Wrapper around Filesystem**:
   - *Rationale for Rejection*: Fails to maintain persistent link graphs, requires re-parsing files on every query, and offers no transactional edit safety.
2. **Monolithic Vector-Only Store**:
   - *Rationale for Rejection*: Vector embeddings lose structural graph relationships (`[[wiki-links]]`, tags, frontmatter taxonomy) and introduce external API/model dependencies into the core engine.

## Risks

- **Index Sync Latency**: Heavy vault changes may take time to re-index. *Mitigation*: Asynchronous background worker pipeline and incremental AST diffing.
- **Dangling References**: Deleted or renamed Markdown files may break links. *Mitigation*: Graph transaction checks and reference audit tools.

## Migration

Not applicable for initial architecture phase.

## Open Questions

- [ ] Standardized format for transactional update proposals submitted by AI agents.

## Future Work

- **RFC-0002**: Domain Model & Vault AST Representation
- **RFC-0003**: Storage Engine & State Rebuilds
- **RFC-0004**: Indexing & Link Graph Engine
- **RFC-0005**: Background Worker & Job Engine

## References

### Internal
- [PROJECT_PRINCIPLES.md](../PROJECT_PRINCIPLES.md)
- [VISION.md](../VISION.md)
- [ARCHITECTURE.md](../ARCHITECTURE.md)
- [docs/glossary.md](../docs/glossary.md)

### External
- [Obsidian Flavored Markdown Specification](https://help.obsidian.md/Editing+and+formatting/Obsidian+Flavored+Markdown)
- [Model Context Protocol Specification](https://modelcontextprotocol.io/)
