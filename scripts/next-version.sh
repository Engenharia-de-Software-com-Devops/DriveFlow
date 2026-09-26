#!/usr/bin/env bash
# Prints the next semantic version (without the leading "v") from the
# Conventional Commits since the last vX.Y.Z tag reachable from HEAD:
#   "type!:" or a "BREAKING CHANGE:" footer -> major
#   "feat:" / "feat(scope):"                -> minor
#   anything else                           -> patch
# If HEAD already carries a version tag (a pipeline re-run), prints that one.
# Needs the full history and tags (actions/checkout with fetch-depth: 0).
set -euo pipefail

current=$(git tag --points-at HEAD --list 'v[0-9]*.[0-9]*.[0-9]*' --sort=-v:refname | head -n 1)
if [ -n "$current" ]; then
  echo "${current#v}"
  exit 0
fi

last=$(git describe --tags --abbrev=0 --match 'v[0-9]*.[0-9]*.[0-9]*' 2>/dev/null || true)
if [ -n "$last" ]; then
  range="$last..HEAD"
else
  # ponytail: 0.1.0 is the last release in CHANGELOG.md, shipped before tags existed.
  last=v0.1.0
  range=HEAD
fi

IFS=. read -r major minor patch <<<"${last#v}"
messages=$(git log --format='%s%n%b' "$range")

if grep -qE '^[a-z]+(\([^)]*\))?!:|^BREAKING CHANGE:' <<<"$messages"; then
  echo "$((major + 1)).0.0"
elif grep -qE '^feat(\([^)]*\))?:' <<<"$messages"; then
  echo "$major.$((minor + 1)).0"
else
  echo "$major.$minor.$((patch + 1))"
fi
