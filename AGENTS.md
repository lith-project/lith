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
2. Ensure the proposal includes machine-readable frontmatter (`status: Draft`, `milestone`, `requires`, `subsystem`).
3. Update [rfcs/index.md](rfcs/index.md) and [ARCHITECTURE.md](ARCHITECTURE.md) to reference the new RFC.

### Code Editing & Refactoring

1. Verify signature changes across all calling sites.
2. Maintain clear, wrapped error messages.
3. Run tests and static analysis (`go test ./...`) after edits.

---

## 3. Task Selection & Claim Protocol

Work is tracked in GitHub's native issue metadata, not in prose. An agent should
never have to parse [ROADMAP.md](ROADMAP.md) to decide what to do next.

```text
GitHub Milestone   M1-A · Lifecycle
  Epic tracker     #14                  has sub-issues — never claim it
    Epic trackers  #15 … #21, #23 … #25 have sub-issues — never claim them
      Task leaves  #27 … #63             claimable, ordered by blocked_by
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

3. Branch from `main` per [CONTRIBUTING.md](CONTRIBUTING.md#3-submitting-pull-requests) —
   `feature/<N>-<slug>`, e.g. `feature/15-config-loading`.
4. Read the issue body in full. **Scope, Non-Goals, Definition of Done, Dependencies,
   and Principles Check are the contract**, and every RFC the issue links is required
   reading before the first line of code.
5. Implement against the Definition of Done. The Non-Goals list is binding: an issue
   that says "no hot-reload" means the PR does not add hot-reload, however small the diff.
6. Open a PR **against `main`** referencing `Closes #<N>`.

### Rules

- **One claim at a time.** Release a claim you are not actively working
  (`gh issue edit <N> --remove-label agent:wip --remove-assignee @me`).
- **`main` is the only long-lived branch.** Cut `feature/<N>-<slug>` from it and target
  it. A pull request merged into any other branch does not close its issue — GitHub
  honours `Closes #<N>` only on the default branch — and it is not seen by the
  workflows, which are keyed to `main`. Both failures are silent.
- **A merged pull request is not a finished task.** Confirm the issue is closed and
  `agent:wip` is gone before claiming the next one. A claimed issue whose work has
  already landed blocks every task behind it, and `tools/next-task.sh` reports the
  queue as empty rather than as stuck.
- **Never work an issue with an open blocker.** If the dependency looks wrong, argue it
  on the issue and let it be re-wired. Do not route around the graph.
- **Never invent scope.** Work not covered by an open issue needs an issue first, and
  architectural work needs an RFC (§2).

### Working in parallel

More than one agent can work at once. What limits it is the `blocked_by` graph, not a
rule — `tools/next-task.sh --all` lists every actionable task, and two agents holding
two claims from that list is the intended shape.

Two constraints decide whether tasks are genuinely concurrent, and both are visible in
the issue body:

- **Files.** Tasks that name disjoint paths in their Files table can run together.
  Tasks that both modify `cmd/lithd/main.go` — the wiring tasks at the end of each
  epic — cannot, and are chained for that reason.
- **Contract.** A task that consumes an exported signature waits for the task that
  defines it. A task that merely runs alongside it does not.

Encode both in `blocked_by` when adding work. A chain that serialises tasks touching
different packages is a bug in the plan: it idles agents and hides which dependencies
are real.

### Adding work

Create tasks as sub-issues of their epic, set `blocked_by` to encode build order, and
assign the milestone. Issue metadata is the plan of record; ROADMAP.md carries the
narrative and the milestone Definitions of Done.

---

## 4. Communication & Summaries

- Provide concise, structured pull request descriptions referencing corresponding issues or RFCs.
- Always use repository-relative links (`VISION.md`, `PROJECT_PRINCIPLES.md`, `rfcs/index.md`) when referring to project documents.


<claude-mem-context>
# Memory Context

# [lith] recent context, 2026-08-06 4:01am GMT+2

Legend: 🎯session 🔴bugfix 🟣feature 🔄refactor ✅change 🔵discovery ⚖️decision
Format: ID TIME TYPE TITLE
Fetch details: get_observations([IDs]) | Search: mem-search skill

Stats: 50 obs (17,603t read) | 0t work

