# Regression Directory

Empty by design. No bug has been fixed yet — no code exists.

Per [docs/testing/test-vault-spec.md §4](../../../../docs/testing/test-vault-spec.md#4-corpus-rules):

* One file per fixed bug, added the moment the bug is fixed and never removed.
* The **file** is named for the issue: `issue-<n>-<short-description>.md`.
* The **edge-case ID** it registers under in the manifest is `EC-REG-<n>`, matching the issue number. The filename carries human context; the manifest ID carries machine identity. Where they appear to conflict, the manifest ID is authoritative.

This file exists only so the empty directory survives git, which does not track empty directories. Delete it the moment the first regression file lands.
