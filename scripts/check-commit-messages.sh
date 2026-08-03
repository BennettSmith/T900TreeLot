#!/usr/bin/env bash
set -euo pipefail

export LC_ALL=C

readonly allowed_types="build|chore|ci|docs|feat|fix|perf|refactor|revert|style|test"
readonly subject_pattern="^(${allowed_types})(\\([a-z0-9][a-z0-9._/-]*\\))?!?: [[:graph:]].*$"

if [[ "$#" -ne 2 ]]; then
  echo "usage: $0 <base-commit> <head-commit>" >&2
  exit 2
fi

base="$1"
head="$2"
git rev-parse --verify --quiet "${base}^{commit}" >/dev/null || {
  echo "base commit is not available: $base" >&2
  exit 2
}
git rev-parse --verify --quiet "${head}^{commit}" >/dev/null || {
  echo "head commit is not available: $head" >&2
  exit 2
}

failed=0
while IFS= read -r commit; do
  subject="$(git show --no-patch --format=%s "$commit")"
  if [[ ! "$subject" =~ $subject_pattern ]]; then
    short_commit="$(git rev-parse --short "$commit")"
    echo "Invalid commit message at ${short_commit}: ${subject}" >&2
    failed=1
  fi
done < <(git rev-list --reverse "${base}..${head}")

if (( failed )); then
  echo "Expected: <type>[optional scope][!]: <description>" >&2
  echo "Allowed types: ${allowed_types//|/, }" >&2
  exit 1
fi
