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

@test "counts every line of a multi-line svelte comment" {
  stage a.svelte $'<!--\n  one\n  two\n  three\n-->\n<div></div>'
  run "$HOOK"
  [ "$status" -eq 1 ]
  [[ "$output" == *"a.svelte:1"* ]]
}

@test "counts every line of an unstarred block comment" {
  stage a.ts $'/* one\n   two\n   three\n   four */\nexport const a = 1;'
  run "$HOOK"
  [ "$status" -eq 1 ]
}

@test "passes one-line block comments in a row of code" {
  stage a.ts $'/* one */\nexport const a = 1;\n<!-- not a comment in ts but harmless -->\nexport const b = 2;'
  run "$HOOK"
  [ "$status" -eq 0 ]
}

@test "checks a renamed file at its new path" {
  stage old.go $'package a\n\n// one\n// two\n// three\nfunc A() {}\n\nfunc B() {}\nfunc C() {}\nfunc D() {}'
  git commit -qm base --no-verify
  git mv old.go new.go
  stage new.go $'package a\n\n// one\n// two\n// three\n// four\nfunc A() {}\n\nfunc B() {}\nfunc C() {}\nfunc D() {}'
  run "$HOOK"
  [ "$status" -eq 1 ]
  [[ "$output" == *"new.go:3"* ]]
}

@test "checks a file whose name git would quote" {
  stage café.go $'package a\n\n// one\n// two\n// three\n// four\nfunc A() {}'
  run "$HOOK"
  [ "$status" -eq 1 ]
  [[ "$output" == *"café.go:3"* ]]
}

@test "does not count comment-like text inside a go raw string" {
  stage a_test.go $'package a\n\nvar in = `\n/* one\n   two\n   three\n   four */\n{"a": 1}\n`'
  run "$HOOK"
  [ "$status" -eq 0 ]
}

@test "ends a block only at its own closing marker" {
  stage a.svelte $'<!--\n  example: /* x */\n  two\n  three\n-->\n<div></div>'
  run "$HOOK"
  [ "$status" -eq 1 ]
}

@test "does not count lines that start with a multiplication" {
  stage a.ts $'export const total = a\n  * b\n  * c\n  * d\n  * e;'
  run "$HOOK"
  [ "$status" -eq 0 ]
}
