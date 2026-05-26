# Sonic JSON Design

## Goal

Replace every `encoding/json` use in the repo with `github.com/bytedance/sonic`.

## Interface

The CLI behavior and output formats stay the same. `oggen` continues to write OpenGraph JSON for `-f json` and node/edge JSONL files for `-f jsonl`.

## Architecture

Use Sonic directly at each JSON call site. Import it as `github.com/bytedance/sonic` and call `sonic.ConfigDefault.NewEncoder`, `sonic.ConfigDefault.NewDecoder`, and `sonic.Unmarshal` explicitly so the active JSON implementation is visible in production code and tests.

Do not add a wrapper package. The repo has only a few JSON call sites, and direct Sonic usage keeps the change simple.

## Data Flow

The generation and streaming flow is unchanged:

1. Stream generated nodes through a Sonic encoder.
2. Stream generated edges through a Sonic encoder.
3. Decode and unmarshal generated output in tests with Sonic.

## Error Handling

Preserve existing error paths. JSON encoding, decoding, and unmarshalling errors should continue to be returned or reported exactly where the current code handles them.

## Testing

Update tests to use Sonic for parsing JSON and JSONL output. Run `go test ./...` after the dependency change.
