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
4. **Plugin Architecture**: The core engine never depends on embeddings, vector databases, or semantic models. MVP search is full-text and metadata only; semantic, vector, graph, and hybrid retrieval are plugin capabilities, explicitly outside the MVP. See [RFC-0001 §4](rfcs/0001-project-vision.md).
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

| RFC # | Title | Status | Milestone | Subsystems | Description |
| ----- | ----- | ------ | --------- | ------------ | ----------- |
| [0001](rfcs/0001-project-vision.md) | [Project Vision & Strategic Architecture](rfcs/0001-project-vision.md) | Draft | M0 | Core Engine, Architecture | Defines the layer model, the state model, the capability model, the plugin boundary, and supported deployment models. |
| [0002](rfcs/0002-domain-model.md) | [Domain Model & Vault AST](rfcs/0002-domain-model.md) | Draft | M1 | Core Engine, Parsing | Specifies entities, byte-range addressing, total parsing, and deterministic link resolution. |
| [0003](rfcs/0003-storage-engine.md) | [Storage Engine & State Rebuilds](rfcs/0003-storage-engine.md) | Draft | M1 | Storage | Specifies the disposable SQLite schema, the canonical dump defining rebuild determinism, and atomic rebuild. No migrations exist. |
| [0004](rfcs/0004-indexing.md) | [Indexing & Link Graph Engine](rfcs/0004-indexing.md) | Draft | M1 | Graph, Indexing | Specifies change observation, the dirty set, incremental non-local link resolution, and link graph materialization. |
| [0005](rfcs/0005-job-engine.md) | [Background Worker & Job Engine](rfcs/0005-job-engine.md) | Draft | M1 | Jobs, Transactions | Specifies job lifecycle, scheduling, recovery, and the Transaction Coordinator — the single write path to the vault. |

> [!NOTE]
> RFCs are **authored** in dependency order — `0001 → 0002 → 0003 → 0005 → 0004` — because RFC-0004 consumes the transaction semantics defined by RFC-0005. Numbering is stable and never reassigned. See [rfcs/index.md](rfcs/index.md).

---

## Technical Documentation

* `rfcs/`: RFC specifications, index (`rfcs/index.md`), template (`rfcs/templates/rfc-template.md`), and lifecycle instructions.
* `docs/glossary.md`: Canonical definitions for domain terminology.
* `docs/architecture/`: Technical design docs and detailed component specifications.
* `docs/diagrams/`: Cross-RFC composite diagrams only. Diagrams belonging to a single RFC live inline in that RFC.
* `docs/testing/`: Test asset specifications, including the [test vault specification](docs/testing/test-vault-spec.md).
* `docs/adr/`: Architecture Decision Records tracking specific trade-offs and choices.
