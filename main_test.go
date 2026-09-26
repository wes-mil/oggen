package main

import (
	"bufio"
	"go/parser"
	"go/token"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/bytedance/sonic"
)

func TestRepositoryUsesSonicForJSON(t *testing.T) {
	forbiddenImport := "encoding" + "/" + "json"
	requiredImport := "github.com/bytedance/sonic"
	importsSonic := false

	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("failed to read repo root: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}

		file, err := parser.ParseFile(token.NewFileSet(), entry.Name(), nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("failed to parse imports from %s: %v", entry.Name(), err)
		}

		for _, importSpec := range file.Imports {
			importPath := strings.Trim(importSpec.Path.Value, `"`)
			if importPath == forbiddenImport {
				t.Errorf("%s imports %s; use %s", entry.Name(), forbiddenImport, requiredImport)
			}
			if importPath == requiredImport {
				importsSonic = true
			}
		}
	}

	if !importsSonic {
		t.Errorf("repo does not import %s", requiredImport)
	}
}

func TestOutputPathsForJSON(t *testing.T) {
	t.Run("appends json extension to base name", func(t *testing.T) {
		paths, err := outputPaths("graph", outputFormatJSON)
		if err != nil {
			t.Fatalf("outputPaths returned error: %v", err)
		}

		if paths.JSON != "graph.json" {
			t.Fatalf("JSON path = %q, want %q", paths.JSON, "graph.json")
		}
	})

	t.Run("keeps existing json extension", func(t *testing.T) {
		paths, err := outputPaths("graph.json", outputFormatJSON)
		if err != nil {
			t.Fatalf("outputPaths returned error: %v", err)
		}

		if paths.JSON != "graph.json" {
			t.Fatalf("JSON path = %q, want %q", paths.JSON, "graph.json")
		}
	})
}

func TestOutputPathsForJSONL(t *testing.T) {
	t.Run("uses base name for node and edge jsonl files", func(t *testing.T) {
		paths, err := outputPaths("graph", outputFormatJSONL)
		if err != nil {
			t.Fatalf("outputPaths returned error: %v", err)
		}

		if paths.Nodes != "graph.nodes.jsonl" {
			t.Fatalf("nodes path = %q, want %q", paths.Nodes, "graph.nodes.jsonl")
		}
		if paths.Edges != "graph.edges.jsonl" {
			t.Fatalf("edges path = %q, want %q", paths.Edges, "graph.edges.jsonl")
		}
	})

	t.Run("keeps existing jsonl stem", func(t *testing.T) {
		paths, err := outputPaths("graph.nodes.jsonl", outputFormatJSONL)
		if err != nil {
			t.Fatalf("outputPaths returned error: %v", err)
		}

		if paths.Nodes != "graph.nodes.jsonl" {
			t.Fatalf("nodes path = %q, want %q", paths.Nodes, "graph.nodes.jsonl")
		}
		if paths.Edges != "graph.edges.jsonl" {
			t.Fatalf("edges path = %q, want %q", paths.Edges, "graph.edges.jsonl")
		}
	})
}

func TestValidateOutputFormatRejectsUnsupportedFormat(t *testing.T) {
	if err := validateOutputFormat("csv"); err == nil {
		t.Fatal("validateOutputFormat did not reject unsupported format")
	}
}

func TestWriteJSONOutputStreamsParseableOpenGraph(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "graph.json")

	err := writeJSONOutput(path, testGenerationConfig())
	if err != nil {
		t.Fatalf("writeJSONOutput returned error: %v", err)
	}

	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("failed to open JSON output: %v", err)
	}
	defer f.Close()

	var og OpenGraph
	if err := sonic.ConfigDefault.NewDecoder(f).Decode(&og); err != nil {
		t.Fatalf("failed to decode JSON output: %v", err)
	}

	if og.Metadata.SourceKind != "OggenBase" {
		t.Fatalf("source kind = %q, want %q", og.Metadata.SourceKind, "OggenBase")
	}
	if len(og.Graph.Nodes) != 4 {
		t.Fatalf("node count = %d, want 4", len(og.Graph.Nodes))
	}
	if len(og.Graph.Edges) != 10 {
		t.Fatalf("edge count = %d, want 10", len(og.Graph.Edges))
	}
	for _, node := range og.Graph.Nodes {
		if len(node.Kinds) != 1 || node.Kinds[0] != "OGGEN_NODE_1" {
			t.Errorf("default node kinds = %v, want [OGGEN_NODE_1]", node.Kinds)
		}
	}
	for _, edge := range og.Graph.Edges {
		if edge.Kind != "OGGEN_EDGE_1" {
			t.Errorf("default edge kind = %q, want OGGEN_EDGE_1", edge.Kind)
		}
	}
}

func TestWriteJSONLOutputStreamsParseableRecords(t *testing.T) {
	dir := t.TempDir()
	nodesPath := filepath.Join(dir, "graph.nodes.jsonl")
	edgesPath := filepath.Join(dir, "graph.edges.jsonl")

	err := writeJSONLOutput(nodesPath, edgesPath, testGenerationConfig())
	if err != nil {
		t.Fatalf("writeJSONLOutput returned error: %v", err)
	}

	nodes := decodeJSONLLines[Node](t, nodesPath)
	edges := decodeJSONLLines[Edge](t, edgesPath)

	if len(nodes) != 4 {
		t.Fatalf("node line count = %d, want 4", len(nodes))
	}
	if len(edges) != 10 {
		t.Fatalf("edge line count = %d, want 10", len(edges))
	}
	if nodes[0].ID != "oggen_0" {
		t.Fatalf("first node ID = %q, want %q", nodes[0].ID, "oggen_0")
	}
	for _, node := range nodes {
		if len(node.Kinds) != 1 || node.Kinds[0] != "OGGEN_NODE_1" {
			t.Errorf("default node kinds = %v, want [OGGEN_NODE_1]", node.Kinds)
		}
	}
	for _, edge := range edges {
		if edge.Kind != "OGGEN_EDGE_1" {
			t.Errorf("default edge kind = %q, want OGGEN_EDGE_1", edge.Kind)
		}
	}
}

