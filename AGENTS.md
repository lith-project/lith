# AGENTS.md - Agent Operational Guidelines

This document outlines mandatory guidelines for AI agents operating within the **Lith** codebase.

---

## 1. Operating Rules

1. **Check Core Documents First**:
   - Read [VISION.md](VISION.md) for executive scope and non-goals.
   - Read [PROJECT_PRINCIPLES.md](PROJECT_PRINCIPLES.md) before proposing architecture or modifying core code.
   - Consult [docs/glossary.md](docs/glossary.md) for canonical domain terminology.
   - Read [ARCHITECTURE.md](ARCHITECTURE.md) and [rfcs/index.md](rfcs/index.md).
2. **Never Violate Constitutional Tenets**:
   - Do not make assumptions that SQLite or vector indexes are permanent sources of truth.
   - Do not bypass transactional mechanisms.
   - Do not hardcode vector embeddings as a mandatory dependency of the core engine.
   - Do not modify `PROJECT_PRINCIPLES.md` without an approved RFC.
3. **Respect Architectural Boundaries**:
   - Do not introduce arbitrary Go directory structures (`pkg/...`, `internal/...`) unless agreed upon in an active RFC.
   - Maintain strict separation between core domain logic and external adapters (CLI, REST, MCP, SDK).

---

## 2. Workflows for Agents

### Proposing Architectural Changes
1. Draft an RFC file in `rfcs/NNNN-title.md` using [rfcs/templates/rfc-template.md](rfcs/templates/rfc-template.md).
2. Ensure the proposal includes machine-readable frontmatter (`status: Draft`, `milestone`, `requires`, `capability`).
3. Update [rfcs/index.md](rfcs/index.md) and [ARCHITECTURE.md](ARCHITECTURE.md) to reference the new RFC.

### Code Editing & Refactoring
1. Verify signature changes across all calling sites.
2. Maintain clear, wrapped error messages.
3. Run tests and static analysis (`go test ./...`) after edits.

---

## 3. Task Selection & Claim Protocol

Work is tracked in GitHub's native issue metadata, not in prose. An agent should
never have to parse [ROADMAP.md](ROADMAP.md) to decide what to do next.

```
GitHub Milestone   M1-A · Lifecycle
  Epic tracker     #14           has sub-issues — a tracker, never claim it
    Epic / task    #15 … #21     leaf work, ordered by blocked_by
```

| Field | Meaning |
| --- | --- |
| Milestone | Which roadmap phase the work belongs to (`M1-A`, `M1-B`, …) |
| Sub-issues | Parent/child. **An issue with sub-issues is a tracker — do not claim it.** |
| `blocked_by` | Build order, and the authoritative one. ROADMAP.md narrates the same graph; it does not override it. |
| Issue type | `Epic` for trackers, `Task` for claimable leaf work |
| `agent:wip` | Claimed by an agent, work in progress |
| `agent:blocked` | Blocked on something not expressible as a dependency — a decision, external input |

### Picking up work

1. Ask for the next actionable issue:

   ```bash
   tools/next-task.sh
   ```

   Pass a milestone prefix (`tools/next-task.sh M1-B`) to target a different phase,
   and `--all` to list every actionable issue rather than the first.

2. Claim it before starting — this is what stops two agents landing on one issue:

   ```bash
   gh issue edit <N> --add-label agent:wip --add-assignee @me
   ```

3. Branch per [CONTRIBUTING.md](CONTRIBUTING.md#3-submitting-pull-requests) —
   `feature/<N>-<slug>`, e.g. `feature/15-config-loading`.
4. Read the issue body in full. **Scope, Non-Goals, Definition of Done, Dependencies,
   and Principles Check are the contract**, and every RFC the issue links is required
   reading before the first line of code.
5. Implement against the Definition of Done. The Non-Goals list is binding: an issue
   that says "no hot-reload" means the PR does not add hot-reload, however small the diff.
6. Open a PR referencing `Closes #<N>`.

### Rules

- **One claim at a time.** Release a claim you are not actively working
  (`gh issue edit <N> --remove-label agent:wip --remove-assignee @me`).
- **Never work an issue with an open blocker.** If the dependency looks wrong, argue it
  on the issue and let it be re-wired. Do not route around the graph.
- **Never invent scope.** Work not covered by an open issue needs an issue first, and
  architectural work needs an RFC (§2).

### Adding work

Create tasks as sub-issues of their epic, set `blocked_by` to encode build order, and
assign the milestone. Issue metadata is the plan of record; ROADMAP.md carries the
narrative and the milestone Definitions of Done.

---

## 4. Communication & Summaries

- Provide concise, structured pull request descriptions referencing corresponding issues or RFCs.
- Always use repository-relative links (`VISION.md`, `PROJECT_PRINCIPLES.md`, `rfcs/index.md`) when referring to project documents.
