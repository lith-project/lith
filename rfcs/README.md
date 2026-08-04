# Lith Requests for Comments (RFCs)

The `rfcs/` directory contains all architectural specifications and design documents for **Lith**. 

Lith follows an RFC-driven development workflow to ensure that architectural changes are thoroughly designed, reviewed, and aligned with [PROJECT_PRINCIPLES.md](../PROJECT_PRINCIPLES.md) before implementation begins.

---

## Engineering Hierarchy

```
PROJECT_PRINCIPLES.md
        │  (Immutable Constitution — Amending requires an RFC)
        ▼
RFCs (architecture)
        │  (Specifications & Subsystem Boundaries)
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

## RFC Index

For a full table of all RFCs filtered by status, milestone, and capability, see **[index.md](index.md)**.

---

## RFC Lifecycle

```
Idea ──> Discussion ──> RFC Draft ──> Pull Request ──> Review ──> Accepted ──> Implementation Issue(s) ──> Implementation PR(s) ──> Released
```

1. **Idea & Discussion**: Start a discussion in [GitHub Discussions](https://github.com/lith-project/lith/discussions) under `Architecture` or `RFCs`.
2. **RFC Draft**: Copy [rfcs/templates/rfc-template.md](templates/rfc-template.md) to `rfcs/NNNN-feature-name.md` using a 4-digit RFC number.
3. **Pull Request**: Open a PR submitting the RFC draft.
4. **Review & Iteration**: Community and maintainers review the RFC design (`status: Review`).
5. **Accepted**: Maintainers merge the PR and update status to `Accepted`. Gated by the [Acceptance Gate](#acceptance-gate) below.
6. **Implementation**: Actionable GitHub Issues are created referencing the accepted RFC and the specific conformance assertions they satisfy.

---

## Acceptance Gate

> [!IMPORTANT]
> **An RFC cannot move to `Accepted` unless every Conformance assertion is testable.**

Every RFC carries a `Conformance` section: numbered assertions (`C-1`, `C-2`, …) stating observable, verifiable requirements. Assertions are referenced project-wide as `RFC-000X/C-N`.

This exists so implementation review is mechanical rather than a matter of taste. A reviewer does not ask *"does this look right?"* — they ask *"which assertion does this violate?"*. A coding agent can produce compiling, passing code that quietly abandons the intended architecture; conformance assertions are what make that detectable.

**Rules:**

1. Every assertion states an outcome observable by a test, a static/CI check, a benchmark threshold, or a documented manual procedure. Prose that cannot be observed belongs in *Proposed Design*, not in *Conformance*.
2. Every assertion names a Verification method and an owning milestone.
3. Assertion identifiers are stable — never renumbered, never reused. Retired assertions remain in place marked `Withdrawn`.
4. An assertion depending on an unresolved *Open Question* blocks acceptance.
5. Implementation issues cite the assertions they satisfy. A PR closing such an issue is reviewed against those assertions first, and against code quality second.
6. The complete gate is the *Acceptance Checklist* at the end of [templates/rfc-template.md](templates/rfc-template.md); every box must be checked.

---

## When an RFC is Required

An RFC is **mandatory** for:
* Amending or modifying [PROJECT_PRINCIPLES.md](../PROJECT_PRINCIPLES.md)
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

## RFC Frontmatter Schema

Every RFC file must begin with a machine-readable YAML frontmatter block:

```yaml
---
rfc: "0001"
title: "Project Vision & Strategic Architecture"
status: Draft        # Draft | Review | Accepted | Rejected | Superseded | Deprecated
milestone: M0        # M0 | M1 | M2 | M3 | M4
authors:
  - Author Name <email@example.com>
created: 2026-08-04
updated: 2026-08-04
discussion: https://github.com/lith-project/lith/discussions
requires: []         # List of prerequisite RFC numbers, e.g. ["0001", "0002"]
capability:          # Domain capabilities affected, e.g. ["Core Engine", "Graph"]
  - Core Engine
  - Architecture
supersedes: []
superseded_by: []
---
```

### Status Values
Architectural status is separate from code implementation state:
* **`Draft`**: Initial proposal being drafted.
* **`Review`**: Proposal undergoing formal maintainer and community review.
* **`Accepted`**: Approved architectural design ready for implementation issues.
* **`Rejected`**: Proposal declined after review.
* **`Superseded`**: Replaced by a newer RFC (specified in `superseded_by`).
* **`Deprecated`**: Design is obsolete and no longer recommended.

> [!IMPORTANT]
> **Never delete RFCs.** Even rejected, deprecated, or superseded RFCs remain in the repository as a historical record of architectural decisions and trade-offs.
