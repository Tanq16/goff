package cmd

import (
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

func TestValidScaleTarget(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want bool
	}{
		{"named tier", "720p", true},
		{"named tier uppercase", "1080P", true},
		{"alias", "4k", true},
		{"explicit dimensions", "1280x720", true},
		{"empty", "", false},
		{"raw ffmpeg filter injection", "iw/2:ih/2", false},
		{"garbage reaching the filter graph", "garbage", false},
		{"zero width", "0x720", false},
		{"zero height", "1280x0", false},
		{"negative", "-1x720", false},
		{"missing separator", "1280720", false},
		{"colon separator", "1280:720", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validScaleTarget(tt.in); got != tt.want {
				t.Errorf("validScaleTarget(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestEnumFlagSet(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{"exact match", "webp", "webp", false},
		{"uppercase is normalized", "WEBP", "webp", false},
		{"mixed case is normalized", "WebP", "webp", false},
		{"not in set", "jpg", "", true},
		{"empty", "", "", true},
		{"prefix of a valid value", "we", "", true},
		{"valid value with whitespace", " webp", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var target string
			f := newEnum(&target, "gif", "gif", "webp")
			err := f.Set(tt.in)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Set(%q) err = %v, wantErr %v", tt.in, err, tt.wantErr)
			}
			if err != nil {
				if target != "gif" {
					t.Errorf("Set(%q) failed but still changed the target to %q", tt.in, target)
				}
				return
			}
			if target != tt.want {
				t.Errorf("Set(%q) stored %q, want %q", tt.in, target, tt.want)
			}
		})
	}
}

func TestBoundedFloatSet(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		wantErr bool
	}{
		{"inside range", "0.5", false},
		{"at inclusive max", "1", false},
		{"just inside min", "0.0001", false},
		{"at exclusive min", "0", true},
		{"below min", "-1", true},
		{"above max", "1.0001", true},
		{"well above max", "5", true},
		{"not a number", "half", true},
		{"empty", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var target float64
			f := newBoundedFloat(&target, 1.0, 0, 1.0, "greater than 0 and at most 1")
			err := f.Set(tt.in)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Set(%q) err = %v, wantErr %v", tt.in, err, tt.wantErr)
			}
			if err != nil && target != 1.0 {
				t.Errorf("Set(%q) failed but still changed the target to %v", tt.in, target)
			}
		})
	}
}
