# Lith Glossary & Canonical Terminology

This document defines canonical definitions for core concepts and domain terms in **Lith**. All RFCs, technical specifications, codebase comments, and client interfaces must use these terms consistently.

---

## Domain Concepts

### Vault
The user's local directory containing Markdown files (`.md`), assets, and configuration files (e.g., an Obsidian vault). A Vault is the canonical, filesystem-based source of truth.

### Knowledge Source
An abstraction representing an input stream or file tree of Markdown documents ingested by Lith. In standard desktop deployment, a Knowledge Source maps to a local Vault.

### Workspace
An isolated operational context in Lith representing a collection of Knowledge Sources, derived index databases, and active query sessions.

### Tenant
An isolated boundary defining user or organizational identity and access controls within multi-workspace or enterprise deployment models.

---

## Structural Entities

### Note (Document)
A single Markdown file within a Vault, parsed into frontmatter metadata, sections, blocks, tags, and link relations.

### Section
A structural subdivision of a Note demarcated by Markdown heading levels (`#` through `######`).

### Block
The smallest addressable structural unit within a Note (e.g., a paragraph, list item, blockquote, code block, or callout). Blocks can be uniquely referenced via Obsidian block identifiers (`^block-id`).

---

## System Abstractions

### Capability
A high-level, domain-specific semantic operation exposed by Lith to external callers (e.g., `query_graph`, `resolve_entities`, `propose_transactional_update`). Capabilities contrast with primitive CRUD operations.

### Job
An asynchronous unit of background work managed by the background job engine (e.g., incremental note parsing, full index rebuilding, background embedding generation).

### Derived State
Any database, relational table, graph cache, or search index maintained by Lith. Derived state is non-canonical and must be completely reconstructible from raw Vault source files at any time.

### Capability Registry
The component owning capability identity, discovery, dispatch, and input validation. It is the single public surface of the engine: interface adapters (CLI, REST, MCP, SDK) reach the engine only through it, never past it into storage, indexing, or parsing.

### Capability Catalog
The living inventory of capabilities that exist, maintained at [docs/reference/capability-catalog.md](reference/capability-catalog.md). The catalog lists instances; the capability *model* — identity, metadata, lifecycle — is defined by RFC-0001. A capability absent from the catalog must not be exposed by any interface.

### Conformance Assertion
A numbered, independently verifiable requirement within an RFC, identified as `C-N` and referenced project-wide as `RFC-000X/C-N`. Every assertion states an outcome observable by a test, a static check, a benchmark threshold, or a documented manual procedure. Assertion identifiers are stable and never reused; each carries an assertion state of `Active` or `Withdrawn`, which is distinct from an RFC's `status`.

### Transaction Coordinator
The only component permitted to write Vault files. It validates a proposed change, checks it against a content-identity precondition, and applies it atomically — or rejects it, leaving every Vault file byte-identical.

### Plugin Host
The component loading optional capabilities and managing their lifecycle. It is never required for the engine to start, and the absence of a plugin is never an error.