### Aug 4, 2026
S36 RFC conformance gate review response committed and pushed (Aug 4 at 6:37 PM)
969 6:56p 🔵 RFC-0002 subsystem field rename completed with acceptance rationale
970 6:57p 🔵 Memory agent observation session persistence established
971 " 🔵 Memory agent observation session continuation after interruption
972 " 🔵 Memory agent observation cycle continues with minimal context
973 6:58p ✅ RFC-0003 audited and RFC-0005 updated with subsystem rename and cancellation bound
974 " 🔵 Memory agent observation session continued with minimal context
975 " 🔵 Memory agent observation session reached eleventh interruption cycle
976 6:59p ✅ RFC-0004 updated with subsystem rename and debounce bounds specification
977 " ✅ RFC-0004 acceptance checklist completed and index column headers renamed to Subsystems
978 7:00p 🔵 Memory agent observation session resumed with no new technical work
979 " 🔵 Memory agent observation session continued after thirteenth interruption
980 7:09p ✅ RFC-0002/0003/0004/0005 reviews completed with cross-stack fixes identified
981 " 🔵 RFC review timing and commit references verified
982 7:10p ✅ All four lith RFC reviews completed and approved
983 " ✅ All four lith RFC reviews completed and approved
984 7:14p ⚖️ Schema refinements to RESOLUTION, DIAGNOSTIC, and META tables in RFC-0003
985 7:16p ✅ Revised reverse index staleness risk explanation in RFC-0004
986 7:17p ✅ Resolved RFC-0004 contradictions and clarified test scope
987 7:27p ✅ RFC-0005 Job Engine specification approved with cross-stack terminology fixes
988 7:32p ✅ RFC acceptance branch created for architecture stack
989 7:33p ✅ M0 architecture RFC stack accepted
990 7:34p ✅ ROADMAP.md updated to reflect RFC acceptance completion
991 7:35p ✅ M0 architecture RFC acceptance committed and pushed
992 7:37p 🟣 M0 architecture RFCs accepted and merged
993 " 🔵 Test vault specification defines deterministic corpus infrastructure
994 7:40p ✅ Corpus .gitattributes prevents line ending normalization
995 7:45p ✅ Regression directory placeholder created
996 7:47p ✅ Generator and benchmark placeholders established
997 7:53p 🔵 APFS Unicode normalization collision requires separate directories
998 7:55p ⚖️ Lith project shifts from repository building to architecture-first development with RFC process
999 7:57p ⚖️ Lith project adopts architecture-first development model with three-track RFC process
1000 " 🔵 Unicode normalization test corpus generation requires programmatic NFC/NFD computation to avoid authoring pipeline silent normalization
1001 7:58p ⚖️ Lith project adopts architecture-first development with RFC-driven workflow before any implementation
1002 8:06p 🔵 Git core.precomposeunicode silently rewrites NFD paths to NFC at commit time on macOS/APFS, making committed NFD test files impossible
1003 11:26p 🟣 GitHub issue dependency automation tool added
1004 11:32p ✅ Epic issue type created at organization level
1005 " ✅ GitHub issue types migrated from Feature to Epic
1006 11:36p 🔵 M1-A milestone epic structure reviewed
1007 11:53p ⚖️ Evaluated project management strategies for multi-agent task coordination
1008 11:54p ⚖️ Evaluated GitHub Projects for multi-agent task coordination
1009 11:55p 🟣 Config system tasks written for lith Go project
1010 11:56p 🟣 Logging system tasks written for lith Go project
1011 11:57p 🟣 Path identity system tasks written for lith vault management
1012 11:58p 🟣 File watcher and debouncer system tasks written for lith indexing
1013 11:59p 🟣 Event queue and daemon lifecycle tasks written for lith
### Aug 5, 2026
1014 2:07a 🔵 GitHub issue graph investigation for RFC-0006 decomposition
1015 2:10a 🔵 GitHub Actions workflows configuration reviewed for lith project
1016 3:16a 🔵 GitHub workflow prevents claiming blocked issues
1017 3:23a 🔵 GitHub workflow correctly prevents blocked task assignment
1018 7:36a ✅ Scoped workflow files staged in isolated worktree
</claude-mem-context>