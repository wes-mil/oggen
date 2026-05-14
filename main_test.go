package main

import (
	"bufio"
	"encoding/json"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"
)

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
	if err := json.NewDecoder(f).Decode(&og); err != nil {
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
		if err := json.Unmarshal(scanner.Bytes(), &record); err != nil {
			t.Fatalf("failed to parse JSONL line %q: %v", scanner.Text(), err)
		}
		records = append(records, record)
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("failed to scan %s: %v", path, err)
	}

	return records
}
