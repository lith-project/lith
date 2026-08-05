#!/usr/bin/env bash
set -euo pipefail

# rewire-blocked-by.sh - Reconcile GitHub issue "Blocked by #N" prose into
# the issue_dependencies API, so tools/next-task.sh sees the real graph.
#
# Reads "Blocked by #N" lines from issue bodies, compares against the current
# blocked_by metadata, and reports (dry-run) or applies (--apply) the delta.
#
# Usage:
#   ./tools/rewire-blocked-by.sh              # dry-run (default)
#   ./tools/rewire-blocked-by.sh --apply      # apply missing edges only
#   ./tools/rewire-blocked-by.sh --milestone M1-A   # scope to one milestone

REPO="${LITH_REPO:-lith-project/lith}"
APPLY=false
MILESTONE=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --apply)     APPLY=true; shift ;;
    --milestone) MILESTONE="$2"; shift 2 ;;
    -h|--help)   sed -n '3,12p' "$0" | sed 's/^# \{0,1\}//'; exit 0 ;;
    *)           echo "Unknown flag: $1" >&2; exit 1 ;;
  esac
done

# -- Fetch issues (open only - these are what we reconcile dependencies for) --
echo "Fetching open issues from $REPO ..." >&2
LIST_ARGS=(--repo "$REPO" --state open --limit 200 --json number,body)
if [[ -n "$MILESTONE" ]]; then
  ISSUES=$(gh issue list "${LIST_ARGS[@]}" --milestone "$MILESTONE")
else
  ISSUES=$(gh issue list "${LIST_ARGS[@]}")
fi
OPEN_NUMBERS=$(jq -r '.[].number' <<<"$ISSUES")

# -- Parse "Blocked by #N[, #M]" from issue bodies -> edges (issue:blocker) --
parse_wanted_edges() {
  jq -r '
    .[]
    | .number as $n
    | .body // ""
    | split("\n")[]
    | select(test("(?i)blocked by"))
    | [ scan("#[0-9]+") | ltrimstr("#") | tonumber ]
    | .[]
    | "\($n):\(.)"
  ' <<<"$ISSUES"
}

# -- Read current blocked_by for one issue -> space-separated numbers --
current_blockers() {
  gh api --paginate --slurp \
    "repos/$REPO/issues/$1/dependencies/blocked_by?per_page=100" \
    2>/dev/null | jq -r 'add // [] | map(.number) | join(" ")' || echo ""
}

# -- Add one blocked_by edge via the dependencies REST API --
add_blocker() {
  local issue_num="$1" blocker_num="$2"
  # The dependencies API needs the database id of the blocker issue.
  local blocker_id
  blocker_id=$(gh api "repos/$REPO/issues/$blocker_num" --jq '.id' 2>/dev/null || echo "")
  if [[ -z "$blocker_id" ]]; then
    echo "FAIL: could not resolve database id for #$blocker_num" >&2
    return 1
  fi
  gh api --method POST \
    "repos/$REPO/issues/$issue_num/dependencies" \
    -f blocked_by_issue_id="$blocker_id" \
    --silent 2>&1
}

# ── Build the delta ───────────────────────────────────────────────────
echo "Reading current blocked_by metadata for each open issue..." >&2

WANTED=()
ALREADY=()
MISSING=()
SEEN_ISSUES=0

while IFS=: read -r issue blocker; do
  [[ -z "$issue" ]] && continue
  [[ "$issue" == "$blocker" ]] && continue
  WANTED+=("$issue:$blocker")
done < <(parse_wanted_edges)

# Deduplicate wanted edges and group by issue for efficient comparison
declare -A WANT_MAP  # issue_num -> "blocker1 blocker2 ..."
for edge in "${WANTED[@]:-}"; do
  [[ -z "$edge" ]] && continue
  iss="${edge%%:*}"; blk="${edge##*:}"
  WANT_MAP[$iss]="${WANT_MAP[$iss]:-} $blk"
done

for issue in $(printf '%s\n' "${!WANT_MAP[@]}" | sort -n); do
  SEEN_ISSUES=$((SEEN_ISSUES + 1))
  wanted_blockers=$(echo "${WANT_MAP[$issue]}" | tr ' ' '\n' | sort -un | tr '\n' ' ')
  current=$(current_blockers "$issue")
  current_sorted=$(echo "$current" | tr ' ' '\n' | sort -un | tr '\n' ' ')

  for blk in $wanted_blockers; do
    if echo " $current " | grep -q " $blk "; then
      ALREADY+=("$issue:$blk")
    else
      MISSING+=("$issue:$blk")
    fi
  done
done

# ── Report ────────────────────────────────────────────────────────────
echo ""
echo "============================================================"
echo " BLOCKED_BY RECONCILIATION"
echo "============================================================"
echo "  Open issues with 'Blocked by' prose: $SEEN_ISSUES"
echo "  Edges already wired:                 ${#ALREADY[@]}"
echo "  Edges missing:                       ${#MISSING[@]}"
echo ""

if [[ ${#ALREADY[@]} -gt 0 ]]; then
  echo "Already wired (no action):"
  for e in "${ALREADY[@]}"; do
    printf "  OK   #%-4s <- #%s\n" "${e%%:*}" "${e##*:}"
  done
  echo ""
fi

if [[ ${#MISSING[@]} -gt 0 ]]; then
  echo "Missing (will be added with --apply):"
  for e in "${MISSING[@]}"; do
    printf "  ADD  #%-4s <- #%s\n" "${e%%:*}" "${e##*:}"
  done
  echo ""
fi

if [[ ${#MISSING[@]} -eq 0 ]]; then
  echo "Graph is complete - nothing to do."
  exit 0
fi

if [[ "$APPLY" == "false" ]]; then
  echo "DRY RUN -- no changes made. Run with --apply to wire missing edges."
  exit 0
fi

# ── Apply ─────────────────────────────────────────────────────────────
echo "Applying ${#MISSING[@]} missing edges..."
echo ""
OK=0; FAIL=0
for e in "${MISSING[@]}"; do
  iss="${e%%:*}"; blk="${e##*:}"
  if add_blocker "$iss" "$blk"; then
    printf "  OK   #%-4s <- #%s\n" "$iss" "$blk"
    OK=$((OK + 1))
  else
    printf "  FAIL #%-4s <- #%s\n" "$iss" "$blk"
    FAIL=$((FAIL + 1))
  fi
done
echo ""
echo "Done: $OK added, $FAIL failed."
