# Lith Project Roadmap

This document outlines the strategic, milestone-driven roadmap for **Lith**. Development is guided by architectural specifications (RFCs) and the constitution in [PROJECT_PRINCIPLES.md](PROJECT_PRINCIPLES.md).

---

## Milestones vs Epics

These are different things and are tracked differently:

```text
Milestone  →  an outcome that is either reached or not reached
Epic       →  a collection of implementation work required to reach it
Issue      →  a discrete task inside an epic
```

Milestones are stable; epics evolve as we learn. A milestone is complete when its **Definition of Done** holds — not when its issues are closed.

Sub-milestones use **letters**, not decimals (`M1-B`, not `M1.1`), so a phase can be inserted later without renumbering everything after it.

---

## Milestone Map

```text
M0  Project Foundation      Governance, principles, RFC system, engineering workflow
M1  Knowledge Engine        A. Lifecycle   B. Parse & Store   C. Query   D. Incremental Indexing
M2  Semantic Platform       A. Graph       B. Capabilities    C. Plugins
M3  Public Interfaces       A. CLI         B. REST            C. MCP     D. SDK
M4  Autonomous Operations   A. Planner     B. Jobs            C. Transactions   D. Refactoring Engine
```

M0 is organizational maturity. M1 onward is product capability. The boundary between them is deliberate and visible.

---

## M0: Project Foundation ✅ *(Complete)*

**Objective**: Establish governance, the RFC process, engineering workflow, and the architectural foundation that makes implementation almost inevitable.

**Deliverables**

