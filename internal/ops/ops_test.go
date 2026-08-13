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
