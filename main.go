package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"math"
	"math/rand"
	"os"
	"time"
)

type OpenGraph struct {
	Graph    Graph    `json:"graph"`
	Metadata Metadata `json:"metadata"`
}

type Graph struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}

type Node struct {
	ID         string         `json:"id"`
	Kinds      []string       `json:"kinds"`
	Properties map[string]any `json:"properties"`
}

type Edge struct {
	Start Connection `json:"start"`
	End   Connection `json:"end"`
	Kind  string     `json:"kind"`
}

type Connection struct {
	MatchBy string `json:"match_by"`
	Value   string `json:"value"`
}

type Metadata struct {
	SourceKind string `json:"source_kind"`
}

var (
	numNodes        int
	numTiers        int
	numEdgesPerTier int
	outputFile      string
)

func main() {
	flag.IntVar(&numNodes, "n", 0, "total number of nodes (n > 0)")
	flag.IntVar(&numTiers, "t", 0, "total number of tiers (25 >= t > 0)")
	flag.IntVar(&numEdgesPerTier, "e", 0, "number of edges generated per tier")
	flag.StringVar(&outputFile, "o", fmt.Sprintf("opengraph-%s.json", time.Now().UTC().Format("20060102T150405Z")), "output file name")

	flag.Parse()

	if numNodes < 1 {
		slog.Error("-n is required and must satisfy n > 0", "n", numNodes)
		os.Exit(1)
	}

	if numTiers < 1 || numTiers > 25 {
		slog.Error("-t is required and must satisfy 25 >= t > 0", "t", numTiers)
		os.Exit(1)
	}

	if numEdgesPerTier < 1 {
		slog.Error("-e is required and must satisfy e > 0", "e", numEdgesPerTier)
		os.Exit(1)
	}

	leading := int(math.Log10(float64(numNodes)))

	og := OpenGraph{
		Metadata: Metadata{
			SourceKind: "OggenBase",
		},
		Graph: Graph{
			Nodes: make([]Node, numNodes),
			Edges: make([]Edge, 0),
		},
	}

	tierSize := numNodes / numTiers

	for i := range numNodes {
		tier := i / tierSize

		og.Graph.Nodes[i] = Node{
			ID:    createId(i, leading),
			Kinds: []string{"OGGEN_NODE"},
			Properties: map[string]any{
				"centrality_tier": tier,
				"created_at":      time.Now(),
				"rohan?":          true,
			},
		}

		id2 := rand.Intn(numNodes)
		og.Graph.Edges = append(og.Graph.Edges, Edge{
			Start: Connection{
				MatchBy: "id",
				Value:   createId(i, leading),
			},
			End: Connection{
				MatchBy: "id",
				Value:   createId(id2, leading),
			},
			Kind: "OGGEN_EDGE",
		})
	}

	for i := range numTiers {
		endingId := min((i+1)*tierSize, numNodes)

		for range numEdgesPerTier {
			id1 := rand.Intn(endingId)
			id2 := rand.Intn(endingId)

			og.Graph.Edges = append(og.Graph.Edges, Edge{
				Start: Connection{
					MatchBy: "id",
					Value:   createId(id1, leading),
				},
				End: Connection{
					MatchBy: "id",
					Value:   createId(id2, leading),
				},
				Kind: "OGGEN_EDGE",
			})
		}
	}

	fmt.Println(og)

	f, err := os.Create(outputFile)
	if err != nil {
		slog.Error("failed to create output file", "err", err)
		os.Exit(1)
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "	")
	err = encoder.Encode(og)
	if err != nil {
		slog.Error("failed to write json", "err", err)
		os.Exit(1)
	}
}

func createId(i int, leading int) string {
	return fmt.Sprintf("oggen_%0*d", leading, i)
}
