#!/usr/bin/env bash
#
# next-task.sh — print the next actionable issue for a coding agent.
#
# An issue is actionable when every one of these holds:
#   * it is open and in the requested milestone
#   * it is leaf work — it has no sub-issues, so it is not an epic tracker
#   * it is unclaimed — no assignee, and no agent:wip / agent:blocked label
#   * it is unblocked — every issue it is blocked_by is closed
#
# Eligibility comes from the blocked_by graph (a task is actionable only when it has
# no open blockers), not from prose in ROADMAP.md. When multiple tasks are actionable,
# output is sorted by issue number for stability.
#
# Usage:
#   tools/next-task.sh                # next task in M1-A
#   tools/next-task.sh M1-B           # next task in M1-B
#   tools/next-task.sh M1-A --all     # every actionable task, not just the first
#
# Environment:
#   LITH_REPO   override the target repository (default: lith-project/lith)
#
# Exit codes:
#   0  an actionable task was printed
#   1  usage error, missing dependency, or API failure
#   2  the milestone exists but nothing is actionable right now
#
set -euo pipefail

REPO="${LITH_REPO:-lith-project/lith}"
DEFAULT_MILESTONE="M1-A"
CLAIM_LABELS='["agent:wip","agent:blocked"]'
PAGE_SIZE=100
EXIT_NONE_ACTIONABLE=2

die() {
	printf 'error: %s\n' "$1" >&2
	exit 1
}

usage() {
	sed -n '3,26p' "$0" | sed 's/^# \{0,1\}//'
}

require_command() {
	command -v "$1" >/dev/null 2>&1 || die "$1 is required but was not found on PATH"
}

fetch_paginated_array() {
	gh api --paginate --slurp "$1" | jq -c 'add // []'
}

# Resolve a milestone title prefix ("M1-A") to its numeric id. Prefix matching
# keeps the caller from having to type the "·" separator in the full title.
resolve_milestone() {
	prefix="$1"
	milestones_json=$(fetch_paginated_array \
		"repos/$REPO/milestones?state=all&per_page=$PAGE_SIZE") ||
		die "failed to list milestones for $REPO"
	matches=$(jq -c --arg p "$prefix" \
		'map(select(.title | startswith($p)))' <<<"$milestones_json")
	match_count=$(jq -r 'length' <<<"$matches")
	[[ "$match_count" -gt 0 ]] ||
		die "no milestone in $REPO with a title starting with '$prefix'"
	[[ "$match_count" -eq 1 ]] ||
		die "milestone prefix '$prefix' is ambiguous in $REPO"
	number=$(jq -r '.[0].number' <<<"$matches")
	printf '%s\n' "$number"
}

fetch_open_issues() {
	fetch_paginated_array \
		"repos/$REPO/issues?milestone=$1&state=open&per_page=$PAGE_SIZE" ||
		die "failed to list open issues for milestone $1 in $REPO"
}

# Everything except the blocked_by check, which costs one API call per issue.
candidate_numbers() {
	jq -r --argjson claimed "$CLAIM_LABELS" '
		def is_pull_request: has("pull_request");
		def is_tracker:      (.sub_issues_summary.total // 0) > 0;
		def is_task:         .type.name == "Task";
		def is_claimed:      (.assignees | length) > 0
		                     or ([.labels[].name] | any(IN($claimed[])));

		map(select(is_task and ((is_pull_request or is_tracker or is_claimed) | not)))
		| sort_by(.number)
		| .[].number
	' <<<"$1"
}

open_blockers() {
	blockers_json=$(fetch_paginated_array \
		"repos/$REPO/issues/$1/dependencies/blocked_by?per_page=$PAGE_SIZE") ||
		die "failed to read blocked_by dependencies for issue #$1"
	jq -r 'map(select(.state == "open") | "#\(.number)") | join(", ")' <<<"$blockers_json"
}

print_issue() {
	jq -r --argjson n "$2" '
		def terminal_safe:
			gsub("[\u0000-\u001f\u007f-\u009f]"; "");
		map(select(.number == $n)) | .[0]
		| "#\(.number)  \(.title | terminal_safe)\n        \(.html_url | terminal_safe)"
	' <<<"$1"
}

main() {
	require_command gh
	require_command jq

	milestone_prefix="$DEFAULT_MILESTONE"
	list_all=false
	for arg in "$@"; do
		case "$arg" in
		--all) list_all=true ;;
		-h | --help)
			usage
			exit 0
			;;
		-*) die "unknown flag: $arg" ;;
		*) milestone_prefix="$arg" ;;
		esac
	done

	milestone_number=$(resolve_milestone "$milestone_prefix")
	issues_json=$(fetch_open_issues "$milestone_number")
	candidates=$(candidate_numbers "$issues_json")
	if [[ -z "$candidates" ]]; then
		printf 'No unclaimed leaf issues open in %s.\n' "$milestone_prefix" >&2
		exit "$EXIT_NONE_ACTIONABLE"
	fi

	found=false
	first_blocked_number=""
	first_blockers=""
	while IFS= read -r number; do
		[[ -n "$number" ]] || continue
		blockers=$(open_blockers "$number")
		if [[ -n "$blockers" ]]; then
			if [[ -z "$first_blocked_number" ]]; then
				first_blocked_number="$number"
				first_blockers="$blockers"
			fi
			continue
		fi

		print_issue "$issues_json" "$number"
		found=true
		[[ "$list_all" == true ]] || break
	done <<<"$candidates"

	if [[ "$found" == false ]]; then
		printf 'Nothing actionable in %s: every unclaimed issue still has an open blocker.\n' \
			"$milestone_prefix" >&2
		if [[ -n "$first_blocked_number" ]]; then
			printf 'First blocked candidate #%s is waiting on %s.\n' \
				"$first_blocked_number" "$first_blockers" >&2
		fi
		exit "$EXIT_NONE_ACTIONABLE"
	fi
}

main "$@"
