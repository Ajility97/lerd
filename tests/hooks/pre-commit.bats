#!/usr/bin/env bats
# Tests for .githooks/pre-commit, the comment block length check.

HOOK="$BATS_TEST_DIRNAME/../../.githooks/pre-commit"

setup() {
  REPO="$BATS_TMPDIR/hookrepo-$$"
  mkdir -p "$REPO"
  cd "$REPO"
  git init -q
  git config user.email t@example.test
  git config user.name t
}

teardown() {
  rm -rf "$BATS_TMPDIR/hookrepo-$$"
}

stage() { printf '%s\n' "$2" > "$1"; git add "$1"; }

@test "passes a three-line comment" {
  stage a.go $'package a\n\n// one\n// two\n// three\nfunc A() {}'
  run "$HOOK"
  [ "$status" -eq 0 ]
}

@test "rejects a four-line comment and names where it is" {
  stage a.go $'package a\n\n// one\n// two\n// three\n// four\nfunc A() {}'
  run "$HOOK"
  [ "$status" -eq 1 ]
  [[ "$output" == *"a.go:3"* ]]
}

@test "rejects one line that grows an old block past three" {
  stage a.go $'package a\n\n// one\n// two\n// three\nfunc A() {}'
  git commit -qm base --no-verify
  stage a.go $'package a\n\n// one\n// two\n// three\n// four\nfunc A() {}'
  run "$HOOK"
  [ "$status" -eq 1 ]
}

@test "leaves an untouched long block alone" {
  stage a.go $'package a\n\n// one\n// two\n// three\n// four\nfunc A() {}'
  git commit -qm base --no-verify
  stage a.go $'package a\n\n// one\n// two\n// three\n// four\nfunc A() {}\n\nfunc B() {}'
  run "$HOOK"
  [ "$status" -eq 0 ]
}

@test "checks TypeScript and Svelte, and block comments" {
  stage a.ts $'/*\n * one\n * two\n * three\n */\nexport const a = 1;'
  run "$HOOK"
  [ "$status" -eq 1 ]
  [[ "$output" == *"a.ts:1"* ]]
}

@test "does not count a go directive as comment" {
  stage a.go $'package a\n\n// one\n// two\n// three\n//go:embed x\nvar x string'
  run "$HOOK"
  [ "$status" -eq 0 ]
}

@test "ignores files it does not check" {
  stage a.md $'# one\n# two\n# three\n# four'
  run "$HOOK"
  [ "$status" -eq 0 ]
}
