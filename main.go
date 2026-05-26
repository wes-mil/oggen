package main

import (
	"bufio"
	"flag"
	"fmt"
	"log/slog"
	"math"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/bytedance/sonic"
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
	outputFormat    string
)

const (
	outputFormatJSON  = "json"
	outputFormatJSONL = "jsonl"
)

type outputTargets struct {
	JSON  string
	Nodes string
	Edges string
}

type generationConfig struct {
	NumNodes        int
	NumTiers        int
	NumEdgesPerTier int
	Now             func() time.Time
	Rand            *rand.Rand
}

func main() {
	flag.IntVar(&numNodes, "n", 0, "total number of nodes (n > 0)")
	flag.IntVar(&numTiers, "t", 0, "total number of tiers (25 >= t > 0)")
	flag.IntVar(&numEdgesPerTier, "e", 0, "number of edges generated per tier")
	flag.StringVar(&outputFile, "o", fmt.Sprintf("opengraph-%s", time.Now().UTC().Format("20060102T150405Z")), "output base name")
	flag.StringVar(&outputFormat, "f", outputFormatJSON, "output format: json or jsonl")

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

	paths, err := outputPaths(outputFile, outputFormat)
	if err != nil {
		slog.Error("invalid output options", "err", err)
		os.Exit(1)
	}

	config := generationConfig{
		NumNodes:        numNodes,
		NumTiers:        numTiers,
		NumEdgesPerTier: numEdgesPerTier,
	}

	switch outputFormat {
	case outputFormatJSON:
		err = writeJSONOutput(paths.JSON, config)
	case outputFormatJSONL:
		err = writeJSONLOutput(paths.Nodes, paths.Edges, config)
	}
	if err != nil {
		slog.Error("failed to write output", "err", err)
		os.Exit(1)
	}
}

func writeJSONOutput(outputFile string, config generationConfig) (err error) {
	f, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer closeFile(f, &err)

	config = config.withDefaults()
	writer := bufio.NewWriter(f)
	encoder := sonic.ConfigDefault.NewEncoder(writer)

	if _, err := writer.WriteString(`{"graph":{"nodes":[`); err != nil {
		return err
	}

	firstNode := true
	if err := generateNodes(config, func(node Node) error {
		if !firstNode {
			if _, err := writer.WriteString(","); err != nil {
				return err
			}
		}
		firstNode = false
		return encoder.Encode(node)
	}); err != nil {
		return err
	}

	if _, err := writer.WriteString(`],"edges":[`); err != nil {
		return err
	}

	firstEdge := true
	if err := generateEdges(config, func(edge Edge) error {
		if !firstEdge {
			if _, err := writer.WriteString(","); err != nil {
				return err
			}
		}
		firstEdge = false
		return encoder.Encode(edge)
	}); err != nil {
		return err
	}

	if _, err := writer.WriteString(`]},"metadata":`); err != nil {
		return err
	}
	if err := encoder.Encode(Metadata{SourceKind: "OggenBase"}); err != nil {
		return err
	}
	if _, err := writer.WriteString("}"); err != nil {
		return err
	}
	if err := writer.Flush(); err != nil {
		return err
	}

	return nil
}

func writeJSONLOutput(nodesFile string, edgesFile string, config generationConfig) (err error) {
	nodes, err := os.Create(nodesFile)
	if err != nil {
		return err
	}
	defer closeFile(nodes, &err)

	edges, err := os.Create(edgesFile)
	if err != nil {
		return err
	}
	defer closeFile(edges, &err)

	config = config.withDefaults()

	nodeWriter := bufio.NewWriter(nodes)
	nodeEncoder := sonic.ConfigDefault.NewEncoder(nodeWriter)
	if err := generateNodes(config, func(node Node) error {
		return nodeEncoder.Encode(node)
	}); err != nil {
		return err
	}
	if err := nodeWriter.Flush(); err != nil {
		return err
	}

	edgeWriter := bufio.NewWriter(edges)
	edgeEncoder := sonic.ConfigDefault.NewEncoder(edgeWriter)
	if err := generateEdges(config, func(edge Edge) error {
		return edgeEncoder.Encode(edge)
	}); err != nil {
		return err
	}
	if err := edgeWriter.Flush(); err != nil {
		return err
	}

	return nil
}

