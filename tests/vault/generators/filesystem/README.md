# Filesystem Generator

Specified in [docs/testing/test-vault-spec.md §5.6 and §8](../../../../docs/testing/test-vault-spec.md#56-generator-only-cases-ec-fs-x--ec-dyn--never-committed). Materializes `EC-FS-X-*` at test time in a temporary directory — case collisions, path-length limits, permission cases, hard links, oversized files — none of which can survive a git checkout across platforms.

No code here yet — implementation lands in M1.
