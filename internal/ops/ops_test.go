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
	fakeVideoProbe := &probe.ProbeResult{
		Streams: []probe.StreamInfo{
			{CodecType: "video", Width: 1920, Height: 1080},
			{CodecType: "audio", SampleRate: "48000", Channels: 2},
		},
	}

	tests := []struct {
		name      string
		input     string
		probe     *probe.ProbeResult
		opts      AudioLoudnormOpts
		wantExt   string
		hasVN     bool
		hasCopy   bool
	}{
		{
			name:    "pure audio to mp3",
			input:   "song.wav",
			probe:   nil,
			opts:    AudioLoudnormOpts{IntegratedLoudness: -14.0, OutputExt: "mp3"},
			wantExt: "mp3",
			hasVN:   false,
			hasCopy: false,
		},
		{
			name:    "video to mp3 audio mastering",
			input:   "video.webm",
			probe:   fakeVideoProbe,
			opts:    AudioLoudnormOpts{IntegratedLoudness: -16.0, OutputExt: "mp3"},
			wantExt: "mp3",
			hasVN:   true,
			hasCopy: false,
		},
		{
			name:    "video normalization preserving video",
			input:   "video.webm",
			probe:   fakeVideoProbe,
			opts:    AudioLoudnormOpts{IntegratedLoudness: -16.0},
			wantExt: "mp4",
			hasVN:   false,
			hasCopy: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := BuildAudioLoudnorm(tt.input, tt.probe, tt.opts)
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if res.TargetExt != tt.wantExt {
				t.Errorf("got target ext %q, want %q", res.TargetExt, tt.wantExt)
			}
			if slices.Contains(res.Args, "-vn") != tt.hasVN {
				t.Errorf("got -vn in args %v, want %v", slices.Contains(res.Args, "-vn"), tt.hasVN)
			}
			if slices.Contains(res.Args, "copy") != tt.hasCopy {
				t.Errorf("got copy in args %v, want %v", slices.Contains(res.Args, "copy"), tt.hasCopy)
			}
		})
	}
}

