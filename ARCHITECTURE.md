# Lith Architecture Overview

This document provides a high-level architectural overview of **Lith** and serves as the official map of accepted and proposed Requests for Comments (RFCs).

For executive vision, see [VISION.md](VISION.md). For constitutional principles, refer to [PROJECT_PRINCIPLES.md](PROJECT_PRINCIPLES.md). For canonical domain terms, see [docs/glossary.md](docs/glossary.md).

---

## Strategic Vision

Lith is a local-first semantic knowledge engine for Markdown vaults and knowledge bases. 

It continuously indexes, structures, and exposes vault knowledge through high-level domain capabilities, ensuring that AI agents and tools interact with connected semantic context rather than unindexed raw text.

---

## Engineering Hierarchy

```
VISION.md              (Executive Vision & Scope)
        │
        ▼
PROJECT_PRINCIPLES.md  (Immutable Constitution — Amending requires an RFC)
        │
        ▼
RFCs (architecture)    (rfcs/ index.md & specifications)
        │
        ▼
Issues (work)          (Actionable GitHub Issues)
        │
        ▼
Pull Requests (impl)   (Code changes & validation)
        │
        ▼
Code
```

---

## Architectural Principles

1. **Source of Truth**: Markdown files are canonical; database state is disposable derived context.
2. **Transactional Integrity**: Mutations and index updates execute atomically.
3. **Capability-Oriented**: Exposes semantic capabilities rather than primitive CRUD operations.
4. **Plugin Architecture**: Features like vector embeddings and semantic search operate as swappable extension plugins rather than mandatory core dependencies.
5. **Peer Interfaces**: Interfaces (CLI, REST API, MCP Protocol, SDK) are equal peer adapters wrapping the core knowledge engine.

---

## High-Level Domain Components

* **Vault Ingestion & Observer**: Monitors local filesystem changes and builds structured domain representations.
* **Knowledge Engine**: Manages entities, link graphs, relations, and transactional semantics.
* **Derived Index Store**: Persists disposable local state (relational graphs and query indexes) for fast context retrieval.
* **AI Interface Layer**: Exposes peer protocols (CLI, REST API, Model Context Protocol, SDK) for external callers.
* **Plugin Hooks**: Manages optional background extensions (e.g. vector embedding generation, external tool integrations).

---

## Deployment Models

Lith is designed to run in multiple deployment modes:
* **Embedded Engine**: Integrated directly into Go applications or CLI binaries as a library.
* **Local Background Daemon**: Running locally alongside markdown editors (such as Obsidian) to maintain live indexes.
* **Sidecar Container / Process**: Operating alongside AI agent runners in isolated local or server environments.

---

## Index of RFCs

Full specifications, metadata filters, and lifecycle guidelines are maintained in **[rfcs/index.md](rfcs/index.md)** and **[rfcs/README.md](rfcs/README.md)**.

| RFC # | Title | Status | Milestone | Capabilities | Description |
| ----- | ----- | ------ | --------- | ------------ | ----------- |
| [0001](rfcs/0001-project-vision.md) | [Project Vision & Core Architecture](rfcs/0001-project-vision.md) | Draft | M0 | Core Engine, Architecture | Defines project philosophy, domain boundaries, and high-level component model. |
| [0002](rfcs/0002-domain-model.md) | [Domain Model & Vault AST](rfcs/0002-domain-model.md) | Draft | M1 | Core Engine, Parsing | Specifies notes, blocks, links, and frontmatter AST representations. |
| [0003](rfcs/0003-storage-engine.md) | [Storage Engine & State Rebuilds](rfcs/0003-storage-engine.md) | Draft | M1 | Storage | Specifies disposable database schemas and full rebuild mechanisms. |
| [0004](rfcs/0004-indexing.md) | [Indexing & Link Graph Engine](rfcs/0004-indexing.md) | Draft | M1 | Graph, Indexing | Specifies link graph extraction, entity resolution, and graph query primitives. |
| [0005](rfcs/0005-job-engine.md) | [Background Worker & Job Engine](rfcs/0005-job-engine.md) | Draft | M1 | Jobs | Specifies non-blocking asynchronous job queues and task lifecycle. |

---

## Technical Documentation

* `rfcs/`: RFC specifications, index (`rfcs/index.md`), template (`rfcs/templates/rfc-template.md`), and lifecycle instructions.
* `docs/glossary.md`: Canonical definitions for domain terminology.
* `docs/architecture/`: Technical design docs and detailed component specifications.
* `docs/diagrams/`: Architectural diagrams and topology visualizers.
* `docs/adr/`: Architecture Decision Records tracking specific trade-offs and choices.
