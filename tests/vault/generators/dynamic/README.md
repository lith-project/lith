# Dynamic Generator

Specified in [docs/testing/test-vault-spec.md §5.6 and §8](../../../../docs/testing/test-vault-spec.md#56-generator-only-cases-ec-fs-x--ec-dyn--never-committed). Materializes `EC-DYN-*` timing and mutation scenarios — modification during a scan, clock skew, atomic-save patterns, directory renames — none of which are static content and so cannot be committed as files.

`EC-DYN-002` (identical `mtime` and size, changed content) is the highest-priority case here: it is what forces content-hash-based change detection rather than stat-only, per [RFC-0003/C-7](../../../../rfcs/0003-storage-engine.md#c-7-content-hash-authority) and [RFC-0004/C-4](../../../../rfcs/0004-indexing.md#c-4-scan-hashes-every-file).

No code here yet — implementation lands in M1.
