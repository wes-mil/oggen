# oggen

Quickly create arbitrarily large BloodHound OpenGraph payloads with tiered centrality

## Install

```bash
go install github.com/wes-mil/oggen
```

## Usage

```bash
oggen -n 1000 -t 10 -e 50 -o output.json
```

`-n` total nodes to create

`-t` the number of centrality tiers

`-e` the number of edges created per tier

`-o` output file name

## How it works

Every node is guaranteed 1 edge connecting to another random node.

Then, in batches, random edges are created connecting nodes in each tier such that there will always be more connectedness for lower tiers.

For example:

If oggen is run with `n=1000`, `-t=10`, and `-e=50` there will be 1000 nodes created each with a random edge. Then, `e` edges will be created for each "tier" of nodes and lower. In this case, tier 0 contains nodes 0-99 (tiersize is $n/t$) so 50 edges will be created randomly connecting nodes 0-99. On the next iteration, edges are randomly created for tier 1 _and_ tier 0. So for tier 1, 50 edges are randomly created connecting nodes 0-199.

This process organically creates a psuedo-uniform distribution of nodes with varying centrality and connectedness based on the parameters.
