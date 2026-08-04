# Test Vault

Specified by [docs/testing/test-vault-spec.md](../../docs/testing/test-vault-spec.md). This file records provenance and the handful of decisions that document did not settle, so they aren't silent.

## Provenance

`corpus/` was authored by a one-off generation script (not committed — the spec commits the corpus, not a generator for it; `generators/` is a different thing, specified in §8 and implemented in M1). Text content is synthetic filler written for this project; binary assets use well-known minimal valid byte sequences (a documented 67-byte 1×1 PNG, a hand-built minimal PDF skeleton verified openable by `file`).

`expected/manifest.json` was built by hashing the corpus as it exists on disk — no digest in it was typed by hand. Regenerating it means rerunning the same hash pass; the entries are a function of the files, not a separate record to keep in sync by hand.

## Decisions the spec left open

**Manifest scope.** `.gitattributes` and `README.md` files are *not* registered in `manifest.json`. They are corpus infrastructure — git configuration and documentation — not edge-case-bearing test content, and the manifest schema requires every entry to carry at least one `EC-*` id. Forcing a fake id onto a README would be worse than excluding it.

**Symlink digests.** A symlink's `digest` is the SHA-256 of its **target path string**, not of the resolved target's content — matching how git itself stores and hashes a symlink (mode `120000`, blob = target text). `size_bytes` is the byte length of that string. This wasn't specified because the spec's manifest example predates any filesystem case being committed.

**`.obsidian/`.** Not created. [§12 Open Questions](../../docs/testing/test-vault-spec.md#12-open-questions) leaves this genuinely undecided (committed vs. synthesized at test setup), and guessing would bake a manifest inconsistency into the corpus. Deferred to whoever resolves that question.

**`EC-FS-005` is not committed.** It was authored, then removed. `git`'s own macOS/APFS auto-detected `core.precomposeunicode=true` silently rewrites an NFD path to NFC at commit time — verified directly: a working-tree file confirmed NFD via `os.listdir` showed up as NFC in `git ls-files -z` after staging, no error, and the same happened via plumbing-level `git update-index --cacheinfo`, ruling out a `git add`-specific quirk. This corrupts the committed path *at commit time, on this authoring host*, regardless of which platform later clones the repo — an NFD file committed here can never test NFD, anywhere. `EC-FS-004` (NFC) is unaffected and remains committed at `corpus/filesystem/unicode-normalization/nfc/café.md`. `EC-FS-005` moves to the generator-only set as `EC-FS-X-006`, the same category `EC-FS-X-001` already occupies for the identical underlying reason. See [`corpus/filesystem/unicode-normalization/README.md`](corpus/filesystem/unicode-normalization/README.md) for the full sequence, including the separate same-directory collision finding that preceded this one.

## What's not here yet

* **Per-note goldens** (`expected/notes/*.json`) — not generated. Byte-accurate offsets, section/block boundaries, and diagnostic codes are governed by RFC-0002 parsing rules that have no reference implementation yet. Hand-computing ~97 files' worth of structured parse output with no parser to check it against risks shipping goldens that are confidently wrong rather than absent — a worse failure mode, since a wrong golden looks like coverage. These wait for either a minimal reference parser or the real M1-B implementation.
* **`generators/*`** and **`benchmarks/*`** — specification placeholders only, per the spec's own "code lands in M1."
* **`regression/`** — empty by design; no bug has been fixed yet.
