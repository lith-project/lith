# Lith Project Roadmap

This document outlines the strategic, milestone-driven roadmap for **Lith**. Development is guided by architectural specifications (RFCs) and the constitution in [PROJECT_PRINCIPLES.md](PROJECT_PRINCIPLES.md).

---

## Milestones vs Epics

These are different things and are tracked differently:

```
Milestone  →  an outcome that is either reached or not reached
Epic       →  a collection of implementation work required to reach it
Issue      →  a discrete task inside an epic
```

Milestones are stable; epics evolve as we learn. A milestone is complete when its **Definition of Done** holds — not when its issues are closed.

Sub-milestones use **letters**, not decimals (`M1-B`, not `M1.1`), so a phase can be inserted later without renumbering everything after it.

---

## Milestone Map

```
M0  Project Foundation      Governance, principles, RFC system, engineering workflow
M1  Knowledge Engine        A. Lifecycle   B. Parse & Store   C. Query   D. Incremental Indexing
M2  Semantic Platform       A. Graph       B. Capabilities    C. Plugins
M3  Public Interfaces       A. CLI         B. REST            C. MCP     D. SDK
M4  Autonomous Operations   A. Planner     B. Jobs            C. Transactions   D. Refactoring Engine
```

M0 is organizational maturity. M1 onward is product capability. The boundary between them is deliberate and visible.

---

## M0: Project Foundation 🚧 *(Current)*

**Objective**: Establish governance, the RFC process, engineering workflow, and the architectural foundation that makes implementation almost inevitable.

**Deliverables**
- [x] Repository housekeeping & Apache 2.0 licensing
- [x] Immutable project principles ([PROJECT_PRINCIPLES.md](PROJECT_PRINCIPLES.md))
- [x] RFC framework, index, template, and lifecycle ([rfcs/README.md](rfcs/README.md))
- [x] Architecture map & documentation structure
- [x] Community templates & GitHub Discussions
- [x] Conformance-assertion requirement + acceptance gate ([rfcs/README.md](rfcs/README.md))
- [x] [Test vault specification](docs/testing/test-vault-spec.md)
- [ ] **RFC-0001** — Project Vision & Strategic Architecture *(draft complete, awaiting review)*
- [ ] **RFC-0002** — Domain Model & Vault AST
- [ ] **RFC-0003** — Storage Engine & State Rebuilds
- [ ] **RFC-0005** — Background Worker & Job Engine *(authored before 0004: it owns transaction semantics)*
- [ ] **RFC-0004** — Indexing & Link Graph Engine
- [ ] Capability catalog (`docs/reference/capability-catalog.md`), instantiating the model defined by RFC-0001
- [ ] Test vault corpus + goldens committed
- [ ] Epics and issues broken out from each accepted RFC

**Definition of Done**
1. RFC-0001 through RFC-0005 are all `Accepted`.
2. Every conformance assertion across those RFCs has a verification method and an owning milestone.
3. The capability catalog exists, and every capability **scheduled for M1** names an owning RFC. Capabilities beyond M1 may be catalogued without one — the catalog is a roadmap, and narrowing it to what is already specified would reduce it to a table of contents.
4. The test vault corpus is committed with goldens and passes the corpus-integrity check.

> No Go code is written until this milestone is done. That is the point of it.

---

## M1: Knowledge Engine

**Objective**: Prove the architectural core — observe a vault, understand it, store it, query it.

### M1-A · Lifecycle

Deliberately boring. Proves the service lifecycle and nothing else.

* Parse configuration
* Start the daemon
* Watch one vault
* Detect filesystem events
* Log them
* Exit cleanly

No parser. No SQLite. No graph. No MCP.

