# Lith Project Principles

This document defines the immutable constitutional principles governing the architecture, design, and implementation of **Lith**. Every RFC, architectural decision, and code contribution must adhere to these core tenets.

---

## Constitution & Immutability

> [!IMPORTANT]
> **Immutability Rule**: This document serves as Lith's constitution. Amending, modifying, or removing any principle defined herein **requires a formal RFC** approved by maintainers.

---

## Core Tenets

### 1. Markdown is the Source of Truth
The user's Markdown vault files are the sole canonical source of authority. Lith builds indexes, graphs, and representations *from* Markdown, but never treats secondary indexes or cache stores as authoritative state.

### 2. SQLite is Disposable
All internal storage—including relational tables, link graphs, caches, and indexes—is derived state. Any database file or index must be reconstructible from scratch at any time by re-scanning the Markdown source of truth.

### 3. AI Never Edits Markdown Directly
AI agents and tools interact with Lith through structured capabilities and transactional operations. Agents never perform direct, unvalidated, or unstructured raw edits on user Markdown files.

### 4. Everything is Transactional
State updates, indexing operations, and knowledge mutations execute with strict transactional guarantees. Partial updates, corrupted index writes, or dangling graph states are unacceptable.

### 5. Knowledge is Semantic
Lith views Markdown files not merely as unstructured blobs of text or raw AST nodes, but as structured, interconnected knowledge domains with rich semantics, entities, relations, and context.

### 6. Synchronization is External
Lith leaves file synchronization, cloud sync, git remotes, and multi-device transport to external specialized systems (e.g., Obsidian Sync, Syncthing, Git). Lith operates locally on the state of the local file system.

### 7. Capabilities over CRUD
Interfaces expose rich, domain-specific semantic capabilities (e.g., query context, extract relations, analyze graph topology) rather than simple generic CRUD (Create, Read, Update, Delete) operations.

### 8. Background before Synchronous
Heavy processing—such as parsing, graph indexing, relation extraction, and background semantic analysis—is performed asynchronously in non-blocking background workers so that user experience and query availability remain responsive.

---

## Adherence

Any proposed feature, pull request, or implementation that violates any of these principles must be rejected or redesigned to comply with this document.
