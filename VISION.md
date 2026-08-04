# Lith Vision & Purpose

This document explains the overarching vision, target audience, and scope of **Lith**. It is written to be accessible to anyone in five minutes, providing clarity on why Lith exists, who it is built for, and what boundaries define its core mission.

---

## 1. Why Does Lith Exist?

Markdown has become the default medium for local notes, personal knowledge management, and technical documentation. However, current AI tools and LLM agents interact with Markdown vaults through crude, unindexed file searches or external cloud databases that destroy structural link graphs (`[[wiki-links]]`), tags, and frontmatter taxonomy.

**Lith exists to provide a local-first, semantic knowledge engine that translates raw Markdown files into a structured, transactional knowledge representation for AI agents and local applications.**

Lith ensures that AI tools interact with connected semantic knowledge rather than unindexed raw text—without compromising file ownership, local privacy, or vault integrity.

---

## 2. Who Is It For?

Lith is built for:
* **Knowledge Management Power Users**: Individuals using Obsidian, logseq, or plain Markdown vaults who want AI capabilities that understand the deep interconnections within their notes.
* **AI Agent Developers & Tool Builders**: Engineers building local AI assistants, CLI utilities, or MCP integrations who need transactional, capability-based context retrieval instead of ad-hoc regex scripts.
* **Privacy-Conscious Developers & Enterprises**: Organizations and developers requiring local-first semantic understanding that runs entirely on local hardware without sending vault data to third-party indexing services.

---

## 3. What Problems Does Lith Solve?

Lith addresses four fundamental challenges:

1. **Context Fragmentation**: Converts raw text and isolated files into a rich, queryable graph of notes, blocks, tags, and relations.
2. **Brittle File Edits**: Prevents unvalidated regex rewrites by enforcing transactional, capability-driven update proposals before any modifications are applied.
3. **Derived State Drift**: Ensures that all indexes, graphs, and caches are strictly disposable derived state that can be rebuilt cleanly from raw Markdown files at any time.
4. **Interface Lock-In**: Exposes a unified core engine through equal peer interfaces (CLI, REST API, Model Context Protocol, and native Go SDK).

---

## 4. What Problems Will Lith Intentionally Never Solve?

To remain focused, coherent, and maintainable, Lith explicitly defines its non-goals. **Lith will intentionally NEVER:**

* **Be a Note Editor or UI App**: Lith is an engine, not a text editor. User interface tools like Obsidian, VS Code, or Neovim fulfill note-editing roles.
* **Be a Cloud Sync Service**: Lith leaves file synchronization, cloud backup, and multi-device transport to dedicated systems (Obsidian Sync, Syncthing, Git). Lith operates locally on the filesystem.
* **Be a Generic Web Scraper**: Lith is built specifically for Markdown and knowledge-vault semantics, not arbitrary web scraping or general media ingestion.
* **Directly Mutate Vault Files Unvalidated**: Lith never allows AI tools to perform unstructured, unverified writes on user Markdown files.
