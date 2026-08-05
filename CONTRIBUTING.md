# Contributing to Lith

Thank you for your interest in contributing to **Lith**! We welcome contributions from developers, researchers, and knowledge-management enthusiasts.

---

## Architectural Principles First

Before proposing features or writing code, please read our [PROJECT_PRINCIPLES.md](PROJECT_PRINCIPLES.md). All contributions must align with these core tenets.

Lith is an **RFC-driven project**. Major architectural additions, new domain abstractions, or breaking protocol changes are proposed and tracked as **GitHub Issues** labeled `rfc`.

---

## How to Contribute

### 1. Architectural & Feature Proposals (RFCs)

If you have an architectural idea or major feature proposal:

1. Open a new GitHub Issue using the **RFC Proposal** issue template.
2. Discuss the proposal in [GitHub Discussions](https://github.com/lith-project/lith/discussions) under the `RFCs` or `Architecture` category or directly on the RFC Issue.
3. Once consensus is reached, accepted RFCs are indexed in [ARCHITECTURE.md](ARCHITECTURE.md).

### 2. Reporting Bugs & Feature Requests

* Search existing GitHub Issues before opening a new one.
* Use the appropriate issue template (**Bug Report** or **Feature Request**).
* Provide clear reproduction steps and context.

### 3. Submitting Pull Requests

1. Fork the repository and create a descriptive branch name (e.g., `feature/link-graph` or `fix/parser-edge-case`). Cut it from `main` and target `main` — it is the only long-lived branch, and the CI workflows and issue auto-closing are both keyed to it.
2. Keep commits concise and logical. Write descriptive commit messages.
3. Ensure all tests and static analysis pass before requesting review.
4. Ensure your PR description references any relevant GitHub Issue / RFC numbers.

---

## Development Standards (Go)

When code development begins:

* **Format**: All code must be formatted using standard `gofmt` / `goimports`.
* **Testing**: New functionality must include unit and integration tests.
* **Linting**: Code must pass `golangci-lint` without warnings.
* **Documentation**: Exported packages, types, and functions must have clear Go docstrings.

---

## Code of Conduct

Lith follows the [Contributor Covenant Code of Conduct](CODE_OF_CONDUCT.md). By participating in this project, you agree to abide by its terms.
