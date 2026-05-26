# Sonic JSON Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace every `encoding/json` use in the repo with explicit `github.com/bytedance/sonic` calls.

**Architecture:** Keep the current CLI, data structures, streaming writers, and tests. Add Sonic as the JSON dependency and update production and test call sites from `json.*` to explicit `sonic.ConfigDefault.*` or `sonic.*` calls.

**Tech Stack:** Go 1.25.3, `github.com/bytedance/sonic`, existing `testing` package.

---

## Files

- Modify: `main.go` to import Sonic and call `sonic.ConfigDefault.NewEncoder`.
- Modify: `main_test.go` to add a dependency-enforcement test, import Sonic, and call `sonic.ConfigDefault.NewDecoder` and `sonic.Unmarshal`.
- Modify: `go.mod` and create/update `go.sum` through `go get` / `go mod tidy`.

## Tasks

### Task 1: Add a failing dependency-enforcement test

- [ ] Add `TestRepositoryUsesSonicForJSON` to `main_test.go`.
- [ ] The test should scan root Go files and fail if any source file contains the forbidden import path built as `"encoding" + "/" + "json"`.
- [ ] Run `go test ./...`.
- [ ] Expected result: failure naming `main.go` because it still imports `encoding/json`.

### Task 2: Replace JSON implementation with Sonic

- [ ] Run `go get github.com/bytedance/sonic@latest`.
- [ ] Update `main.go` to import `github.com/bytedance/sonic` and call `sonic.ConfigDefault.NewEncoder`.
- [ ] Update `main_test.go` to import `github.com/bytedance/sonic` and call `sonic.ConfigDefault.NewDecoder` and `sonic.Unmarshal`.
- [ ] Run `gofmt -w main.go main_test.go`.
- [ ] Run `go mod tidy`.

### Task 3: Verify behavior

- [ ] Run `go test ./...`.
- [ ] Confirm the dependency-enforcement test passes and the existing JSON/JSONL output tests still parse generated output.