func TestBuildAudioConvert(t *testing.T) {
	fakeVideoProbe := &probe.ProbeResult{
		Streams: []probe.StreamInfo{
			{CodecType: "video", Width: 1920, Height: 1080},
			{CodecType: "audio", SampleRate: "48000", Channels: 2},
		},
	}

	res, err := BuildAudioConvert("video.webm", fakeVideoProbe, AudioConvertOpts{
		Format:  "flac",
		Bitrate: "160k",
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if res.TargetExt != "flac" {
		t.Errorf("got target ext %q, want flac", res.TargetExt)
	}
	if !slices.Contains(res.Args, "-vn") {
		t.Errorf("expected -vn for video input in args: %v", res.Args)
	}
}

func TestBuildVideoTrim(t *testing.T) {
	tests := []struct {
		name     string
		opts     VideoTrimOpts
		wantCopy bool
		wantCRF  bool
	}{
		{
			name:     "fast copy trim",
			opts:     VideoTrimOpts{Start: "00:00:10", Duration: "5", Accurate: false},
			wantCopy: true,
			wantCRF:  false,
		},
		{
			name:     "accurate re-encode trim",
			opts:     VideoTrimOpts{Start: "00:00:10", End: "00:00:15", Accurate: true},
			wantCopy: false,
			wantCRF:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := BuildVideoTrim("video.mp4", nil, tt.opts)
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if slices.Contains(res.Args, "copy") != tt.wantCopy {
				t.Errorf("got copy in args %v, want %v", slices.Contains(res.Args, "copy"), tt.wantCopy)
			}
			if slices.Contains(res.Args, "-crf") != tt.wantCRF {
				t.Errorf("got -crf in args %v, want %v", slices.Contains(res.Args, "-crf"), tt.wantCRF)
			}
		})
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

func TestBuildMultiMixFilterGraph(t *testing.T) {
	tests := []struct {
		name        string
		sources     []AudioSource
		fit         string
		wantFilters []string
		absent      []string
	}{
		{
			name:        "plain sources need no per-input chain",
			sources:     []AudioSource{{Path: "a.mp3", Volume: 1.0}, {Path: "b.mp3", Volume: 1.0}},
			wantFilters: []string{"[0:a][1:a]amix=inputs=2:duration=longest:normalize=0[aout]"},
			absent:      []string{"adelay", "volume="},
		},
		{
			name:        "offset becomes an adelay chain",
			sources:     []AudioSource{{Path: "a.mp3", Volume: 1.0}, {Path: "b.mp3", DelayMS: 5000, Volume: 1.0}},
			wantFilters: []string{"[1:a]adelay=5000:all=1[mix1]", "[0:a][mix1]amix=inputs=2"},
		},
		{
			name:        "volume becomes a volume chain",
			sources:     []AudioSource{{Path: "a.mp3", Volume: 1.0}, {Path: "b.mp3", Volume: 0.3}},
			wantFilters: []string{"[1:a]volume=0.3[mix1]"},
		},
		{
			name:        "offset and volume share one chain",
			sources:     []AudioSource{{Path: "a.mp3", Volume: 1.0}, {Path: "b.mp3", DelayMS: 2500, Volume: 0.5}},
			wantFilters: []string{"[1:a]adelay=2500:all=1,volume=0.5[mix1]"},
		},
		{
			name:        "fit shortest reaches amix",
			sources:     []AudioSource{{Path: "a.mp3", Volume: 1.0}, {Path: "b.mp3", Volume: 1.0}},
			fit:         "shortest",
			wantFilters: []string{"duration=shortest"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := BuildMultiMix(MultiMixOpts{Sources: tt.sources, Fit: tt.fit})
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			idx := slices.Index(res.Args, "-filter_complex")
			if idx < 0 || idx+1 >= len(res.Args) {
				t.Fatalf("no -filter_complex in args: %v", res.Args)
			}
			graph := res.Args[idx+1]
			for _, want := range tt.wantFilters {
				if !strings.Contains(graph, want) {
					t.Errorf("graph %q missing %q", graph, want)
				}
			}
			for _, gone := range tt.absent {
				if strings.Contains(graph, gone) {
					t.Errorf("graph %q should not contain %q", graph, gone)
				}
			}
		})
	}
}

func TestBuildMultiMixRejectsSingleInput(t *testing.T) {
	if _, err := BuildMultiMix(MultiMixOpts{Sources: []AudioSource{{Path: "a.mp3"}}}); err == nil {
		t.Error("expected an error for a single input")
	}
}

func TestBuildMultiMuxAudioModes(t *testing.T) {
	sources := []AudioSource{{Path: "music.mp3", Volume: 1.0}}

	tests := []struct {
		name    string
		mode    string
		hasAudi bool
		want    []string
		absent  []string
	}{
		{
			name:    "mix folds the original into one track",
			mode:    MuxModeMix,
			hasAudi: true,
			want:    []string{"[0:a][1:a]amix=inputs=2"},
		},
		{
			name:    "mix on a silent video uses the source alone",
			mode:    MuxModeMix,
			hasAudi: false,
			want:    []string{"[1:a]amix=inputs=1"},
		},
		{
			name:    "replace drops the original",
			mode:    MuxModeReplace,
			hasAudi: true,
			want:    []string{"[1:a]amix=inputs=1"},
			absent:  []string{"[0:a]"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := BuildMultiMuxAudio(MultiMuxAudioOpts{
				VideoInput:    "in.mp4",
				Sources:       sources,
				Mode:          tt.mode,
				Fit:           "video",
				VideoHasAudio: tt.hasAudi,
			})
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			graph := res.Args[slices.Index(res.Args, "-filter_complex")+1]
			for _, want := range tt.want {
				if !strings.Contains(graph, want) {
					t.Errorf("graph %q missing %q", graph, want)
				}
			}
			for _, gone := range tt.absent {
				if strings.Contains(graph, gone) {
					t.Errorf("graph %q should not contain %q", graph, gone)
				}
			}
		})
	}

	t.Run("separate keeps every track selectable", func(t *testing.T) {
		res, err := BuildMultiMuxAudio(MultiMuxAudioOpts{
			VideoInput:    "in.mp4",
			Sources:       sources,
			Mode:          MuxModeSeparate,
			Fit:           "video",
			VideoHasAudio: true,
		})
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		if slices.Contains(res.Args, "-filter_complex") {
			t.Errorf("separate should not mix: %v", res.Args)
		}
		if !slices.Contains(res.Args, "0:a") || !slices.Contains(res.Args, "1:a:0") {
			t.Errorf("expected both audio streams mapped: %v", res.Args)
		}
	})

	t.Run("fit longest does not truncate to the video", func(t *testing.T) {
		res, err := BuildMultiMuxAudio(MultiMuxAudioOpts{
			VideoInput: "in.mp4", Sources: sources, Mode: MuxModeMix, Fit: "longest", VideoHasAudio: true,
		})
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		if slices.Contains(res.Args, "-shortest") {
			t.Errorf("fit longest should not pass -shortest: %v", res.Args)
		}
	})
}

func TestBuildVideoTransformKeepsFiltersAlongsideSpeed(t *testing.T) {
	withAudio := &probe.ProbeResult{
		Streams: []probe.StreamInfo{
			{CodecType: "video", Width: 1920, Height: 1080},
			{CodecType: "audio"},
		},
	}

	res, err := BuildVideoTransform("in.mp4", withAudio, VideoTransformOpts{
		Rotate: "90",
		Scale:  "720p",
		Speed:  2.0,
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	idx := slices.Index(res.Args, "-filter_complex")
	if idx < 0 {
		t.Fatalf("expected a filter_complex when speed retimes audio: %v", res.Args)
	}
	graph := res.Args[idx+1]
	for _, want := range []string{"transpose=1", "scale=", "setpts=", "atempo="} {
		if !strings.Contains(graph, want) {
			t.Errorf("graph %q dropped %q", graph, want)
		}
	}
}

func TestBuildVideoTransformSpeedWithoutAudioStaysSimple(t *testing.T) {
	silent := &probe.ProbeResult{
		Streams: []probe.StreamInfo{{CodecType: "video", Width: 1920, Height: 1080}},
	}

	res, err := BuildVideoTransform("in.mp4", silent, VideoTransformOpts{Scale: "720p", Speed: 2.0})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if slices.Contains(res.Args, "-filter_complex") {
		t.Errorf("no audio to retime, expected -vf: %v", res.Args)
	}
	vf := res.Args[slices.Index(res.Args, "-vf")+1]
	for _, want := range []string{"scale=", "setpts="} {
		if !strings.Contains(vf, want) {
			t.Errorf("vf chain %q dropped %q", vf, want)
		}
	}
}
