package presets

import (
	"testing"

	"github.com/Tanq16/goff/internal/probe"
)

func TestBuiltinPresets(t *testing.T) {
	fakeProbe := &probe.ProbeResult{
		Format: probe.FormatInfo{
			DurationStr: "120.0",
		},
		Streams: []probe.StreamInfo{
			{
				CodecType: "video",
				Width:     1920,
				Height:    1080,
			},
			{
				CodecType: "audio",
				Channels:  2,
			},
		},
	}

	all := List()
	if len(all) == 0 {
		t.Fatalf("expected registered presets, got 0")
	}

	expectedIDs := []string{
		"web-optimize",
		"discord-25mb",
		"discord-10mb",
		"fast-remux",
		"animated-gif",
		"extract-mp3-320",
		"strip-audio",
		"vertical-9-16",
		"podcast-master",
		"normalize-audio",
	}

	for _, id := range expectedIDs {
		p, ok := Get(id)
		if !ok {
			t.Errorf("expected preset %q to be registered", id)
			continue
		}
		res, err := p.Build("test_video.mp4", fakeProbe)
		if err != nil {
			t.Errorf("failed to build preset %q: %v", id, err)
			continue
		}
		if len(res.Args) == 0 {
			t.Errorf("preset %q returned empty args", id)
		}
		if res.TargetExt == "" {
			t.Errorf("preset %q returned empty target ext", id)
		}
	}
}
