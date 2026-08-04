# Scale Generator

Specified in [docs/testing/test-vault-spec.md §8](../../../../docs/testing/test-vault-spec.md#8-generator). Produces the S/M/L benchmark vaults ([§9](../../../../docs/testing/test-vault-spec.md#9-benchmark-tiers)).

No code here yet — implementation lands in M1. This directory exists so the layout is stable before then. Never write generated output into `corpus/`; the spec makes that a hard error.
