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
capability: []
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

## Open Questions
List unresolved questions or design trade-offs requiring community feedback:
- [ ] Question 1

## Future Work
What technical capabilities or RFCs naturally follow from this design?

## References

### Internal
- [PROJECT_PRINCIPLES.md](../../PROJECT_PRINCIPLES.md)
- [docs/glossary.md](../../docs/glossary.md)

### External
- [Specification Link](https://example.com)
