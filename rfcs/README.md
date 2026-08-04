# Lith Requests for Comments (RFCs)

The `rfcs/` directory contains all architectural specifications and design documents for **Lith**. 

Lith follows an RFC-driven development workflow to ensure that architectural changes are thoroughly designed, reviewed, and aligned with [PROJECT_PRINCIPLES.md](../PROJECT_PRINCIPLES.md) before implementation begins.

---

## Engineering Hierarchy

```
PROJECT_PRINCIPLES.md
        │  (Constitution)
        ▼
RFCs (architecture)
        │  (Specifications & Boundaries)
        ▼
Issues (work)
        │  (Discrete Actionable Tasks)
        ▼
Pull Requests (implementation)
        │  (Code & Validation)
        ▼
Code
```

---

## RFC Lifecycle

```
Idea ──> Discussion ──> RFC Draft ──> Pull Request ──> Review ──> Accepted ──> Implementation Issue(s) ──> Implementation PR(s) ──> Released
```

1. **Idea & Discussion**: Start a discussion in [GitHub Discussions](https://github.com/lith-project/lith/discussions) under `Architecture` or `RFCs`.
2. **RFC Draft**: Copy [rfcs/templates/rfc-template.md](templates/rfc-template.md) to `rfcs/NNNN-feature-name.md` using a 4-digit RFC number.
3. **Pull Request**: Open a PR submitting the RFC draft.
4. **Review & Iteration**: Community and maintainers review the RFC design.
5. **Accepted**: Maintainers merge the PR and update status to `Accepted`.
6. **Implementation**: Actionable GitHub Issues are created referencing the accepted RFC.

---

## When an RFC is Required

An RFC is **mandatory** for:
* New architectural components or subsystems
* Storage engine, schema, or indexing model changes
* Public API or protocol modifications (CLI, REST, MCP, SDK)
* Plugin architecture or extension hook changes
* Job engine or background worker lifecycle changes
* Changes to transactional semantics or state rebuild models

An RFC is **NOT required** for:
* Bug fixes
* Documentation or README updates
* Unit or integration test additions
* Internal refactoring that does not alter architectural boundaries
* Minor dependency updates
* Implementation details already specified in an accepted RFC

---

## RFC Rules & Standards

### 1. Four-Digit Numbering
RFC numbers are always 4 digits (`0001`, `0002`, `0003`, ..., `0128`). Numbers are sequential and assigned upon draft submission.

### 2. Machine-Readable Frontmatter
Every RFC file must begin with a YAML frontmatter block:

```yaml
---
rfc: "0001"
title: "Project Vision & Core Architecture"
status: Draft
authors:
  - Author Name <email@example.com>
created: 2026-08-04
updated: 2026-08-04
discussion: https://github.com/lith-project/lith/discussions/
supersedes:
superseded_by:
---
```

### 3. RFC Status Values
* **`Draft`**: Proposal under active community review.
* **`Accepted`**: Approved architectural design ready for implementation issues.
* **`Implemented`**: Feature fully built and released.
* **`Deprecated`**: Design is obsolete and no longer recommended.
* **`Superseded`**: Replaced by a newer RFC (link specified in `superseded_by`).
* **`Rejected`**: Proposal declined after review.

> [!IMPORTANT]
> **Never delete RFCs.** Even rejected, deprecated, or superseded RFCs remain in the repository as a historical record of architectural decisions and trade-offs.

---

## Index of RFCs

| RFC # | Title | Status | Authors |
| ----- | ----- | ------ | ------- |
| [0001](0001-project-vision.md) | [Project Vision & Core Architecture](0001-project-vision.md) | Draft | Lith Maintainers |
| [0002](0002-domain-model.md) | Domain Model & Vault AST | Draft | Pending |
| [0003](0003-storage-engine.md) | Storage Engine & State Rebuilds | Draft | Pending |
| [0004](0004-indexing.md) | Indexing & Link Graph Engine | Draft | Pending |
| [0005](0005-job-engine.md) | Background Worker & Job Engine | Draft | Pending |
