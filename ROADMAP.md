# Lith Project Roadmap

This document outlines the strategic, milestone-driven roadmap for **Lith**. Development is guided by architectural specifications (RFCs) and core principles documented in [PROJECT_PRINCIPLES.md](PROJECT_PRINCIPLES.md).

---

## Milestone Milestones

```
+-----------------------------------------------------------------------+
| M0 Foundation       -> Governance, RFC Specs, Repository Infra        |
| M1 Knowledge Engine -> Vault Indexing, Graph Relations, Disposable State|
| M2 Semantic Platform-> Query Capabilities, Embeddings Plugin Architecture|
| M3 AI Interfaces    -> CLI, REST API, MCP Protocol, SDK Integration   |
| M4 Autonomous       -> Continuous Understanding, Proactive Workflows  |
+-----------------------------------------------------------------------+
```

---

### M0: Foundation 🚧 *(Current Milestone)*
* **Objective**: Establish project governance, RFC specification process, repository infrastructure, and architectural foundation.
* **Deliverables**:
  - [x] Repository Housekeeping & Apache 2.0 Licensing
  - [x] Immutable Project Principles (`PROJECT_PRINCIPLES.md`)
  - [x] RFC Framework & [RFC-0001 (Project Vision & Strategic Architecture)](https://github.com/lith-project/lith/issues/3)
  - [x] Architecture map & documentation structure (`docs/architecture`, `docs/diagrams`, `docs/adr`)
  - [x] Community templates & GitHub Discussions enablement
  - [ ] RFC-0002 (Domain Model & Vault AST Representation)
  - [ ] RFC-0003 (Link Graph & Transactional Indexing Lifecycle)

---

### M1: Knowledge Engine
* **Objective**: Build the core engine for observing local Markdown vaults, parsing link structures, and building disposable graph indexes.
* **Deliverables**:
  - Vault file system observer & change detection
  - Markdown AST parsing & frontmatter extraction
  - Wiki-link (`[[link]]`), block embed, and tag graph builder
  - Transactional local index storage & full rebuild capabilities
  - Core Go project structure and unit test suite

---

### M2: Semantic Platform
* **Objective**: Expand graph capabilities and introduce extensible plugin interfaces for semantic understanding.
* **Deliverables**:
  - Graph query engine & entity resolution API
  - Plugin architecture for optional embedding generation & vector search
  - Local caching strategies and background worker processing
  - Vault integrity auditing & dangling reference detection

---

### M3: AI Interfaces
* **Objective**: Expose peer interfaces allowing AI agents, developer tooling, and client applications to interact seamlessly with Lith.
* **Deliverables**:
  - Command-line interface (`lith CLI`)
  - Local HTTP REST API
  - Model Context Protocol (MCP) server integration
  - Native Go SDK for embedded applications
  - Transactional capability API for proposed vault updates

---

### M4: Autonomous Workflows
* **Objective**: Enable proactive knowledge maintenance, multi-agent collaboration, and enterprise vault integration.
* **Deliverables**:
  - Proactive background knowledge analysis and orphan detection
  - Multi-vault linking & domain boundaries
  - External event streaming & webhook notifications
  - Advanced agent workflows & continuous vault synthesis
