#!/usr/bin/env bash
# Checks scripts/next-version.sh against throwaway git repositories.
# Run: bash scripts/next-version_test.sh
set -euo pipefail

script="$(cd "$(dirname "$0")" && pwd)/next-version.sh"
failures=0

# repo <commit message>... creates a fresh repository with one commit per message.
repo() {
  dir=$(mktemp -d)
  git -C "$dir" init -q
  for message in "$@"; do commit -m "$message"; done
}

tag() { git -C "$dir" tag "$1"; }
commit() { git -C "$dir" -c user.name=t -c user.email=t@t commit -q --allow-empty "$@"; }

expect() {
  got=$(cd "$dir" && bash "$script")
  if [ "$got" != "$2" ]; then
    echo "FAIL $1: expected $2, got $got"
    failures=$((failures + 1))
  else
    echo "ok   $1"
  fi
  rm -rf "$dir"
}

repo "feat: first"
expect "no tag starts from 0.1.0 and bumps minor on feat" 0.2.0

repo "chore: base"; tag v1.2.3
commit -m "fix: bug"
expect "fix bumps patch" 1.2.4

repo "chore: base"; tag v1.2.3
commit -m "ci: pipeline"
expect "any other type bumps patch" 1.2.4

repo "chore: base"; tag v1.2.3
commit -m "fix: bug"
commit -m "feat(api): endpoint"
expect "feat wins over fix and resets patch" 1.3.0

repo "chore: base"; tag v1.2.3
commit -m "feat(api)!: drop field"
expect "bang marks a breaking change" 2.0.0

repo "chore: base"; tag v1.2.3
commit -m "fix: rename" -m "BREAKING CHANGE: field renamed"
expect "BREAKING CHANGE footer bumps major" 2.0.0

repo "chore: base"; tag v1.2.3
commit -m "Merge pull request #9 from org/feat/x" -m "feat: inside merge body"
expect "conventional type in a merge body counts" 1.3.0

repo "chore: base"; tag v0.9.0; tag v1.2.3
expect "HEAD already tagged returns that version (re-run)" 1.2.3

[ "$failures" -eq 0 ] || exit 1
