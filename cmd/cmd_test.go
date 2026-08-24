package cmd

import (
	"bytes"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

func findCommand(name string) *cobra.Command {
	for _, c := range rootCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func TestRootHelp(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"--help"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected help error: %v", err)
	}
	if buf.String() == "" {
		t.Errorf("expected help output, got empty")
	}
}

func TestVerbsRegistered(t *testing.T) {
	verbs := []string{
		"inspect", "compress", "remux", "extract", "convert", "normalize",
		"trim", "gif", "hls", "rotate", "scale", "speed", "crop", "mute",
		"watermark", "concat", "mux",
	}
	registered := make(map[string]bool)
	for _, c := range rootCmd.Commands() {
		registered[c.Name()] = true
	}
	for _, v := range verbs {
		if !registered[v] {
			t.Errorf("verb %q not registered on rootCmd", v)
		}
	}
	if registered["batch"] {
		t.Errorf("batch command should have been removed")
	}
}

func TestSharedFlagsArePersistent(t *testing.T) {
	for _, name := range []string{"output", "yes", "jobs", "debug", "for-ai"} {
		if rootCmd.PersistentFlags().Lookup(name) == nil {
			t.Errorf("--%s must be persistent so every verb inherits it", name)
		}
	}
}

func TestTransformVerbsAcceptSiblingFlags(t *testing.T) {
	tests := []struct {
		verb     string
		own      string
		siblings []string
	}{
		{"rotate", "by", []string{"scale", "speed", "crop", "mute"}},
		{"scale", "by", []string{"rotate", "speed", "crop", "mute"}},
		{"speed", "by", []string{"rotate", "scale", "crop", "mute"}},
		{"crop", "", []string{"rotate", "scale", "speed", "mute"}},
		{"mute", "", []string{"rotate", "scale", "speed", "crop"}},
	}
	for _, tt := range tests {
		t.Run(tt.verb, func(t *testing.T) {
			cmd := findCommand(tt.verb)
			if cmd == nil {
				t.Fatalf("verb %q not registered", tt.verb)
			}
			if tt.own != "" && cmd.Flags().Lookup(tt.own) == nil {
				t.Errorf("%s missing its own --%s flag", tt.verb, tt.own)
			}
			if tt.own == "" && cmd.Flags().Lookup("by") != nil {
				t.Errorf("%s takes no parameter and must not register --by", tt.verb)
			}
			for _, s := range tt.siblings {
				if cmd.Flags().Lookup(s) == nil {
					t.Errorf("%s cannot compose --%s in one encode", tt.verb, s)
				}
			}
		})
	}
}

func TestParseSizeBudget(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    float64
		wantErr bool
	}{
		{"discord 25MB keeps half a MB back", "25MB", 24.5, false},
		{"discord 10MB keeps half a MB back", "10MB", 9.5, false},
		{"lowercase unit", "25mb", 24.5, false},
		{"short unit", "25m", 24.5, false},
		{"bare number", "25", 24.5, false},
		{"surrounding space", "  25 MB  ", 24.5, false},
		{"large budget scales headroom by percent", "100MB", 98.0, false},
		{"fractional budget", "2.5MB", 2.0, false},
		{"empty", "", 0, true},
		{"unit only", "MB", 0, true},
		{"not a number", "big", 0, true},
		{"zero", "0MB", 0, true},
		{"negative", "-5MB", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseSizeBudget(tt.in)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseSizeBudget(%q) err = %v, wantErr %v", tt.in, err, tt.wantErr)
			}
			if err != nil {
				return
			}
			if math.Abs(got-tt.want) > 1e-9 {
				t.Errorf("parseSizeBudget(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestClaimOutputAvoidsCollisions(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"clip.mp4", "clip.mkv", "clip.webm"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0o644); err != nil {
			t.Fatalf("seeding %s: %v", name, err)
		}
	}

	claimed.Lock()
	claimed.paths = nil
	claimed.Unlock()
	rootFlags.overwrite = false
	rootFlags.output = ""

	seen := make(map[string]bool)
	for _, name := range []string{"clip.mp4", "clip.mkv", "clip.webm"} {
		out, err := claimOutput(filepath.Join(dir, name), "audio", "mp3")
		if err != nil {
			t.Fatalf("claimOutput(%s) unexpected error: %v", name, err)
		}
		if seen[out] {
			t.Fatalf("claimOutput(%s) reused %q, which would overwrite an earlier output", name, out)
		}
		seen[out] = true
	}
	if len(seen) != 3 {
		t.Errorf("expected 3 distinct outputs, got %d", len(seen))
	}
}

func TestClaimOutputHonorsOverwriteOptIn(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "clip.mp4")
	if err := os.WriteFile(input, []byte("x"), 0o644); err != nil {
		t.Fatalf("seeding input: %v", err)
	}

	claimed.Lock()
	claimed.paths = nil
	claimed.Unlock()
	rootFlags.overwrite = true
	rootFlags.output = ""
	defer func() { rootFlags.overwrite = false }()

	first, err := claimOutput(input, "audio", "mp3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	second, err := claimOutput(input, "audio", "mp3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if first != second {
		t.Errorf("with -y the same input must resolve to the same path, got %q then %q", first, second)
	}
}
