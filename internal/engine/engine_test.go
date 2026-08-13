package engine

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScanProgress(t *testing.T) {
	raw := `out_time_us=50000000
total_size=10485760
speed=2.50x
fps=60.0
progress=continue
out_time_us=100000000
total_size=20971520
speed=2.55x
fps=60.0
progress=end
`
	var updates []ProgressUpdate
	err := ScanProgress(strings.NewReader(raw), 100.0, func(p ProgressUpdate) {
		updates = append(updates, p)
	})
	if err != nil {
		t.Fatalf("unexpected err from ScanProgress: %v", err)
	}

	if len(updates) != 2 {
		t.Fatalf("expected 2 updates, got %d", len(updates))
	}

	u1 := updates[0]
	if u1.Percent != 50 || u1.Speed != "2.50x" || u1.FPS != 60.0 || u1.OutBytes != 10485760 {
		t.Errorf("u1 mismatch: %+v", u1)
	}

	u2 := updates[1]
	if !u2.Done || u2.Percent != 100 || u2.Speed != "2.55x" {
		t.Errorf("u2 mismatch: %+v", u2)
	}
}

func TestResolveOutputName(t *testing.T) {
	tempDir := t.TempDir()
	src := filepath.Join(tempDir, "video.mp4")
	if err := os.WriteFile(src, []byte("fake"), 0644); err != nil {
		t.Fatal(err)
	}

	out1, err := ResolveOutputName(src, "optimized", "mp4", "", false)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	expected1 := filepath.Join(tempDir, "video.optimized.mp4")
	if out1 != expected1 {
		t.Errorf("got %q, want %q", out1, expected1)
	}

	if err := os.WriteFile(out1, []byte("fake2"), 0644); err != nil {
		t.Fatal(err)
	}

	out2, err := ResolveOutputName(src, "optimized", "mp4", "", false)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	expected2 := filepath.Join(tempDir, "video.optimized.1.mp4")
	if out2 != expected2 {
		t.Errorf("got %q, want %q", out2, expected2)
	}
}
