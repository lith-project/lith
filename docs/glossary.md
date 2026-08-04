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
