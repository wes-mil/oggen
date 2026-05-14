# Streaming Output Formats Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `-f json|jsonl` output selection and stream both OpenGraph JSON and split JSONL outputs.

**Architecture:** Keep the CLI in `main.go`, add small helpers for format validation, output path derivation, generation, and streaming writers. Tests exercise helper behavior and parse real small outputs from the streaming paths.

**Tech Stack:** Go standard library: `flag`, `encoding/json`, `os`, `path/filepath`, `strings`, `testing`.

---

## Files

- Modify: `main.go` for CLI flag parsing, output path derivation, streaming writers, and shared generation.
- Create: `main_test.go` for output path, format validation, JSON parsing, and JSONL parsing tests.
- Modify: `README.md` for the new `-f` flag and `-o` base-name behavior.

## Tasks

### Task 1: Tests for output names and format validation

- [ ] Create `main_test.go` with tests for `outputPaths` and `validateOutputFormat`.
- [ ] Run `go test ./...` and confirm it fails because helpers do not exist.
- [ ] Implement the helpers in `main.go`.
- [ ] Run `go test ./...` and confirm these tests pass.

### Task 2: Tests for streaming JSON and JSONL

- [ ] Add tests that call `writeJSONOutput` and `writeJSONLOutput` with small counts.
- [ ] Run `go test ./...` and confirm it fails because streaming writers do not exist.
- [ ] Implement generation helpers and streaming writers in `main.go`.
- [ ] Run `go test ./...` and confirm all tests pass.

### Task 3: Wire CLI and docs

- [ ] Update `main` to parse `-f`, validate it, derive output files, and call the correct streaming writer.
- [ ] Update `README.md` usage and flag descriptions.
- [ ] Run `gofmt -w main.go main_test.go`.
- [ ] Run `go test ./...`.
- [ ] Run small manual commands for `-f json` and `-f jsonl` in a temp directory and parse/count output.
