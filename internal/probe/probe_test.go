package probe

import (
	"encoding/json"
	"testing"
)

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		sec  float64
		want string
	}{
		{0, "00:00"},
		{45.5, "00:45"},
		{125, "02:05"},
		{3665, "01:01:05"},
	}

	for _, tt := range tests {
		got := FormatDuration(tt.sec)
		if got != tt.want {
			t.Errorf("FormatDuration(%f) = %q, want %q", tt.sec, got, tt.want)
		}
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes int64
		want  string
	}{
		{500, "500 B"},
		{1024 * 1024, "1.0 MB"},
		{1536 * 1024 * 1024, "1.5 GB"},
	}

	for _, tt := range tests {
		got := FormatBytes(tt.bytes)
		if got != tt.want {
			t.Errorf("FormatBytes(%d) = %q, want %q", tt.bytes, got, tt.want)
		}
	}
}

func TestParseFPS(t *testing.T) {
	tests := []struct {
		in   string
		want float64
	}{
		{"30/1", 30.0},
		{"60/1", 60.0},
		{"0/0", 0.0},
		{"", 0.0},
		{"25", 25.0},
	}

	for _, tt := range tests {
		got := ParseFPS(tt.in)
		if got != tt.want {
			t.Errorf("ParseFPS(%q) = %f, want %f", tt.in, got, tt.want)
		}
	}
}

func TestProbeResultMethods(t *testing.T) {
	rawJSON := `{
		"format": {
			"filename": "/tmp/test_video.mkv",
			"nb_streams": 2,
			"format_name": "matroska,webm",
			"duration": "120.500000",
			"size": "52428800"
		},
		"streams": [
			{
				"index": 0,
				"codec_type": "video",
				"codec_name": "hevc",
				"width": 3840,
				"height": 2160,
				"pix_fmt": "yuv420p10le",
				"avg_frame_rate": "60/1",
				"color_transfer": "smpte2084",
				"color_primaries": "bt2020",
				"disposition": {"default": 1}
			},
			{
				"index": 1,
				"codec_type": "audio",
				"codec_name": "aac",
				"sample_rate": "48000",
				"channels": 2,
				"channel_layout": "stereo",
				"disposition": {"default": 1}
			}
		]
	}`

	var res ProbeResult
	if err := json.Unmarshal([]byte(rawJSON), &res); err != nil {
		t.Fatalf("json unmarshal failed: %v", err)
	}

	if !res.IsVideo() {
		t.Errorf("expected IsVideo() = true")
	}
	if !res.IsHDR() {
		t.Errorf("expected IsHDR() = true for smpte2084 + bt2020")
	}
	if res.HDRType() != "HDR10 (PQ)" {
		t.Errorf("got HDRType %q, want HDR10 (PQ)", res.HDRType())
	}
}