- [x] Repository housekeeping & Apache 2.0 licensing
- [x] Immutable project principles ([PROJECT_PRINCIPLES.md](PROJECT_PRINCIPLES.md))
- [x] RFC framework, index, template, and lifecycle ([rfcs/README.md](rfcs/README.md))
- [x] Architecture map & documentation structure
- [x] Community templates & GitHub Discussions
- [x] Conformance-assertion requirement + acceptance gate ([rfcs/README.md](rfcs/README.md))
- [x] [Test vault specification](docs/testing/test-vault-spec.md)
- [x] **RFC-0001** — Project Vision & Strategic Architecture — `Accepted`
- [x] **RFC-0002** — Domain Model & Vault AST — `Accepted`
- [x] **RFC-0003** — Storage Engine & State Rebuilds — `Accepted`
- [x] **RFC-0005** — Background Worker & Job Engine — `Accepted` *(authored before 0004: it owns transaction semantics)*
- [x] **RFC-0004** — Indexing & Link Graph Engine — `Accepted`
- [x] Capability catalog (`docs/reference/capability-catalog.md`), instantiating the model defined by RFC-0001
- [x] Test vault corpus committed — 96 edge-case files, byte-integrity-verified manifest ([tests/vault/](tests/vault/))
- [x] Epics and issues broken out for the next milestone to begin — M1-A, tracked in [#14](https://github.com/lith-project/lith/issues/14)

> **On the epic breakout.** This was originally scoped as "epics and issues broken out from *each* accepted RFC". It is deliberately narrowed to *the next milestone to begin*, and repeated at the start of each subsequent milestone. Writing M1-D's issues today means writing them before M1-A, M1-B, and M1-C have taught us anything — they would be rewritten, and in the meantime they would look like decisions rather than guesses. The RFCs are the durable artefact; issues are the disposable plan for the work immediately in front of us.

**Definition of Done**

1. RFC-0001 through RFC-0005 are all `Accepted`.
2. Every conformance assertion across those RFCs has a verification method and an owning milestone.
3. The capability catalog exists, and every capability **scheduled for M1** names an owning RFC. Capabilities beyond M1 may be catalogued without one — the catalog is a roadmap, and narrowing it to what is already specified would reduce it to a table of contents.
4. The test vault corpus is committed and passes the corpus-integrity check. Goldens are an M1-B deliverable, produced against a reference parser rather than by hand — see [tests/vault/README.md](tests/vault/README.md).

> No Go code is written until this milestone is done. That is the point of it.

> **Completion verified 2026-08-06.** [PR #22](https://github.com/lith-project/lith/pull/22)
> closed M0 after all four Definition of Done items were evidenced. The accepted
> RFC index, capability catalog, and corpus remain present on `main`, and the
> corpus-integrity check still passes.

---

## M1: Knowledge Engine 🚧 *(Current)*

**Objective**: Prove the architectural core — observe a vault, understand it, store it, query it.

### M1-A · Lifecycle ✅ *(Complete)*

Deliberately boring. Proves the service lifecycle and nothing else.

- Parse configuration
- Start the daemon
- Watch one vault
- Detect filesystem events
- Log them
- Exit cleanly

No parser. No SQLite. No graph. No MCP.

**Definition of Done**

1. Daemon starts from a config file and exits cleanly on signal.
2. Filesystem events for one vault are detected and logged.
3. Restart leaves no orphaned state or lock.
4. Conformance: [RFC-0001/C-1](rfcs/0001-project-vision.md#c-1-core-semantic-independence) and [C-6](rfcs/0001-project-vision.md#c-6-plugin-absence-safety) pass in CI.

**Epics** — tracked in [#14](https://github.com/lith-project/lith/issues/14), in build order:

[#23](https://github.com/lith-project/lith/issues/23) Go project bootstrap · [#24](https://github.com/lith-project/lith/issues/24) CI pipeline · [#25](https://github.com/lith-project/lith/issues/25) corpus test harness · [#15](https://github.com/lith-project/lith/issues/15) config loading · [#16](https://github.com/lith-project/lith/issues/16) structured logging · [#17](https://github.com/lith-project/lith/issues/17) filesystem path identity · [#18](https://github.com/lith-project/lith/issues/18) filesystem watcher · [#19](https://github.com/lith-project/lith/issues/19) debouncer · [#20](https://github.com/lith-project/lith/issues/20) event queue · [#21](https://github.com/lith-project/lith/issues/21) daemon lifecycle & signals

Path identity is its own epic, ahead of the watcher, and was not in the original candidate list. Building the M0 corpus hit Unicode NFC/NFD corruption twice in one sitting — silently, in tooling — and every note identity in M1-B is built on it.

**Completion evidence** — verified 2026-08-06:

- All 37 implementation leaves, [#27](https://github.com/lith-project/lith/issues/27) through [#63](https://github.com/lith-project/lith/issues/63), are closed.
- [PR #109](https://github.com/lith-project/lith/pull/109) proves the lifecycle through the real `lithd` binary; [PR #110](https://github.com/lith-project/lith/pull/110) removes the shutdown-log race found during the milestone audit.
- Go CI, lint/security, and architectural-conformance workflows passed on both PR heads. A fresh `origin/main` archive also passes build, tests with the race detector, conformance, and corpus integrity locally.

> **Transition cleanup.** The product milestone is complete by its Definition of
> Done, but GitHub metadata has not caught up: tracker [#14](https://github.com/lith-project/lith/issues/14)
> and epics #15–#21 and #23–#25 remain open, while closed leaves #27–#36 and #63
> retain stale `agent:wip` claims. Close the completed trackers and clear those
> claims before enabling M1-B task selection.

### M1-B · Parse & Store 🚧 *(Next)*

Filesystem → Markdown parser → SQLite metadata. Still no AI.

**Definition of Done**

1. Per-note parse goldens are generated for the committed corpus and reviewed as a diff — carried over from M0, where they were deliberately not hand-written ([tests/vault/README.md](tests/vault/README.md)).
2. Every note in the test vault corpus parses to its golden result.
3. A full index of the corpus completes with no unhandled error.
4. Deleting all derived state and re-indexing yields logically identical state — [RFC-0001/C-2](rfcs/0001-project-vision.md#c-2-rebuild-determinism).
5. No component outside the transaction coordinator writes to the vault — [RFC-0001/C-3](rfcs/0001-project-vision.md#c-3-single-write-path).

*Candidate epics:* markdown parser · frontmatter extraction · AST & domain model · block addressing · parse goldens · SQLite schema · transactional persistence · full rebuild

**Next development steps**

1. Complete the M1-A transition cleanup above, then create the M1-B GitHub milestone and tracker. No M1-B issues exist yet as of 2026-08-06.
2. Break the candidate epics into claimable leaf tasks. Encode file overlap and exported-contract dependencies in `blocked_by`; the GitHub graph, not this prose, determines concurrency and build order.
3. Establish the RFC-0002 domain model and parser path, then generate and review per-note goldens from the committed corpus.
4. Implement the RFC-0003 SQLite schema, canonical dump, transactional persistence, and full rebuild against those reviewed parse results.
5. Finish with a real-daemon corpus index and destructive-rebuild scenario proving logical identity, vault isolation, and the single-write-path constraint.

This is the milestone entry sequence, not a substitute task list. Once the issue
graph exists, `tools/next-task.sh M1-B` is the only source for claimable work.

### M1-C · Query

`lith search ceph` returns correct results against an indexed vault. If this works, the architectural core is proven.

**Definition of Done**

1. Full-text and metadata search return correct, deterministic results — [RFC-0001/C-7](rfcs/0001-project-vision.md#c-7-deterministic-core-search).
2. Every exposed capability appears in the catalog — [RFC-0001/C-4](rfcs/0001-project-vision.md#c-4-capability-catalog-completeness).
3. The CLI reaches the engine only through the capability registry — [RFC-0001/C-5](rfcs/0001-project-vision.md#c-5-interface-adapter-purity), enforced mechanically by the boundary checker once `internal/adapter/` and `cmd/lith` exist — [RFC-0006/C-3](rfcs/0006-package-layout.md#c-3-adapter-purity). M1-A deliberately builds neither subject package, so C-3 is carried forward to here.
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

- **M2-A · Graph** — link graph queries, backlinks, orphan and dangling-reference detection, entity resolution
- **M2-B · Capabilities** — capability surface beyond search; catalog promotion to `Stable`
- **M2-C · Plugins** — plugin host, loading, lifecycle, sandboxing; first optional capability (semantic search) proves the seam. Plugin purity — [RFC-0006/C-7](rfcs/0006-package-layout.md#c-7-plugin-purity) — is enforced by the boundary checker from here, the milestone that first creates `internal/plugin/`

Embeddings, vector storage, and semantic ranking enter here as plugins — never in the core ([RFC-0001 §4](rfcs/0001-project-vision.md)).

---

## M3: Public Interfaces

**Objective**: Expose the engine through peer adapters. Each is an adapter over the same capability registry; none is privileged.

- **M3-A · CLI** — full command surface
- **M3-B · REST** — local HTTP API
- **M3-C · MCP** — Model Context Protocol server
- **M3-D · SDK** — native Go SDK for embedded use

---

## M4: Autonomous Operations

**Objective**: Proactive knowledge maintenance and safe agent-driven change.

- **M4-A · Planner** — proactive analysis, orphan detection, continuous synthesis
- **M4-B · Jobs** — long-running and scheduled operations
- **M4-C · Transactions** — agent-proposed change with validation and approval
- **M4-D · Refactoring Engine** — safe semantic modification of vault content

---

## References

- [RFC-0001: Project Vision & Strategic Architecture](rfcs/0001-project-vision.md)
- [RFC Index](rfcs/index.md) · [RFC Process](rfcs/README.md)
- [Test Vault Specification](docs/testing/test-vault-spec.md)
- [PROJECT_PRINCIPLES.md](PROJECT_PRINCIPLES.md) · [VISION.md](VISION.md) · [ARCHITECTURE.md](ARCHITECTURE.md)
