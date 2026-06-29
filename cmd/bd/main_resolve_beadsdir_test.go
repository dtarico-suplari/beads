package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/steveyegge/beads/internal/utils"
)

// TestResolveCommandBeadsDir_HonorsExplicitBeadsDir reproduces the shared-server
// bug: an absolute BEADS_DOLT_DATA_DIR makes DatabasePath() collapse to the same
// data dir for every rig, so the data-path matcher would resolve to the wrong
// (town) .beads and bind the store to the wrong dolt_database. When BEADS_DIR is
// explicitly set to a valid rig .beads, it must win. Without the BEADS_DIR
// short-circuit this returns the dbPath's parent dir, not the rig .beads (red).
func TestResolveCommandBeadsDir_HonorsExplicitBeadsDir(t *testing.T) {
	root := t.TempDir()

	rigBeads := filepath.Join(root, "rig", ".beads")
	if err := os.MkdirAll(rigBeads, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(rigBeads, "metadata.json"),
		[]byte(`{"backend":"dolt","dolt_mode":"server","dolt_database":"suplari_assistant"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	// Shared absolute data dir whose owning tree is NOT the rig — this is what
	// dbPath collapses to, and what the old matcher would resolve incorrectly.
	sharedData := filepath.Join(root, "town", ".dolt-data")
	if err := os.MkdirAll(sharedData, 0o755); err != nil {
		t.Fatal(err)
	}

	t.Setenv("BEADS_DIR", rigBeads)

	got := resolveCommandBeadsDir(sharedData)
	want := utils.CanonicalizePath(rigBeads)
	if got != want {
		t.Fatalf("resolveCommandBeadsDir(%q) with BEADS_DIR=%q = %q, want %q",
			sharedData, rigBeads, got, want)
	}
}

// TestResolveCommandBeadsDir_IgnoresInvalidBeadsDir ensures the short-circuit only
// fires for a real .beads (with metadata.json); a stale/invalid BEADS_DIR must not
// hijack resolution, preserving standalone/non-rig behavior.
func TestResolveCommandBeadsDir_IgnoresInvalidBeadsDir(t *testing.T) {
	root := t.TempDir()
	bogus := filepath.Join(root, "does-not-exist", ".beads")
	t.Setenv("BEADS_DIR", bogus)

	dbParent := filepath.Join(root, "proj")
	dbPath := filepath.Join(dbParent, ".beads", "dolt")
	if err := os.MkdirAll(filepath.Join(dbParent, ".beads"), 0o755); err != nil {
		t.Fatal(err)
	}

	got := resolveCommandBeadsDir(dbPath)
	if got == utils.CanonicalizePath(bogus) {
		t.Fatalf("invalid BEADS_DIR %q should not be returned, got %q", bogus, got)
	}
}
