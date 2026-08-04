---
rfc: "000X"
title: "RFC Title"
status: Draft
milestone: M0
authors:
  - Author Name <email@example.com>
created: YYYY-MM-DD
updated: YYYY-MM-DD
discussion: https://github.com/lith-project/lith/discussions
requires: []
subsystem: []
supersedes: []
superseded_by: []
---

# RFC-000X: RFC Title

## Summary
A short paragraph summarizing the proposal. If someone only reads this section, they should understand the core proposal and its purpose.

## Motivation
Why does this feature or change need to exist? What specific problem or limitation does it solve?

## Goals
List concrete, verifiable goals of this proposal:
- Goal 1
- Goal 2

## Non-Goals
Explicitly define what this RFC does NOT attempt to do or solve:
- Non-goal 1
- Non-goal 2

## Background
*(Optional)* Background context, prior architectural discussions, or references to related RFCs.

## Proposed Design
Detailed technical design. Cover component responsibilities, data flows, interfaces, domain entities, transaction semantics, and diagrams where helpful.

### Component Topology & Data Flow
```
[ Component A ] ---> [ Component B ] ---> [ Disposable State Store ]
```

### Domain Interfaces & Capabilities
Detailed specification of capabilities, data structures, and interactions.

## Alternatives Considered
List at least two alternative approaches considered and why they were rejected:
1. **Alternative 1**: Description and rationale for rejection.
2. **Alternative 2**: Description and rationale for rejection.

## Risks
Identify technical, operational, performance, security, or maintenance risks and mitigation strategies:
- **Risk 1**: Mitigation...

## Migration
Describe how existing systems, configurations, databases, or client applications transition to this design. If none, state "None".

## Conformance

Numbered, independently verifiable assertions defining what it means for an implementation to conform to this RFC. Reviewers check implementations against these assertions, not against prose.

**Rules:**
* Every assertion has a stable identifier `C-N`, referenced project-wide as `RFC-000X/C-N`.
* Assertions use RFC 2119 keywords (`MUST`, `MUST NOT`, `SHOULD`, `MAY`).
* Every assertion states an **observable** outcome. If it cannot be observed by a test, a static check, a benchmark threshold, or a documented manual procedure, it is not an assertion — move it to *Proposed Design*.
* Identifiers are never renumbered or reused.
* Every assertion carries an **assertion state**: `Active` or `Withdrawn`. This two-value vocabulary belongs to assertions alone and is deliberately disjoint from the RFC `status` enum — an RFC has a status, a clause within it has an assertion state, and the two are never mixed. A retired assertion stays in place marked `Withdrawn` with the RFC number that retired it, so its identifier is never reused.
* An RFC **cannot** move to `Accepted` while any assertion lacks a Verification method.

### C-1: Short assertion title
**Assertion:** Implementations MUST ...
**Verification:** How conformance is demonstrated (unit test, integration test, property test, static/CI check, benchmark with a stated threshold, or documented manual procedure) and the expected observable outcome.
**Milestone:** M1-A

### C-2: Short assertion title
**Assertion:** Implementations MUST NOT ...
**Verification:** ...
**Milestone:** M1-B

## Open Questions
List unresolved questions or design trade-offs requiring community feedback:
- [ ] Question 1

## Future Work
What technical capabilities or RFCs naturally follow from this design?

## Acceptance Checklist

An RFC is `Accepted` only when every box is checked:

- [ ] Every `Conformance` assertion has a Verification method and an owning milestone
- [ ] No assertion depends on unresolved *Open Questions*
- [ ] *Non-Goals* are explicit
- [ ] At least one diagram covering the primary data flow, component topology, or state lifecycle of the proposal. A diagram must carry information the prose does not; a box-and-arrow restatement of a paragraph does not satisfy this item
- [ ] Every diagram validated as Mermaid by a parser, not by eye
- [ ] All domain terms used normatively exist in [docs/glossary.md](../../docs/glossary.md); new terms added there in the same PR
- [ ] No conflict with [PROJECT_PRINCIPLES.md](../../PROJECT_PRINCIPLES.md); any proposed amendment is stated verbatim in this RFC
- [ ] **If** this RFC names any capability, that capability exists in [docs/reference/capability-catalog.md](../../docs/reference/capability-catalog.md) with a `CAP-NNNN` identifier. Mark **N/A** when the RFC names none — the catalog is instantiated by the RFC that defines the capability model, and this item is unsatisfiable, not failed, before then
- [ ] [rfcs/index.md](../index.md) and [ARCHITECTURE.md](../../ARCHITECTURE.md) rows updated
- [ ] Reviewed and approved by maintainers

## References

### Internal
- [PROJECT_PRINCIPLES.md](../../PROJECT_PRINCIPLES.md)
- [docs/glossary.md](../../docs/glossary.md)

### External
- [Specification Link](https://example.com)