func TestWriteJSONOutputUsesIndependentKindCounts(t *testing.T) {
	path := filepath.Join(t.TempDir(), "graph.json")
	config := testGenerationConfig()
	config.NumNodes = 100
	config.NumTiers = 1
	config.NumEdgesPerTier = 100
	config.NumNodeKinds = 10
	config.NumEdgeKinds = 2
	if err := writeJSONOutput(path, config); err != nil {
		t.Fatalf("writeJSONOutput returned error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var graph OpenGraph
	if err := sonic.Unmarshal(data, &graph); err != nil {
		t.Fatal(err)
	}

	seenNode10 := false
	for _, node := range graph.Graph.Nodes {
		if len(node.Kinds) != 1 {
			t.Fatalf("node kinds = %v, want exactly one kind", node.Kinds)
		}
		number, err := strconv.Atoi(strings.TrimPrefix(node.Kinds[0], "OGGEN_NODE_"))
		if err != nil || !strings.HasPrefix(node.Kinds[0], "OGGEN_NODE_") || number < 1 || number > 10 {
			t.Errorf("node kind %q is outside OGGEN_NODE_1 through OGGEN_NODE_10", node.Kinds[0])
		}
		seenNode10 = seenNode10 || number == 10
	}
	if !seenNode10 {
		t.Error("configured upper node kind OGGEN_NODE_10 was never generated")
	}

	seenEdge2 := false
	for _, edge := range graph.Graph.Edges {
		if edge.Kind != "OGGEN_EDGE_1" && edge.Kind != "OGGEN_EDGE_2" {
			t.Errorf("edge kind = %q, want OGGEN_EDGE_1 or OGGEN_EDGE_2", edge.Kind)
		}
		seenEdge2 = seenEdge2 || edge.Kind == "OGGEN_EDGE_2"
	}
	if !seenEdge2 {
		t.Error("configured upper edge kind OGGEN_EDGE_2 was never generated")
	}
}

func TestCLIKindCountFlagsApplyToJSONL(t *testing.T) {
	base := filepath.Join(t.TempDir(), "graph")
	cmd := exec.Command("go", "run", ".", "-n", "40", "-t", "2", "-e", "20", "-f", "jsonl", "-o", base, "-node-kinds", "2", "-edge-kinds", "3")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("oggen failed: %v\n%s", err, output)
	}

	nodes := decodeJSONLLines[Node](t, base+".nodes.jsonl")
	edges := decodeJSONLLines[Edge](t, base+".edges.jsonl")
	seenNode2 := false
	for _, node := range nodes {
		if len(node.Kinds) != 1 || (node.Kinds[0] != "OGGEN_NODE_1" && node.Kinds[0] != "OGGEN_NODE_2") {
			t.Errorf("node kinds = %v, want one of OGGEN_NODE_1 or OGGEN_NODE_2", node.Kinds)
			continue
		}
		seenNode2 = seenNode2 || node.Kinds[0] == "OGGEN_NODE_2"
	}
	if !seenNode2 {
		t.Error("OGGEN_NODE_2 was never generated")
	}
	seenEdge3 := false
	for _, edge := range edges {
		if edge.Kind != "OGGEN_EDGE_1" && edge.Kind != "OGGEN_EDGE_2" && edge.Kind != "OGGEN_EDGE_3" {
			t.Errorf("edge kind = %q, want OGGEN_EDGE_1 through OGGEN_EDGE_3", edge.Kind)
		}
		seenEdge3 = seenEdge3 || edge.Kind == "OGGEN_EDGE_3"
	}
	if !seenEdge3 {
		t.Error("OGGEN_EDGE_3 was never generated")
	}
}

func TestCLIRejectsKindCountsOutsideOneToTen(t *testing.T) {
	for _, tc := range []struct {
		flag  string
		value string
	}{
		{"-node-kinds", "0"},
		{"-node-kinds", "11"},
		{"-edge-kinds", "0"},
		{"-edge-kinds", "11"},
	} {
		t.Run(tc.flag+"="+tc.value, func(t *testing.T) {
			base := filepath.Join(t.TempDir(), "graph")
			cmd := exec.Command("go", "run", ".", "-n", "4", "-t", "2", "-e", "2", "-o", base, tc.flag, tc.value)
			output, err := cmd.CombinedOutput()
			if err == nil {
				t.Fatalf("oggen accepted %s=%s", tc.flag, tc.value)
			}
			if !strings.Contains(string(output), tc.flag) {
				t.Errorf("error %q does not identify %s", output, tc.flag)
			}
		})
	}
}

func testGenerationConfig() generationConfig {
	return generationConfig{
		NumNodes:        4,
		NumTiers:        2,
		NumEdgesPerTier: 3,
		Now: func() time.Time {
			return time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC)
		},
		Rand: rand.New(rand.NewSource(1)),
	}
}

func decodeJSONLLines[T any](t *testing.T, path string) []T {
	t.Helper()

	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("failed to open %s: %v", path, err)
	}
	defer f.Close()

	var records []T
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var record T
		if err := sonic.Unmarshal(scanner.Bytes(), &record); err != nil {
			t.Fatalf("failed to parse JSONL line %q: %v", scanner.Text(), err)
		}
		records = append(records, record)
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("failed to scan %s: %v", path, err)
	}

	return records
}
