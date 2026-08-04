---
rfc: "0002"
title: "Domain Model & Vault AST Representation"
status: Draft
authors:
  - Lith Maintainers <maintainers@lith.dev>
created: 2026-08-04
updated: 2026-08-04
discussion: https://github.com/lith-project/lith/discussions
supersedes:
superseded_by:
---

# RFC-0002: Domain Model & Vault AST Representation

## Summary
Defines the core domain entities, abstract syntax tree (AST) structures, frontmatter schemas, and note representations used by Lith to model Markdown vaults.

## Motivation
Lith requires a consistent, memory-efficient domain model to represent Markdown documents, wiki-links, block embeds, frontmatter metadata, and entity relations across all components.

## Goals
- Define canonical Go domain structures for notes, blocks, links, and tags.
- Support lossless parsing of Obsidian-flavored Markdown constructs.

## Non-Goals
- Defining database storage implementations (covered in RFC-0003).
- Defining query engine implementations (covered in RFC-0004).

## Background
See [RFC-0001](0001-project-vision.md).

## Proposed Design
*To be elaborated during M0 architecture phase.*

## Alternatives Considered
1. Raw string manipulation.
2. Generic unstructured JSON AST.

## Risks
TBD.

## Migration
None.

## Open Questions
- [ ] Handling custom frontmatter types and Obsidian YAML extensions.

## Future Work
- Implementation in core engine packages.

## References
- [RFC-0001](0001-project-vision.md)
