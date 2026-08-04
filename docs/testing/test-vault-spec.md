# Test Vault Specification

**Status:** Draft · **Owner:** [RFC-0001](../../rfcs/0001-project-vision.md) · **Milestone:** M0 · **Updated:** 2026-08-04

This document specifies the **test vault** — the fixed input against which every parser test, indexing test, regression test, and benchmark in Lith runs. It is a specification only: no corpus files and no generator code are created by it.

The test vault is expected to outlive most of the code that consumes it. It is therefore specified before implementation, not discovered during it.

---

## 1. Purpose

One corpus, used by everything:

* **Parser tests** — RFC-0002 domain model conformance.
* **Storage tests** — RFC-0003 rebuild determinism ([RFC-0001/C-2](../../rfcs/0001-project-vision.md#c-2-rebuild-determinism)).
* **Indexing tests** — RFC-0004 incremental correctness against a known-good full index.
* **Search tests** — deterministic ranking ([RFC-0001/C-7](../../rfcs/0001-project-vision.md#c-7-deterministic-core-search)).
* **Regression tests** — every fixed bug adds a corpus file and a golden, permanently.
* **Benchmarks** — reproducible timings across commits and machines.

If a test needs a vault, it uses this one. Ad-hoc inline fixtures are permitted only for pure unit tests with no filesystem involvement.

---

## 2. Design Principles

1. **Deterministic.** Same corpus + same code → same output, byte for byte. No wall-clock timestamps, no random ordering, no locale dependence, no network.
2. **Split by nature.** Edge cases are **committed** (small, curated, reviewable). Scale is **generated** (seeded, reproducible, never committed).
3. **Golden-backed.** A corpus without expected results is data, not a test asset. Every committed note has a golden parse result.
4. **License-clean.** All content is written for this project. No copied third-party notes, no scraped content, no real user vaults, no PII.
5. **Frozen by default.** Changing a corpus file changes goldens for every consumer. Corpus edits are reviewed as behaviour changes, not as fixture tweaks.

---

## 3. Layout

```
tests/vault/
├── corpus/                  # committed, deterministic, small
│   ├── content/             # semantic/content edge cases      EC-CNT-*
│   ├── frontmatter/         # frontmatter variants             EC-FM-*
│   ├── links/               # link and embed variants          EC-LNK-*
│   ├── encoding/            # bytes, line endings, BOM         EC-ENC-*
│   ├── filesystem/          # names, paths, non-md files       EC-FS-*
│   ├── assets/              # images, PDFs, canvas, bases
│   ├── regression/          # one file per fixed bug           EC-REG-*
│   └── .gitattributes       # mandatory — see §6
├── expected/                # goldens for corpus/
│   ├── manifest.json
│   └── notes/…
├── generators/              # specs here; code lands in M1
│   ├── scale/               # S/M/L vaults for benchmarks
│   ├── malformed/           # fuzz/hostile input
│   ├── filesystem/          # cases that cannot be committed   EC-FS-X-*
│   └── dynamic/             # runtime/timing scenarios         EC-DYN-*
└── benchmarks/              # fixed scenarios + thresholds
```

`corpus/` is a **valid vault**: it opens in Obsidian without error, including a minimal `.obsidian/` directory, so a human can inspect what the tests are asserting about.

---

## 4. Corpus Rules

* **Size ceiling:** ~200–300 files. Beyond that, reviewers stop reading diffs, which defeats the point.
* **Stable ordering:** filenames sort deterministically; no test depends on directory iteration order.
* **One concern per file.** A file exercising CRLF does not also exercise malformed YAML. Debugging a compound failure costs more than the extra file.
* **Every file is registered** in `expected/manifest.json` with its edge-case IDs. An unregistered corpus file fails the corpus-integrity check.
* **Regression files are permanent.** Never deleted, never "cleaned up". Naming: the **file** is named for the issue it came from (`issue-142-crlf-in-callout.md`); the **edge-case ID** it registers under in the manifest is `EC-REG-142`, matching the issue number. The filename carries human context, the ID carries machine identity — where they appear to conflict, the manifest ID is authoritative.

---

## 5. Edge-Case Matrix

Every case has a stable ID, referenced by tests and by RFC conformance assertions. IDs are never reused.

### 5.1 Content & Structure (`EC-CNT-*`) — committed

| ID | Case | Stresses |
| -- | ---- | -------- |
| EC-CNT-001 | Large note (~1 MB, deep heading tree) | Parser memory, section indexing |
| EC-CNT-002 | Tiny note (one word, no heading) | Degenerate AST |
| EC-CNT-003 | Empty note (whitespace only) | Empty-document handling |
| EC-CNT-004 | Daily notes, ISO-dated filenames | Date inference, sequence detection |
| EC-CNT-005 | Massive MOC (500+ outbound links) | Graph fan-out |
| EC-CNT-006 | Duplicate concepts (near-identical notes) | Entity resolution, dedup |
| EC-CNT-007 | Deeply nested lists (10 levels) | Block addressing |
| EC-CNT-008 | Tasks: open, done, cancelled, scheduled, recurring | Task capability |
| EC-CNT-009 | Dataview inline fields and query blocks | Non-standard syntax survives parsing |
| EC-CNT-010 | Mermaid, LaTeX, nested/indented code fences | Fenced-content isolation |
| EC-CNT-011 | Callouts, nested callouts, admonitions | Block typing |
| EC-CNT-012 | Templates with unresolved placeholders | Template-vs-note distinction |
| EC-CNT-013 | Tables including malformed alignment rows | Table parsing |
| EC-CNT-014 | Footnotes, definition lists | Reference resolution |
| EC-CNT-015 | HTML embedded in Markdown | Passthrough handling |
| EC-CNT-016 | Note whose entire body is one code fence | False-positive link/tag detection |

### 5.2 Frontmatter (`EC-FM-*`) — committed

| ID | Case |
| -- | ---- |
| EC-FM-001 | No frontmatter |
| EC-FM-002 | Empty frontmatter (`---` / `---`) |
| EC-FM-003 | Malformed YAML (unclosed quote, bad indent) |
| EC-FM-004 | Duplicate keys |
| EC-FM-005 | `tags` as string vs list vs comma-separated string |
| EC-FM-006 | `aliases` variants including a bare string |
| EC-FM-007 | Nested objects and arrays of objects |
| EC-FM-008 | Dates: ISO, quoted, ambiguous, invalid |
| EC-FM-009 | Unicode keys and values |
| EC-FM-010 | Tabs used for indentation (invalid YAML) |
| EC-FM-011 | `---` appearing inside a code block (false delimiter) |
| EC-FM-012 | Frontmatter not at byte 0 (preceded by blank line or text) |
| EC-FM-013 | Very large frontmatter (~100 keys) |
| EC-FM-014 | Null / empty-value keys |

### 5.3 Links & Embeds (`EC-LNK-*`) — committed

| ID | Case |
| -- | ---- |
| EC-LNK-001 | `[[Note]]` |
| EC-LNK-002 | `[[Note\|alias]]` |
| EC-LNK-003 | `[[Note#Heading]]` |
| EC-LNK-004 | `[[Note#^block-id]]` |
| EC-LNK-005 | `![[Embed]]`, image embed, PDF embed |
| EC-LNK-006 | Markdown links: relative, absolute, URL-encoded |
| EC-LNK-007 | External links (`https://`, `obsidian://`, `mailto:`) |
| EC-LNK-008 | Link inside fenced code block — MUST NOT resolve |
| EC-LNK-009 | Link inside inline code — MUST NOT resolve |
| EC-LNK-010 | Escaped brackets — MUST NOT resolve |
| EC-LNK-011 | Broken link (target absent) |
| EC-LNK-012 | Circular links (A→B→A) and self-link |
| EC-LNK-013 | Ambiguous link: same basename in two directories |
| EC-LNK-014 | Link differing only by case from its target |
| EC-LNK-015 | Link inside a frontmatter value |
| EC-LNK-016 | Link with unicode / emoji in the target name |
| EC-LNK-017 | Nested embed (A embeds B which embeds A) |

### 5.4 Encoding & Bytes (`EC-ENC-*`) — committed

Line endings are the most fragile axis here; §6 exists solely so these survive git.

| ID | Case |
| -- | ---- |
| EC-ENC-001 | LF only |
| EC-ENC-002 | CRLF only |
| EC-ENC-003 | Mixed LF and CRLF in one file |
| EC-ENC-004 | Lone CR (classic Mac) |
| EC-ENC-005 | UTF-8 BOM |
| EC-ENC-006 | UTF-16LE with BOM |
| EC-ENC-007 | Invalid UTF-8 byte sequence |
| EC-ENC-008 | No trailing newline |
| EC-ENC-009 | Multiple trailing newlines |
| EC-ENC-010 | Zero-byte `.md` file |
| EC-ENC-011 | NUL byte inside a `.md` file |
| EC-ENC-012 | Combining characters, RTL text, zero-width joiners |
| EC-ENC-013 | Non-breaking spaces and exotic whitespace |

### 5.5 Filesystem (`EC-FS-*`) — committed where possible

| ID | Case |
| -- | ---- |
| EC-FS-001 | Spaces in filename |
| EC-FS-002 | `#`, `[`, `]`, `%`, `&` in filename |
| EC-FS-003 | Leading/trailing dots and spaces |
| EC-FS-004 | Unicode filename, NFC-normalized |
| EC-FS-005 | Unicode filename, NFD-normalized (macOS normalizes these — same logical name, different bytes) |
| EC-FS-006 | Emoji filename |
| EC-FS-007 | Same basename in multiple directories |
| EC-FS-008 | Deeply nested path (15 levels) |
| EC-FS-009 | Non-Markdown files: `.txt`, `.png`, `.pdf`, `.canvas`, `.base` |
| EC-FS-010 | Hidden files, `.obsidian/`, `.trash/` |
| EC-FS-011 | Uppercase extension `.MD`, and `.markdown` |
| EC-FS-012 | File named like a directory; directory named `*.md` |
| EC-FS-013 | Symlink to a file inside the vault |
| EC-FS-014 | Symlink escaping the vault root (must not be followed out) |
| EC-FS-015 | Broken symlink |

### 5.6 Generator-Only Cases (`EC-FS-X-*`, `EC-DYN-*`) — never committed

These cannot exist in a git checkout, or cannot exist identically across platforms. The generator materializes them at test time in a temporary directory.

| ID | Case | Why it cannot be committed |
| -- | ---- | -------------------------- |
| EC-FS-X-001 | Case collision: `Note.md` + `note.md` | Cannot coexist on default macOS APFS or on Windows |
| EC-FS-X-002 | Path exceeding 255-byte component / 4096-byte total | Platform-dependent; breaks checkout |
| EC-FS-X-003 | Read-only file; directory without execute permission | Permissions not portably preserved by git |
| EC-FS-X-004 | Hard link between two vault paths | Not representable in git |
| EC-FS-X-005 | Single file >100 MB; 100k-line file | Repository weight |
| EC-DYN-001 | File modified **during** a scan | Timing, not content |
| EC-DYN-002 | Identical `mtime` + identical size, different content | Forces content hashing over stat-only change detection |
| EC-DYN-003 | `mtime` in the future / clock skew | Timing |
| EC-DYN-004 | Rapid create→rename→delete burst | Watcher debouncing |
| EC-DYN-005 | Atomic-save pattern (write temp → rename over target) | Editor behaviour, not a static file |
| EC-DYN-006 | Whole directory renamed | Watcher event coalescing |
| EC-DYN-007 | Vault on a case-insensitive vs case-sensitive volume | Volume property |

> `EC-DYN-002` deserves emphasis: change detection based on `mtime` and size alone passes every static test and silently loses updates in the field. It is a required scenario before RFC-0004 can be considered conformant.

---

## 6. `.gitattributes` (mandatory)

Without this, git normalizes line endings on checkout and every `EC-ENC-*` case silently becomes `EC-ENC-001`. The corpus then tests nothing.

`tests/vault/corpus/.gitattributes` MUST:

1. Disable all text conversion for the corpus (`* -text`), so bytes on disk equal bytes in the object store on every platform.
2. Never rely on the contributor's `core.autocrlf` setting.
3. Mark binary assets explicitly (`*.png binary`, `*.pdf binary`).
4. Exclude the corpus from export/archive filters that would rewrite content.

A CI check MUST verify, after a fresh clone on Linux **and** on Windows, that the SHA-256 of every corpus file matches `expected/manifest.json`. That check is the actual guarantee; `.gitattributes` is only the mechanism.

---

## 7. Goldens

### 7.1 Corpus manifest — `tests/vault/expected/manifest.json`

```json
{
  "spec_version": 1,
  "corpus_id": "lith-corpus-v1",
  "generated_at": "2026-08-04T00:00:00Z",
  "corpus_digest": "sha256:0000000000000000000000000000000000000000000000000000000000000000",
  "file_count": 214,
  "files": [
    {
      "path": "encoding/crlf-only.md",
      "size_bytes": 482,
      "digest": "sha256:1111111111111111111111111111111111111111111111111111111111111111",
      "edge_cases": ["EC-ENC-002"]
    }
  ]
}
```

* Digests in the example above are **illustrative placeholders**. Real values are the SHA-256 of the file's bytes as they exist on disk, hex-encoded, prefixed `sha256:`.
* `generated_at` is a **fixed synthetic** RFC 3339 UTC value, not the actual generation time. Wall-clock in a golden makes every regeneration a diff.
* `corpus_digest` is a stable hash over the sorted `(path, digest)` pairs.
* Every corpus file appears exactly once, with at least one edge-case ID.

### 7.2 Per-note goldens — `tests/vault/expected/notes/<path>.json`

Canonical parse result for one note:

```json
{
  "spec_version": 1,
  "path": "links/alias-link.md",
  "frontmatter": {},
  "title": "Alias Link",
  "sections": [],
  "blocks": [],
  "links_out": [],
  "tags": [],
  "tasks": [],
  "embeds": [],
  "warnings": []
}
```

Canonicalization rules — all mandatory, all in service of determinism:

* Object keys sorted; arrays in document order.
* Byte offsets, not line/column — line/column is line-ending dependent, so `EC-ENC-*` cases would fail spuriously.
* No absolute paths, no timestamps, no hostname, no version string.
* `warnings` carries stable diagnostic codes, not human-readable text.

Field *semantics* are owned by **RFC-0002**. This document owns the file format and the determinism rules only.

### 7.3 Regeneration workflow

1. Goldens are regenerated only via an explicit `--update-goldens` flag. Never automatically, never on test failure.
2. Regeneration output is reviewed **as a diff**. An unexplained golden change is a behaviour change.
3. A PR touching goldens states which edge-case IDs changed, and why.
4. CI runs goldens read-only; a job that can rewrite goldens can never fail.

---

## 8. Generator

Specified here; implemented in M1. The generator produces vaults that are too large, too hostile, or too platform-specific to commit.

**Requirements:**

* **Seeded.** Explicit integer seed in config; identical seed → byte-identical vault. Never seeded from the clock.
* **Versioned.** Emits `generator_version` + `spec_version`; benchmark results are comparable only within the same pair.
* **Isolated.** Writes only to a caller-provided temporary directory. Writing into `corpus/` is a hard error.
* **Deterministic timestamps.** All `mtime` values derive from the seed and a fixed epoch, except where the scenario is specifically about time (`EC-DYN-003`).
* **Self-describing.** Emits a manifest in the same format as §7.1, so generated vaults are inspectable with the same tooling.
* **Skippable.** A scenario unsupported on the host platform (e.g. `EC-FS-X-001` on case-insensitive APFS) reports `skipped` with a reason — never a silent pass.

**Configuration knobs:** note count · directory depth and breadth · link density (mean outbound links per note) · orphan ratio · broken-link ratio · MOC count and size · tag cardinality and reuse · frontmatter variant mix · note size distribution · asset count and size · daily-note date range.

**Scenario families:** `scale/` (volume) · `malformed/` (fuzz and hostile input) · `filesystem/` (`EC-FS-X-*`) · `dynamic/` (`EC-DYN-*`, timing and mutation).

---

## 9. Benchmark Tiers

Fixed seeds, fixed configs — so a number measured today is comparable to one measured next year.

| Tier | Notes | Seed | Purpose |
| ---- | ----- | ---- | ------- |
| **S** | 1,000 | `lith-s-1` | Runs in CI on every PR; regression gate |
| **M** | 10,000 | `lith-m-1` | Nightly; realistic power-user vault |
| **L** | 100,000 | `lith-l-1` | Release gate; scale ceiling |

Each tier ships a **named configuration profile** committed alongside the generator, so a result is reproducible from the tier name alone without knowing the full knob set. A profile pins every knob in §8 plus the seed; changing any value in a profile is a new profile name, never an edit to an existing one — otherwise yesterday's numbers silently stop being comparable to today's.

Every tier measures: cold full index · warm incremental index of a single-file change · full rebuild after deleting all derived state · core search latency (p50/p95/p99) · peak RSS.

Thresholds are set once a first implementation exists — a threshold invented before any measurement is noise. What is fixed **now** is that the tiers, seeds, and measured operations never change silently.

---

## 10. Conformance Hooks

| Assertion | What the corpus provides |
| --------- | ------------------------ |
| [RFC-0001/C-2](../../rfcs/0001-project-vision.md#c-2-rebuild-determinism) | Fixed input for index → snapshot → delete → re-index → snapshot equality |
| [RFC-0001/C-7](../../rfcs/0001-project-vision.md#c-7-deterministic-core-search) | Golden query results over a frozen corpus |
| RFC-0002 *(future)* | Per-note goldens for every `EC-CNT-*`, `EC-FM-*`, `EC-LNK-*` |
| RFC-0003 *(future)* | Rebuild and corruption-recovery scenarios |
| RFC-0004 *(future)* | `EC-DYN-*` — incremental result must equal full-index result |

**Corpus integrity check** (CI, every build): every corpus file is registered in the manifest; every registered file exists; every digest matches; every edge-case ID referenced by a test exists in this document.

---

## 11. Non-Goals

* Not a demo vault, not documentation, not a showcase.
* No real user data, no PII, no third-party or copyrighted content.
* Not a home for large binary assets — asset cases use the smallest valid file that exercises the code path.
* Does not define parse semantics. Golden *contents* are owned by RFC-0002; this document owns format and determinism.

## 12. Open Questions

- [ ] One golden file per note, or one aggregated file per corpus directory? (Per-note gives readable diffs; aggregated gives fewer files. Leaning per-note.)
- [ ] Is `.obsidian/` committed inside the corpus, or synthesized at test setup?
- [ ] Does the corpus carry a second vault root, to pin cross-vault link behaviour before multi-vault work begins?

## 13. References

- [RFC-0001: Project Vision & Strategic Architecture](../../rfcs/0001-project-vision.md)
- [PROJECT_PRINCIPLES.md](../../PROJECT_PRINCIPLES.md)
- [docs/glossary.md](../glossary.md)
- [ROADMAP.md](../../ROADMAP.md)
