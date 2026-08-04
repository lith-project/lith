---
name: Agent Task
about: A single, self-contained unit of work an autonomous coding agent can complete end to end
title: '[TASK] '
labels: 'enhancement'
assignees: ''
---

<!--
  This template is a contract, not a suggestion box.

  A task is correctly written when an agent with no prior knowledge of Lith, no
  access to this conversation, and no ability to ask a question can complete it
  from this text alone. That standard is what every section below serves.

  Rules for the author:
    * One coherent outcome. If two parts can land independently, write two tasks.
    * Name exact file paths and exact interfaces, signatures, schema, or generated
      artifacts when they apply. Never leave the implementation surface implicit.
    * Distil constraints into literal rules. "The queue is not durable state" is
      a design principle an agent can satisfy in the wrong way; "the Queue type
      must not have a Save or Load method, and must not import os or database/sql"
      is a rule it can obey.
    * Keep required reading bounded. Restate every governing RFC constraint here
      as a rule and link the source assertion under Provenance.
    * Verification must be a command that can be pasted into a shell.

  Delete these comments before submitting.
-->

## Parent

Part of #<epic-issue-number>.

## Objective

<!-- One sentence. What exists after this task that did not exist before. -->

## Scope

<!-- Enumerate the exact behavior and artifacts this task owns. -->

-

## Files

<!-- Exact paths. Mark each created or modified. -->

| Path | Action |
| --- | --- |
| `<exact/path>` | create / modify |

## Interface / Contract

<!--
  State exact exported signatures, command behavior, schema, generated artifacts,
  or other externally observable contract. Write "None" when none applies.
  Do not leave this blank.
-->

-

## Rules

<!--
  Literal, checkable constraints. Each one should be something a reviewer can
  point at a line of code and say "this violates rule N".
-->

1.
2.

## Definition of Done

<!-- Observable outcomes. Every line must be true for the task to be complete. -->

- [ ]
- [ ]

## Verification

<!-- Commands, exactly as they should be run from the repository root. -->

```bash
<commands runnable from the repository root>
```

## Non-Goals

<!--
  What an agent might reasonably add and must not. This section prevents more
  wasted work than any other; be specific and be generous with it.
-->

-

## Dependencies

<!--
  List every blocked_by issue and why it must land first. Write "None" when the
  task has no dependency. GitHub blocked_by metadata must match this section.
-->

-

## Principles Check

<!--
  Name the PROJECT_PRINCIPLES.md tenets affected by this task and state how the
  contract preserves them. Write "No constitutional impact" only after checking.
-->

-

## Provenance

<!--
  Where the rules above come from, for a reviewer or a human reader. An agent
  executing this task does not need to open these.
-->

- RFC-000X/C-N — <what it constrains>