func generateNodes(config generationConfig, visit func(Node) error) error {
	leading := idWidth(config.NumNodes)
	tierSize := generationTierSize(config)

	for i := range config.NumNodes {
		tier := i / tierSize
		node := Node{
			ID:    createId(i, leading),
			Kinds: []string{"OGGEN_NODE"},
			Properties: map[string]any{
				"centrality_tier": tier,
				"created_at":      config.Now(),
				"rohan?":          true,
			},
		}

		if err := visit(node); err != nil {
			return err
		}
	}

	return nil
}

func generateEdges(config generationConfig, visit func(Edge) error) error {
	leading := idWidth(config.NumNodes)
	tierSize := generationTierSize(config)

	for i := range config.NumNodes {
		id2 := config.Rand.Intn(config.NumNodes)
		if err := visit(createEdge(i, id2, leading)); err != nil {
			return err
		}
	}

	for i := range config.NumTiers {
		endingId := min((i+1)*tierSize, config.NumNodes)

		for range config.NumEdgesPerTier {
			id1 := config.Rand.Intn(endingId)
			id2 := config.Rand.Intn(endingId)
			if err := visit(createEdge(id1, id2, leading)); err != nil {
				return err
			}
		}
	}

	return nil
}

func createEdge(startID int, endID int, leading int) Edge {
	return Edge{
		Start: Connection{
			MatchBy: "id",
			Value:   createId(startID, leading),
		},
		End: Connection{
			MatchBy: "id",
			Value:   createId(endID, leading),
		},
		Kind: "OGGEN_EDGE",
	}
}

func outputPaths(base string, format string) (outputTargets, error) {
	if base == "" {
		return outputTargets{}, fmt.Errorf("output base name is required")
	}
	if err := validateOutputFormat(format); err != nil {
		return outputTargets{}, err
	}

	switch format {
	case outputFormatJSON:
		if strings.HasSuffix(base, ".json") {
			return outputTargets{JSON: base}, nil
		}
		return outputTargets{JSON: base + ".json"}, nil
	case outputFormatJSONL:
		stem := jsonlStem(base)
		return outputTargets{
			Nodes: stem + ".nodes.jsonl",
			Edges: stem + ".edges.jsonl",
		}, nil
	default:
		return outputTargets{}, fmt.Errorf("unsupported output format %q", format)
	}
}

func validateOutputFormat(format string) error {
	switch format {
	case outputFormatJSON, outputFormatJSONL:
		return nil
	default:
		return fmt.Errorf("unsupported output format %q; expected %q or %q", format, outputFormatJSON, outputFormatJSONL)
	}
}

func jsonlStem(base string) string {
	for _, suffix := range []string{".nodes.jsonl", ".edges.jsonl", ".jsonl", ".json"} {
		if strings.HasSuffix(base, suffix) {
			return strings.TrimSuffix(base, suffix)
		}
	}

	return base
}

func (config generationConfig) withDefaults() generationConfig {
	if config.Now == nil {
		config.Now = time.Now
	}
	if config.Rand == nil {
		config.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))
	}

	return config
}

func generationTierSize(config generationConfig) int {
	return max(1, config.NumNodes/config.NumTiers)
}

func idWidth(numNodes int) int {
	return int(math.Log10(float64(numNodes)))
}

func closeFile(f *os.File, err *error) {
	if closeErr := f.Close(); *err == nil && closeErr != nil {
		*err = closeErr
	}
}

func createId(i int, leading int) string {
	return fmt.Sprintf("oggen_%0*d", leading, i)
}
