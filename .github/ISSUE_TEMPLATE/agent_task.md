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
    * One package. If the work spans two, it is two tasks.
    * Name exact file paths and exact exported signatures. Never "add a config
      loader" — always `internal/core/config/config.go`, `func Load(path string)
      (*Config, error)`.
    * Distil constraints into literal rules. "The queue is not durable state" is
      a design principle an agent can satisfy in the wrong way; "the Queue type
      must not have a Save or Load method, and must not import os or database/sql"
      is a rule it can obey.
    * Required reading should be zero. If an RFC constraint governs this task,
      restate it here as a rule and link the RFC as provenance, not as homework.
    * Verification must be a command that can be pasted into a shell.

  Delete these comments before submitting.
-->

## Parent

Part of #<epic-issue-number>.

## Objective

<!-- One sentence. What exists after this task that did not exist before. -->

## Files

<!-- Exact paths. Mark each created or modified. -->

| Path | Action |
| --- | --- |
| `internal/core/<pkg>/<file>.go` | create |
| `internal/core/<pkg>/<file>_test.go` | create |

## Public API

<!--
  Exact signatures this task must expose. If the task adds no exported symbol,
  write "None — internal to the package." Do not leave this blank.
-->

```go
// package <pkg>
```

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
go build ./...
go test ./internal/core/<pkg>/...
golangci-lint run ./internal/core/<pkg>/...
```

## Out of Scope

<!--
  What an agent might reasonably add and must not. This section prevents more
  wasted work than any other; be specific and be generous with it.
-->

-

## Provenance

<!--
  Where the rules above come from, for a reviewer or a human reader. An agent
  executing this task does not need to open these.
-->

- RFC-000X/C-N — <what it constrains>
