package cmd

import (
	"math"
	"os"
	"path/filepath"
	"testing"
)

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
		want    float64
		wantErr bool
	}{
		{"inside range", "0.5", 0.5, false},
		{"at inclusive max", "1", 1.0, false},
		{"at inclusive min", "0.01", 0.01, false},
		{"just below min clamps up", "0.009", 0.01, false},
		{"zero clamps up", "0", 0.01, false},
		{"negative clamps up", "-1", 0.01, false},
		{"just above max clamps down", "1.0001", 1.0, false},
		{"well above max clamps down", "5", 1.0, false},
		{"infinity clamps down", "Inf", 1.0, false},
		{"not a number", "half", 1.0, true},
		{"nan", "NaN", 1.0, true},
		{"empty", "", 1.0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var target float64
			f := newBoundedFloat(&target, 1.0, 0.01, 1.0)
			err := f.Set(tt.in)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Set(%q) err = %v, wantErr %v", tt.in, err, tt.wantErr)
			}
			if target != tt.want {
				t.Errorf("Set(%q) stored %v, want %v", tt.in, target, tt.want)
			}
		})
	}
}

func TestBoundedIntSet(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    int
		wantErr bool
	}{
		{"inside range", "30", 30, false},
		{"at inclusive min", "1", 1, false},
		{"at inclusive max", "63", 63, false},
		{"zero clamps up", "0", 1, false},
		{"negative clamps up", "-5", 1, false},
		{"above max clamps down", "200", 63, false},
		{"fractional", "28.5", 21, true},
		{"not a number", "high", 21, true},
		{"empty", "", 21, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var target int
			f := newBoundedInt(&target, 21, 1, 63)
			err := f.Set(tt.in)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Set(%q) err = %v, wantErr %v", tt.in, err, tt.wantErr)
			}
			if target != tt.want {
				t.Errorf("Set(%q) stored %d, want %d", tt.in, target, tt.want)
			}
		})
	}
}

func TestParseTimestamp(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    float64
		wantErr bool
	}{
		{"bare seconds", "90", 90, false},
		{"fractional seconds", "90.5", 90.5, false},
		{"minutes and seconds", "1:30", 90, false},
		{"hours minutes seconds", "00:01:30", 90, false},
		{"fractional last component", "00:01:30.5", 90.5, false},
		{"zero", "0", 0, false},
		{"fractional minutes", "1.5:30", 0, true},
		{"four components", "1:2:3:4", 0, true},
		{"negative", "-5", 0, true},
		{"empty component", "1::30", 0, true},
		{"not a number", "banana", 0, true},
		{"empty", "", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTimestamp(tt.in)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseTimestamp(%q) err = %v, wantErr %v", tt.in, err, tt.wantErr)
			}
			if err == nil && got != tt.want {
				t.Errorf("parseTimestamp(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestParseMediaInput(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    mediaInput
		wantErr bool
	}{
		{"path only", "music.mp3", mediaInput{Path: "music.mp3", Volume: 1.0}, false},
		{"offset in seconds", "music.mp3:at=5", mediaInput{Path: "music.mp3", DelayMS: 5000, Volume: 1.0}, false},
		{"offset with colons", "music.mp3:at=00:01:30", mediaInput{Path: "music.mp3", DelayMS: 90000, Volume: 1.0}, false},
		{"volume only", "music.mp3:vol=0.3", mediaInput{Path: "music.mp3", Volume: 0.3}, false},
		{"both options", "music.mp3:at=00:00:05:vol=0.3", mediaInput{Path: "music.mp3", DelayMS: 5000, Volume: 0.3}, false},
		{"options in either order", "music.mp3:vol=0.3:at=5", mediaInput{Path: "music.mp3", DelayMS: 5000, Volume: 0.3}, false},
		{"colon in path without options", "odd:name.mp3", mediaInput{Path: "odd:name.mp3", Volume: 1.0}, false},
		{"colon in path with options", "odd:name.mp3:at=5", mediaInput{Path: "odd:name.mp3", DelayMS: 5000, Volume: 1.0}, false},
		{"fractional offset", "music.mp3:at=1.5", mediaInput{Path: "music.mp3", DelayMS: 1500, Volume: 1.0}, false},
		{"malformed offset", "music.mp3:at=banana", mediaInput{}, true},
		{"zero volume", "music.mp3:vol=0", mediaInput{}, true},
		{"negative volume", "music.mp3:vol=-1", mediaInput{}, true},
		{"volume above cap", "music.mp3:vol=11", mediaInput{}, true},
		{"no path before options", ":at=5", mediaInput{}, true},
		{"empty", "", mediaInput{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseMediaInput(tt.in)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseMediaInput(%q) err = %v, wantErr %v", tt.in, err, tt.wantErr)
			}
			if err == nil && got != tt.want {
				t.Errorf("parseMediaInput(%q) = %+v, want %+v", tt.in, got, tt.want)
			}
		})
	}
}
