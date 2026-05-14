# Streaming Output Formats Design

## Goal

Add an output format flag so `oggen` can create either one normal OpenGraph JSON file or two JSONL files: one for nodes and one for edges. Both modes must stream output instead of constructing the full graph in memory.

## Interface

Add `-f` as the output format flag with allowed values `json` and `jsonl`. The default is `json`.

Treat `-o` as the output base name:

- `-f json -o graph` writes `graph.json`.
- `-f jsonl -o graph` writes `graph.nodes.jsonl` and `graph.edges.jsonl`.

If `-o` already includes the expected suffix for the chosen format, the program must avoid duplicating it.

## Architecture

Keep the existing data shapes for `Node`, `Edge`, `Connection`, and `Metadata`. Remove the in-memory `OpenGraph` assembly from the generation path.

Introduce streaming writer functions:

- A JSON writer that writes the OpenGraph object framing, streams each node into the `graph.nodes` array, streams each edge into the `graph.edges` array, and then closes the JSON document.
- A JSONL writer that opens the node and edge output files and writes each generated record as one JSON object per line.

Use shared generation loops for both output formats so node and edge behavior stays consistent.

## Data Flow

Compute the common generation inputs once: node count, tier count, edges per tier, tier size, and ID width.

Then stream records in this order:

1. Generate all nodes from `0` to `n - 1`.
2. Generate the guaranteed one random edge per node.
3. Generate the tier-weighted random edges.

JSON mode writes all three phases into one OpenGraph document. JSONL mode writes node records to the node file and edge records to the edge file.

## Error Handling

Validate required numeric flags as today. Validate `-f` and exit with a logged error for unsupported values.

Fail fast if any output file cannot be created or if any encoder/write operation fails. Close files after writing and report close errors when they occur.

## Testing

Add focused Go tests for:

- Output base-name derivation for JSON and JSONL.
- Rejection of invalid formats.
- Small streamed JSON output parses as an OpenGraph payload with expected node and edge counts.
- Small streamed JSONL output creates parseable node and edge lines with expected counts.