**Definition of Done**
1. Daemon starts from a config file and exits cleanly on signal.
2. Filesystem events for one vault are detected and logged.
3. Restart leaves no orphaned state or lock.
4. Conformance: [RFC-0001/C-1](rfcs/0001-project-vision.md#c-1-core-semantic-independence) and [C-6](rfcs/0001-project-vision.md#c-6-plugin-absence-safety) pass in CI.

*Candidate epics:* filesystem abstraction · watcher · event queue · debouncer · daemon lifecycle & signals · config loading · structured logging

### M1-B · Parse & Store

Filesystem → Markdown parser → SQLite metadata. Still no AI.

**Definition of Done**
1. Every note in the test vault corpus parses to its golden result.
2. A full index of the corpus completes with no unhandled error.
3. Deleting all derived state and re-indexing yields logically identical state — [RFC-0001/C-2](rfcs/0001-project-vision.md#c-2-rebuild-determinism).
4. No component outside the transaction coordinator writes to the vault — [RFC-0001/C-3](rfcs/0001-project-vision.md#c-3-single-write-path).

*Candidate epics:* markdown parser · frontmatter extraction · AST & domain model · block addressing · SQLite schema · transactional persistence · full rebuild

### M1-C · Query

`lith search ceph` returns correct results against an indexed vault. If this works, the architectural core is proven.

**Definition of Done**
1. Full-text and metadata search return correct, deterministic results — [RFC-0001/C-7](rfcs/0001-project-vision.md#c-7-deterministic-core-search).
2. Every exposed capability appears in the catalog — [RFC-0001/C-4](rfcs/0001-project-vision.md#c-4-capability-catalog-completeness).
3. The CLI reaches the engine only through the capability registry — [RFC-0001/C-5](rfcs/0001-project-vision.md#c-5-interface-adapter-purity).
4. Query latency on benchmark tier S is measured and recorded as the baseline.

*Candidate epics:* capability registry · full-text index · metadata filters · deterministic ranking · CLI adapter · result formatting

### M1-D · Incremental Indexing

Change a file; only what changed is re-indexed.

**Definition of Done**
1. Incremental index state is identical to a full re-index of the same vault.
2. `EC-DYN-*` scenarios pass, including identical `mtime` + size with changed content.
3. A missed or dropped event is recovered by the reconciliation scan.
4. Benchmark tier M sustains a single-file change without a full re-index.

*Candidate epics:* dirty set · incremental parser · recovery/reconciliation scanner · rename & move detection · event coalescing · benchmark suite

**M1 Definition of Done** — the foundation is correct when: the daemon starts; watches a local vault; survives restarts; handles thousands of files; maintains an index; CLI queries return in milliseconds; and a rebuild from scratch succeeds after deleting all derived state. Nothing else.

---

## M2: Semantic Platform

**Objective**: Extend understanding beyond a single note, and open the plugin seam.

* **M2-A · Graph** — link graph queries, backlinks, orphan and dangling-reference detection, entity resolution
* **M2-B · Capabilities** — capability surface beyond search; catalog promotion to `Stable`
* **M2-C · Plugins** — plugin host, loading, lifecycle, sandboxing; first optional capability (semantic search) proves the seam

Embeddings, vector storage, and semantic ranking enter here as plugins — never in the core ([RFC-0001 §4](rfcs/0001-project-vision.md)).

---

## M3: Public Interfaces

**Objective**: Expose the engine through peer adapters. Each is an adapter over the same capability registry; none is privileged.

* **M3-A · CLI** — full command surface
* **M3-B · REST** — local HTTP API
* **M3-C · MCP** — Model Context Protocol server
* **M3-D · SDK** — native Go SDK for embedded use

---

## M4: Autonomous Operations

**Objective**: Proactive knowledge maintenance and safe agent-driven change.

* **M4-A · Planner** — proactive analysis, orphan detection, continuous synthesis
* **M4-B · Jobs** — long-running and scheduled operations
* **M4-C · Transactions** — agent-proposed change with validation and approval
* **M4-D · Refactoring Engine** — safe semantic modification of vault content

---

## References

- [RFC-0001: Project Vision & Strategic Architecture](rfcs/0001-project-vision.md)
- [RFC Index](rfcs/index.md) · [RFC Process](rfcs/README.md)
- [Test Vault Specification](docs/testing/test-vault-spec.md)
- [PROJECT_PRINCIPLES.md](PROJECT_PRINCIPLES.md) · [VISION.md](VISION.md) · [ARCHITECTURE.md](ARCHITECTURE.md)
