#!/usr/bin/env bash
set -euo pipefail

ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
TEMP_ROOT=$(mktemp -d "${TMPDIR:-/tmp}/lith-next-task.XXXXXX")
trap 'rm -rf "$TEMP_ROOT"' EXIT
mkdir -p "$TEMP_ROOT/bin"

cat >"$TEMP_ROOT/bin/gh" <<'FAKE_GH'
#!/usr/bin/env bash
set -euo pipefail

[[ " $* " == *" --paginate "* && " $* " == *" --slurp "* ]] || exit 91
endpoint="${*: -1}"

if [[ "${FAKE_GH_MODE:-}" == fail ]]; then
	exit 1
fi

case "$endpoint" in
repos/*/milestones*)
	printf '%s\n' '[[],[{"number":9,"title":"M1-A · Lifecycle"},{"number":10,"title":"M1-B · Vault Loading"}]]'
	;;
repos/*/issues\?milestone=9*)
	if [[ "${FAKE_GH_MODE:-}" == blocked ]]; then
		printf '%s\n' '[[{"number":40,"title":"blocked","html_url":"https://example.test/40","type":{"name":"Task"},"assignees":[],"labels":[],"sub_issues_summary":{"total":0}}],[]]'
	else
		printf '%s\n' '[[{"number":14,"title":"tracker","html_url":"https://example.test/14","type":{"name":"Epic"},"assignees":[],"labels":[],"sub_issues_summary":{"total":1}},{"number":18,"title":"wip","html_url":"https://example.test/18","type":{"name":"Task"},"assignees":[],"labels":[{"name":"agent:wip"}],"sub_issues_summary":{"total":0}},{"number":19,"title":"empty epic","html_url":"https://example.test/19","type":{"name":"Epic"},"assignees":[],"labels":[],"sub_issues_summary":{"total":0}},{"number":20,"title":"claimed","html_url":"https://example.test/20","type":{"name":"Task"},"assignees":[{"login":"agent"}],"labels":[],"sub_issues_summary":{"total":0}},{"number":21,"title":"agent blocked","html_url":"https://example.test/21","type":{"name":"Task"},"assignees":[],"labels":[{"name":"agent:blocked"}],"sub_issues_summary":{"total":0}},{"number":30,"title":"pull request","html_url":"https://example.test/30","type":{"name":"Task"},"assignees":[],"labels":[],"sub_issues_summary":{"total":0},"pull_request":{}}],[{"number":40,"title":"blocked","html_url":"https://example.test/40","type":{"name":"Task"},"assignees":[],"labels":[],"sub_issues_summary":{"total":0}},{"number":41,"title":"ready\u001b task","html_url":"https://example.test/41\u0007","type":{"name":"Task"},"assignees":[],"labels":[],"sub_issues_summary":{"total":0}}]]'
	fi
	;;
repos/*/issues/40/dependencies/blocked_by*)
	printf '%s\n' '[[{"number":26,"state":"open"}],[]]'
	;;
repos/*/issues/41/dependencies/blocked_by*)
	printf '%s\n' '[[],[{"number":39,"state":"closed"}]]'
	;;
*)
	exit 92
	;;
esac
FAKE_GH
chmod +x "$TEMP_ROOT/bin/gh"

PATH="$TEMP_ROOT/bin:$PATH"

output=$("$ROOT/tools/next-task.sh" M1-A --all)
expected=$'#41  ready task\n        https://example.test/41'
[[ "$output" == "$expected" ]]

set +e
blocked_output=$(FAKE_GH_MODE=blocked "$ROOT/tools/next-task.sh" M1-A --all 2>&1)
blocked_rc=$?
missing_output=$("$ROOT/tools/next-task.sh" M9-Z 2>&1)
missing_rc=$?
ambiguous_output=$("$ROOT/tools/next-task.sh" M1 2>&1)
ambiguous_rc=$?
failure_output=$(FAKE_GH_MODE=fail "$ROOT/tools/next-task.sh" M1-A 2>&1)
failure_rc=$?
set -e

[[ "$blocked_rc" -eq 2 ]]
[[ "$blocked_output" == "Nothing actionable in M1-A: every unclaimed issue still has an open blocker." ]]
[[ "$missing_rc" -eq 1 ]]
[[ "$missing_output" == "error: no milestone in lith-project/lith with a title starting with 'M9-Z'" ]]
[[ "$ambiguous_rc" -eq 1 ]]
[[ "$ambiguous_output" == "error: milestone prefix 'M1' is ambiguous in lith-project/lith" ]]
[[ "$failure_rc" -eq 1 ]]
[[ "$failure_output" == "error: failed to list milestones for lith-project/lith" ]]

printf 'next-task selector tests passed\n'
