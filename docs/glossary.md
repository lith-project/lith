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

---

## Domain Model Terms

### Asset
A non-Markdown file within a Vault — image, PDF, canvas, base, or other. Assets are identified and tracked, but their internal structure is not interpreted.

### Diagnostic
A structured observation recorded while processing a Note or the store: a stable code, a severity (`Info`, `Warning`, `Error`), and a byte range. `Error` severity means a Note is degraded, never that indexing failed. Codes are part of the public contract; message text is not.

### Opaque Block
A Block whose syntax Lith recognizes but deliberately does not interpret — a Dataview query, a Mermaid diagram, a Bases fragment. Content is preserved exactly. This is what lets the model survive the long tail of plugin syntax without understanding it.

---

## Storage & Indexing Terms

### Canonical Dump
An ordered, normalized projection of all durable derived state: durable columns only, sorted by natural key, fixed table order, fixed encoding. Its hash is the logical identity of the store, and equality of two dumps is the definition of "logically identical" used by the rebuild guarantee.

### Durable Column
A column that is part of logical state, included in the Canonical Dump, and recomputable from the Vault.

### Volatile Column
A column excluded from the Canonical Dump. A volatile column may never affect any query answer; if it can, it is durable.

### Vault Fingerprint
The recorded identity of the Vault a store was built from — its canonical root path and normalization form. A store opened against a different Vault is rebuilt rather than adopted.

### Dirty Set
The set of Notes awaiting re-index, each with the reason it was marked. Membership is idempotent, and entries leave only when their batch commits, so an interrupted batch loses no work.

### Reconciliation Scan
An authoritative walk of the Vault that hashes every candidate file. Size and modification time may order the work but never conclude that a file is unchanged. It is the mechanism by which correctness is recovered when the filesystem watcher misses events.

### Resolution Key
A value under which a Note can be referenced — its basename, its Vault-relative path, or one of its aliases. A reverse index from Resolution Keys to referring Notes bounds the work of recomputing link resolution when the Vault changes.

### Watcher Gap
Any condition invalidating the filesystem watcher's guarantees: queue overflow, notification error, platform limit, suspended process, or unsupported filesystem. A Watcher Gap forces a Reconciliation Scan.

---

## Job & Transaction Terms

### Change Proposal
A requested modification to a Note, expressed as a content-identity precondition plus a set of sorted, non-overlapping byte-range splices. It cannot express reformatting, because the domain model provides no serializer with which to express it.

### Intent Journal
A short-lived crash-recovery record written outside the Vault before a multi-file transaction is applied. It permits recovery to either complete or reverse the transaction, since atomic replacement of several files is not available on POSIX.

### Identity Key
The coalescing key of a Job. Two queued Jobs sharing an Identity Key are one Job, which is how an event storm becomes bounded work.
