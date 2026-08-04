# Lith Architecture Overview

This document provides a high-level architectural overview of **Lith** and serves as the official map of accepted and proposed Requests for Comments (RFCs).

Lith is designed around clear separation between architecture and implementation details. For constitutional principles, refer to [PROJECT_PRINCIPLES.md](PROJECT_PRINCIPLES.md).

---

## Strategic Vision

Lith is a local-first semantic knowledge engine for Markdown vaults and knowledge bases. 

It continuously indexes, structures, and exposes vault knowledge through high-level domain capabilities, ensuring that AI agents and tools interact with connected semantic context rather than unindexed raw text.

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

RFCs are created, tracked, and discussed directly as [GitHub Issues](https://github.com/lith-project/lith/issues?q=is%3Aissue+label%3Arfc).

| RFC # | Title | Status | Description |
| ----- | ----- | ------ | ----------- |
| [#3](https://github.com/lith-project/lith/issues/3) | Project Vision & Core Architecture | Proposed | Defines project philosophy, domain boundaries, and high-level component model. |

---

## Sub-Directories

* `docs/architecture/`: Technical design docs and detailed component specifications.
* `docs/diagrams/`: Architectural diagrams and topology visualizers.
* `docs/adr/`: Architecture Decision Records tracking specific trade-offs and choices.
