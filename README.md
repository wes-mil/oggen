# oggen

Quickly create arbitrarily large BloodHound OpenGraph payloads with tiered centrality

## Install

```bash
go install github.com/wes-mil/oggen
```

## Usage

```bash
oggen -n 1000 -t 10 -e 50 -node-kinds 10 -edge-kinds 10 -f json -o output
```

`-n` total nodes to create

`-t` the number of centrality tiers

`-e` the number of edges created per tier

`-node-kinds` number of possible node kinds, from 1 to 10 (default: 1)

`-edge-kinds` number of possible edge kinds, from 1 to 10 (default: 1)

`-f` output format: `json` for one OpenGraph file, or `jsonl` for split node and edge files

`-o` output base name

For `-f json`, `-o output` writes `output.json`.

For `-f jsonl`, `-o output` writes `output.nodes.jsonl` and `output.edges.jsonl`.

Each node gets one randomly selected kind from `OGGEN_NODE_1` through `OGGEN_NODE_N`, where `N` is `-node-kinds`. Each edge gets one randomly selected kind from `OGGEN_EDGE_1` through `OGGEN_EDGE_N`, where `N` is `-edge-kinds`.

Every edge also has a `properties` object with `created_at` (timestamp), `weight` (random integer from 1 to 100), and `active` (random boolean).

## How it works

Every node is guaranteed 1 edge connecting to another random node.

Then, in batches, random edges are created connecting nodes in each tier such that there will always be more connectedness for lower tiers.

For example:

If oggen is run with `n=1000`, `-t=10`, and `-e=50` there will be 1000 nodes created each with a random edge. Then, `e` edges will be created for each "tier" of nodes and lower. In this case, tier 0 contains nodes 0-99 (tiersize is $n/t$) so 50 edges will be created randomly connecting nodes 0-99. On the next iteration, edges are randomly created for tier 1 _and_ tier 0. So for tier 1, 50 edges are randomly created connecting nodes 0-199.

This process organically creates a psuedo-uniform distribution of nodes with varying centrality and connectedness based on the parameters.

## Notes

I was able to generate and render a full graph based on the following command, but any larger was too heavy for the UI

```bash
oggen -n 500 -t 10 -e 50
```
