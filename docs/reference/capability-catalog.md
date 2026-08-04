# Lith Capability Catalog

**Status:** Draft · **Model owner:** [RFC-0001 §3](../../rfcs/0001-project-vision.md) · **Updated:** 2026-08-04

This is the living inventory of Lith's **capabilities** — the domain-level semantic operations exposed to external callers. It is the project's public API roadmap and its unit of product scope.

> [!IMPORTANT]
> This document lists **which capabilities exist**. It does not define **what a capability is** — that is [RFC-0001 §3](../../rfcs/0001-project-vision.md), which owns the model, the metadata schema, the lifecycle, and the identifier scheme. Changing the *model* amends that RFC. Adding or promoting a capability is a change to this file.

---

## How to Read This Catalog

Every capability has a stable identifier `CAP-NNNN`, assigned once and never reused. RFCs, issues, and code reference the **identifier**, never the display name, so renaming a capability breaks nothing.

**Naming rule.** A capability names an outcome, not a mechanism. `Search` is a capability; `SQLite FTS Search` is not — storage technology belongs to [RFC-0003](../../rfcs/0003-storage-engine.md), and naming it here would make the public surface hostage to an implementation detail. No entry below names a database, a file format, or a library.

**Type** is an architectural commitment, not a label:

| Type | Meaning |
| ---- | ------- |
| **Core** | Works in a default installation with zero plugins loaded |
| **Plugin** | Optional by construction; its absence is never an error ([RFC-0001/C-6](../../rfcs/0001-project-vision.md#c-6-plugin-absence-safety)) |
| **Future** | Recognized as desirable; no committed design |

**Status** follows the capability lifecycle in [RFC-0001 §3](../../rfcs/0001-project-vision.md): `Proposed → Accepted → Implemented → Stable → Deprecated → Removed`. This is *not* the RFC status enum — an RFC is an architectural decision that reaches a terminal state; a capability is a product feature that keeps moving after that decision.

**Every capability below is currently `Proposed`.** No RFC is `Accepted` and no code exists. That will read as repetitive; it is also the honest state of the project today.

---

## Summary

| ID | Name | Type | MVP | Owner RFC | Milestone | Status |
| -- | ---- | ---- | --- | --------- | --------- | ------ |
| [CAP-0001](#cap-0001--search) | Search | Core | ✅ | [0003](../../rfcs/0003-storage-engine.md) | M1-C | Proposed |
| [CAP-0002](#cap-0002--metadata) | Metadata | Core | ✅ | [0002](../../rfcs/0002-domain-model.md) | M1-C | Proposed |
| [CAP-0003](#cap-0003--graph) | Graph | Core | ❌ | [0004](../../rfcs/0004-indexing.md) | M2-A | Proposed |
| [CAP-0004](#cap-0004--jobs) | Jobs | Core | ❌ | [0005](../../rfcs/0005-job-engine.md) | M4-B | Proposed |
| [CAP-0005](#cap-0005--transactions) | Transactions | Core | ❌ | [0005](../../rfcs/0005-job-engine.md) | M4-C | Proposed |
| [CAP-0006](#cap-0006--tasks) | Tasks | Core | ❌ | *unassigned* | M2-B | Proposed |
| [CAP-0007](#cap-0007--refactor) | Refactor | Core | ❌ | *unassigned* | M4-D | Proposed |
| [CAP-0008](#cap-0008--semantic-search) | Semantic Search | Plugin | ❌ | *unassigned* | M2-C | Proposed |
| [CAP-0009](#cap-0009--canvas) | Canvas | Future | ❌ | *unassigned* | — | Proposed |
| [CAP-0010](#cap-0010--bases) | Bases | Future | ❌ | *unassigned* | — | Proposed |

**The MVP surface is CAP-0001 and CAP-0002.** Nothing else ships in the first usable release.

---

## CAP-0001 — Search

```yaml
id: CAP-0001
name: Search
status: Proposed
type: Core
category: Knowledge
mvp: true
owner_rfc: "0003"
milestone: M1-C
depends_on: []
experimental: false
```

Find notes matching a query. One capability, multiple providers behind a single contract.

**MVP scope is full-text and metadata filtering with deterministic ranking.** Semantic, vector, graph, and hybrid retrieval are separate capabilities, explicitly outside the MVP ([RFC-0001 §4](../../rfcs/0001-project-vision.md)). Semantic search is not "better search" — it is another provider of this same capability, and the core must not be able to tell whether it is present.

**Search is not retrieval.** Search answers *find matching notes*. Retrieval answers *return the most useful context*. The core owns search; plugins may enhance retrieval.

Results are deterministic: identical corpus plus identical query yields an identically ordered result set, with no dependency on a model, an embedding, or floating-point similarity ([RFC-0001/C-7](../../rfcs/0001-project-vision.md#c-7-deterministic-core-search)).

**Non-goals:** ranking tuned to an individual user; natural-language query understanding; cross-vault search.

---

## CAP-0002 — Metadata

```yaml
id: CAP-0002
name: Metadata
status: Proposed
type: Core
category: Knowledge
mvp: true
owner_rfc: "0002"
milestone: M1-C
depends_on: []
experimental: false
```

Read frontmatter, tags, aliases, sections, and structural facts about notes, together with the diagnostics produced while parsing them.

Includes the vault's own defects as first-class answers: broken links, ambiguous references, malformed frontmatter. A knowledge engine that silently discards what it could not understand is less useful than one that reports it.

**Non-goals:** writing metadata — that is [CAP-0005](#cap-0005--transactions); imposing a frontmatter schema, because the user's vocabulary is the user's.

---

## CAP-0003 — Graph

```yaml
id: CAP-0003
name: Graph
status: Proposed
type: Core
category: Knowledge
mvp: false
owner_rfc: "0004"
milestone: M2-A
depends_on: ["CAP-0002"]
experimental: false
```

Traverse relationships between notes: forward links, backlinks, orphans, dangling references, and ambiguous references.

The index making this possible is built during M1 ([RFC-0004](../../rfcs/0004-indexing.md)); the *exposed* capability is M2. Building the graph and offering it as a public operation are separate commitments, and the second needs a query surface this catalog does not yet describe.

**Non-goals:** a general-purpose graph query language; algorithms beyond neighbourhood traversal (centrality, clustering, community detection).

---

## CAP-0004 — Jobs

```yaml
id: CAP-0004
name: Jobs
status: Proposed
type: Core
category: Operations
mvp: false
owner_rfc: "0005"
milestone: M4-B
depends_on: []
experimental: false
```

Observe and control long-running work: progress, cancellation, and outcome of rebuilds, scans, and index batches.

The job engine itself is M1 infrastructure ([RFC-0005](../../rfcs/0005-job-engine.md)) — the daemon cannot function without it. This entry is the *externally exposed* surface, which arrives much later. Internal machinery and public capability are deliberately not the same milestone.

**Non-goals:** user-defined or scriptable jobs; scheduling as a user-facing feature.

---

## CAP-0005 — Transactions

```yaml
id: CAP-0005
name: Transactions
status: Proposed
type: Core
category: Mutation
mvp: false
owner_rfc: "0005"
milestone: M4-C
depends_on: ["CAP-0002"]
experimental: false
```

Propose a validated change to vault content and have it applied atomically, or rejected without touching a byte.

This is the only capability that writes to the vault, and the only path by which an agent may cause a file to change ([RFC-0001/C-3](../../rfcs/0001-project-vision.md#c-3-single-write-path)). A proposal is structurally validated, checked against a content-hash precondition, applied to a buffer, re-parsed, and rejected if the result introduces new errors — all before anything reaches disk.

It cannot express reformatting, because [RFC-0002](../../rfcs/0002-domain-model.md) leaves no serializer with which to express it.

**Non-goals:** approval workflow and agent policy — mechanism is specified, policy is deferred; multi-vault transactions.

---

## CAP-0006 — Tasks

```yaml
id: CAP-0006
name: Tasks
status: Proposed
type: Core
category: Knowledge
mvp: false
owner_rfc: unassigned
milestone: M2-B
depends_on: ["CAP-0002"]
experimental: false
```

Query and reason about task items across the vault: open, done, scheduled, recurring, and the long tail of plugin-specific states.

Task *parsing* is specified in [RFC-0002 §8](../../rfcs/0002-domain-model.md), which deliberately hardcodes no plugin's vocabulary — alternative markers are preserved as `Other` so this capability can interpret them later without a parser change. The capability itself has no owning RFC yet.

**Non-goals:** task mutation, which belongs to [CAP-0005](#cap-0005--transactions); imposing a task syntax on the user.

---

## CAP-0007 — Refactor

```yaml
id: CAP-0007
name: Refactor
status: Proposed
type: Core
category: Mutation
mvp: false
owner_rfc: unassigned
milestone: M4-D
depends_on: ["CAP-0003", "CAP-0005"]
experimental: false
```

Semantic modification across many notes at once: rename a concept and update every reference, merge duplicate notes, split an overgrown note, repair a class of broken links.

Depends on both the graph (to know what a change touches) and transactions (to apply it safely). It is the capability that most obviously requires multi-file atomicity, which is why [RFC-0005 §7](../../rfcs/0005-job-engine.md) specifies an intent journal rather than claiming a guarantee POSIX does not provide.

**Non-goals:** unattended refactoring without review; content generation.

---

## CAP-0008 — Semantic Search

```yaml
id: CAP-0008
name: Semantic Search
status: Proposed
type: Plugin
category: Knowledge
mvp: false
owner_rfc: unassigned
milestone: M2-C
depends_on: ["CAP-0001"]
experimental: true
```

Embedding-based retrieval, offered as an alternative provider behind the [CAP-0001](#cap-0001--search) contract.

Explicitly a plugin, and explicitly outside the MVP. The core engine never depends on embeddings, vector databases, or semantic models — this is the proposed ninth principle in [RFC-0001 §4](../../rfcs/0001-project-vision.md), and this entry is the capability that principle keeps at arm's length. Its absence is never an error.

**Non-goals:** becoming a core dependency; fixing a storage or model strategy in advance of a design.

---

## CAP-0009 — Canvas

```yaml
id: CAP-0009
name: Canvas
status: Proposed
type: Future
category: Knowledge
mvp: false
owner_rfc: unassigned
milestone: unscheduled
depends_on: ["CAP-0003"]
experimental: false
```

Understanding of canvas documents as structured, linked knowledge rather than opaque files.

Today canvases are `Asset`s with a declared kind ([RFC-0002 §1](../../rfcs/0002-domain-model.md)) — recognized, preserved, uninterpreted. Promoting them to parsed entities requires a domain model this project does not yet have.

---

## CAP-0010 — Bases

```yaml
id: CAP-0010
name: Bases
status: Proposed
type: Future
category: Knowledge
mvp: false
owner_rfc: unassigned
milestone: unscheduled
depends_on: ["CAP-0002"]
experimental: false
```

Understanding of base definitions as queryable structure.

As with [CAP-0009](#cap-0009--canvas), currently an `Asset` kind: carried, not understood.

---

## Deliberately Absent

Things that look like capabilities and are not:

| Not a capability | Why |
| ---------------- | --- |
| Indexing | Internal machinery. Users do not ask Lith to index; they ask questions and expect current answers. |
| Parsing | Internal. Its outputs surface through [CAP-0002](#cap-0002--metadata). |
| Storage / Rebuild | Derived-state management. Disposable by principle; not an operation offered to callers. |
| Watching | An optimization ([RFC-0004](../../rfcs/0004-indexing.md)), not a promise. |
| Sync | An explicit project non-goal ([VISION.md](../../VISION.md)). |
| Editing | Lith is not an editor. Mutation is [CAP-0005](#cap-0005--transactions), and it is validated and transactional. |

---

## Adding or Changing a Capability

1. Adding a capability, or promoting one along the lifecycle, is a PR against this file that names the owning RFC.
2. Changing the *model* — metadata schema, lifecycle, identifier scheme — amends [RFC-0001](../../rfcs/0001-project-vision.md).
3. Identifiers are never reused. A removed capability keeps its `CAP-NNNN` and its entry, marked `Removed`.
4. Every capability exposed by any interface must appear here ([RFC-0001/C-4](../../rfcs/0001-project-vision.md#c-4-capability-catalog-completeness)); interfaces must not expose operations absent from this catalog.

---

## Open Questions

- [ ] **Five capabilities have no owning RFC** (CAP-0006 through CAP-0010). The M0 definition of done in [ROADMAP.md](../../ROADMAP.md) states that *every* capability names an owning RFC, which this catalog does not satisfy and cannot until those RFCs exist. Either the catalog narrows to owned capabilities and loses its value as a roadmap, or the M0 criterion narrows to capabilities scheduled for M1. *Recommend the latter.*
- [ ] Is `category` a fixed vocabulary or free text? Used here as `Knowledge`, `Operations`, `Mutation` without being declared normatively anywhere.
- [ ] Should `Jobs` and `Transactions` split into read and write capabilities? Observing a job and cancelling one are quite different commitments.

---

## References

- [RFC-0001: Project Vision & Strategic Architecture](../../rfcs/0001-project-vision.md) — defines the capability model
- [RFC-0002: Domain Model & Vault AST](../../rfcs/0002-domain-model.md)
- [RFC-0003: Storage Engine & State Rebuilds](../../rfcs/0003-storage-engine.md)
- [RFC-0004: Indexing & Link Graph Engine](../../rfcs/0004-indexing.md)
- [RFC-0005: Background Worker & Job Engine](../../rfcs/0005-job-engine.md)
- [docs/glossary.md](../glossary.md) · [ROADMAP.md](../../ROADMAP.md) · [VISION.md](../../VISION.md)
