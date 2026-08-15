package ops

import (
	"slices"
	"strings"
	"testing"

	"github.com/Tanq16/goff/internal/probe"
)

func TestBuildVideoOptimize(t *testing.T) {
	fakeProbe := &probe.ProbeResult{
		Format: probe.FormatInfo{
			DurationStr: "60.0",
		},
		Streams: []probe.StreamInfo{
			{
				CodecType: "video",
				Width:     3840,
				Height:    2160,
			},
		},
	}

	res, err := BuildVideoOptimize("input.mp4", fakeProbe, VideoOptimizeOpts{
		Codec: "hevc",
		CRF:   28,
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !slices.Contains(res.Args, "libx265") {
		t.Errorf("expected libx265 in args: %v", res.Args)
	}
	if !slices.Contains(res.Args, "28") {
		t.Errorf("expected crf 28 in args: %v", res.Args)
	}
	if res.TargetExt != "mp4" {
		t.Errorf("got target ext %q, want mp4", res.TargetExt)
	}
}

func TestBuildVideoGIF(t *testing.T) {
	res, err := BuildVideoGIF("input.mp4", nil, VideoGIFOpts{
		Width:  320,
		FPS:    12,
		Format: "gif",
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if res.TargetExt != "gif" {
		t.Errorf("got target ext %q, want gif", res.TargetExt)
	}
	hasPalette := false
	for _, a := range res.Args {
		if strings.Contains(a, "palettegen") {
			hasPalette = true
			break
		}
	}
	if !hasPalette {
		t.Errorf("missing palettegen filter in GIF args: %v", res.Args)
	}
}

func TestBuildVideoExtract(t *testing.T) {
	res, err := BuildVideoExtract("input.mkv", nil, VideoExtractOpts{
		Format:  "mp3",
		Bitrate: "320k",
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if res.TargetExt != "mp3" {
		t.Errorf("got target ext %q, want mp3", res.TargetExt)
	}
	if !slices.Contains(res.Args, "-vn") {
		t.Errorf("missing -vn in extract args: %v", res.Args)
	}
}

func TestBuildAudioLoudnorm(t *testing.T) {
	res, err := BuildAudioLoudnorm("song.wav", nil, AudioLoudnormOpts{
		IntegratedLoudness: -14.0,
		OutputExt:          "mp3",
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	hasLoudnorm := false
	for _, a := range res.Args {
		if strings.Contains(a, "loudnorm=I=-14.0") {
			hasLoudnorm = true
			break
		}
	}
	if !hasLoudnorm {
		t.Errorf("missing loudnorm filter in args: %v", res.Args)
	}
}

func TestBuildMultiWatermark(t *testing.T) {
	tests := []struct {
		name      string
		opts      MultiWatermarkOpts
		wantErr   bool
		checkText string
	}{
		{
			name: "top-right standard",
			opts: MultiWatermarkOpts{
				VideoInput:     "vid.mp4",
				WatermarkInput: "logo.png",
				Position:       PosTopRight,
				ScalePercent:   15,
			},
			wantErr:   false,
			checkText: "scale2ref",
		},
		{
			name: "center with opacity",
			opts: MultiWatermarkOpts{
				VideoInput:     "vid.mp4",
				WatermarkInput: "logo.png",
				Position:       PosCenter,
				ScalePercent:   25,
				Opacity:        0.5,
			},
			wantErr:   false,
			checkText: "colorchannelmixer=aa=0.50",
		},
		{
			name: "missing inputs",
			opts: MultiWatermarkOpts{
				VideoInput: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := BuildMultiWatermark(tt.opts)
			if (err != nil) != tt.wantErr {
				t.Fatalf("BuildMultiWatermark() err = %v, wantErr = %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if res.TargetExt != "mp4" {
					t.Errorf("got target ext %q, want mp4", res.TargetExt)
				}
				matched := false
				for _, a := range res.Args {
					if strings.Contains(a, tt.checkText) {
						matched = true
						break
					}
				}
				if !matched {
					t.Errorf("expected %q in args: %v", tt.checkText, res.Args)
				}
			}
		})
	}
}

func TestBuildVideoHLS(t *testing.T) {
	tests := []struct {
		name      string
		format    HLSFormat
		wantSeg   string
		wantInit  bool
	}{
		{"fmp4 packaging", HLSFormatFMP4, "fmp4", true},
		{"mpegts packaging", HLSFormatMPEGTS, "mpegts", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := BuildVideoHLS("input.mp4", nil, VideoHLSOpts{
				Format: tt.format,
			})
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if res.TargetExt != "m3u8" {
				t.Errorf("got target ext %q, want m3u8", res.TargetExt)
			}
			if !slices.Contains(res.Args, tt.wantSeg) {
				t.Errorf("expected %q in args: %v", tt.wantSeg, res.Args)
			}
			hasInit := slices.Contains(res.Args, "init.mp4")
			if hasInit != tt.wantInit {
				t.Errorf("got init.mp4 presence %v, want %v", hasInit, tt.wantInit)
			}
		})
	}
}

func TestBuildVideoTransform_Rotation(t *testing.T) {
	tests := []struct {
		name       string
		rot        string
		wantFilter string
		wantSuffix string
	}{
		{"90 cw", "90_cw", "transpose=1", "rot90"},
		{"90 ccw", "90_ccw", "transpose=2", "rot270"},
		{"180", "180", "transpose=1,transpose=1", "rot180"},
		{"hflip", "hflip", "hflip", "hflip"},
		{"vflip", "vflip", "vflip", "vflip"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := BuildVideoTransform("input.mp4", nil, VideoTransformOpts{
				Rotate: tt.rot,
			})
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if res.Suffix != tt.wantSuffix {
				t.Errorf("got suffix %q, want %q", res.Suffix, tt.wantSuffix)
			}
			matched := false
			for _, a := range res.Args {
				if strings.Contains(a, tt.wantFilter) {
					matched = true
					break
				}
			}
			if !matched {
				t.Errorf("expected filter %q in args: %v", tt.wantFilter, res.Args)
			}
			if !slices.Contains(res.Args, "rotate=0") {
				t.Errorf("expected rotate=0 metadata in args: %v", res.Args)
			}
		})
	}
}

