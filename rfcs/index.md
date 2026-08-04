# Lith RFC Index

This document is the official index of all Requests for Comments (RFCs) in **Lith**. 

For RFC lifecycle guidelines, metadata specifications, and contribution workflows, see [README.md](README.md).

---

## Index Table

| RFC # | Title | Status | Milestone | Subsystems | Requires |
| ----- | ----- | ------ | --------- | ------------ | -------- |
| [0001](0001-project-vision.md) | [Project Vision & Strategic Architecture](0001-project-vision.md) | Accepted | M0 | Core Engine, Architecture | - |
| [0002](0002-domain-model.md) | [Domain Model & Vault AST](0002-domain-model.md) | Accepted | M1 | Core Engine, Parsing | 0001 |
| [0003](0003-storage-engine.md) | [Storage Engine & State Rebuilds](0003-storage-engine.md) | Accepted | M1 | Storage | 0001, 0002 |
| [0004](0004-indexing.md) | [Indexing & Link Graph Engine](0004-indexing.md) | Accepted | M1 | Graph, Indexing | 0001, 0002, 0003, 0005 |
| [0005](0005-job-engine.md) | [Background Worker & Job Engine](0005-job-engine.md) | Accepted | M1 | Jobs, Transactions | 0001, 0002, 0003 |

---

## Specification Order

RFCs are **authored** in dependency order, which is not numeric order:

```
0001 → 0002 → 0003 → 0005 → 0004
```

RFC-0004 (incremental indexing) consumes the transaction semantics defined by RFC-0005, so RFC-0005 is specified first. Numbering is stable and never reassigned.

Implementation begins only when all five are `Accepted`.

---

## Filtered Views

### By Milestone

#### Milestone M0 (Foundation)
- [RFC-0001: Project Vision & Strategic Architecture](0001-project-vision.md)

#### Milestone M1 (Knowledge Engine)
- [RFC-0002: Domain Model & Vault AST](0002-domain-model.md)
- [RFC-0003: Storage Engine & State Rebuilds](0003-storage-engine.md)
- [RFC-0004: Indexing & Link Graph Engine](0004-indexing.md)
- [RFC-0005: Background Worker & Job Engine](0005-job-engine.md)

---

### By Status

#### Accepted
- [RFC-0001](0001-project-vision.md), [RFC-0002](0002-domain-model.md), [RFC-0003](0003-storage-engine.md), [RFC-0004](0004-indexing.md), [RFC-0005](0005-job-engine.md)
