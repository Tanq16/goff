package ops

import (
	"cmp"
	"fmt"
	"strconv"
	"strings"

	"github.com/Tanq16/goff/internal/probe"
)

const AudioSyncFilter = "aresample=async=1:first_pts=0"

func ParseSizeBudget(s string) (float64, error) {
	trimmed := strings.TrimSpace(strings.ToLower(s))
	trimmed = strings.TrimSuffix(trimmed, "b")
	trimmed = strings.TrimSuffix(trimmed, "m")
	trimmed = strings.TrimSpace(trimmed)
	if trimmed == "" {
		return 0, fmt.Errorf("must look like 25MB")
	}
	limit, err := strconv.ParseFloat(trimmed, 64)
	if err != nil {
		return 0, fmt.Errorf("must look like 25MB")
	}
	if limit <= 0 {
		return 0, fmt.Errorf("must be greater than zero")
	}
	return limit - max(limit*0.02, 0.5), nil
}

type VideoOptimizeOpts struct {
	Codec         string
	CRF           int
	Lossless      bool
	Preset        string
	MaxHeight     int
	TargetSizeMB  float64
	AudioBitrate  string
	AudioRate     int
	CustomSuffix  string
	TargetExt     string
	Keep10Bit     bool
	KeepHiFiAudio bool
	CopyVideo     bool
	Subs          string
}

func BuildVideoOptimize(inputPath string, p *probe.ProbeResult, opts VideoOptimizeOpts) (*OpResult, error) {
	keepDepth := opts.Keep10Bit || opts.Lossless
	keepChannels := opts.KeepHiFiAudio || opts.Lossless

	args := []string{"-i", inputPath, "-map", "0:v:0"}
	outputIsHEVC := false

	if opts.CopyVideo {
		if v := p.PrimaryVideoStream(); v != nil && strings.EqualFold(v.CodecName, "hevc") {
			outputIsHEVC = true
		}
		args = append(args, "-c:v", "copy")
	} else {
		codec := cmp.Or(strings.ToLower(opts.Codec), "hevc")

		crf := opts.CRF
		preset := opts.Preset
		var videoEncoder string
		var losslessArgs []string
		crfMax := 51

		switch codec {
		case "av1", "libsvtav1", "svtav1":
			videoEncoder = "libsvtav1"
			losslessArgs = []string{"-svtav1-params", "lossless=1"}
			crfMax = 63
			if crf == 0 {
				crf = 32
			}
			if preset == "" {
				preset = "6"
			}
		case "h264", "libx264", "x264":
			videoEncoder = "libx264"
			losslessArgs = []string{"-crf", "0"}
			if crf == 0 {
				crf = 23
			}
			if preset == "" {
				preset = "medium"
			}
		default:
			videoEncoder = "libx265"
			losslessArgs = []string{"-x265-params", "lossless=1"}
			outputIsHEVC = true
			if crf == 0 {
				crf = 30
			}
			if preset == "" {
				preset = "medium"
			}
		}
		crf = min(crf, crfMax)

		if !opts.Lossless {
			if !keepDepth && p != nil && p.IsHDR() {
				args = append(args, "-vf", probe.ToneMapFilter())
			} else {
				maxH := opts.MaxHeight
				if maxH == 0 {
					maxH = 1080
				}
				args = append(args, "-vf", fmt.Sprintf("scale='min(1920,iw)':'min(%d,ih)':force_original_aspect_ratio=decrease,scale=trunc(iw/2)*2:trunc(ih/2)*2", maxH))
			}
		}

		args = append(args, "-c:v", videoEncoder)

		switch {
		case opts.Lossless:
			args = append(args, losslessArgs...)
			if preset != "" {
				args = append(args, "-preset", preset)
			}
		case opts.TargetSizeMB > 0 && p != nil && p.TotalDuration() > 0:
			dur := p.TotalDuration()
			audioKbps := 128.0
			totalBitrateKbps := (opts.TargetSizeMB * 8192.0) / dur
			videoBitrateKbps := max(totalBitrateKbps-audioKbps, 64)
			args = append(args,
				"-b:v", fmt.Sprintf("%dk", int(videoBitrateKbps)),
				"-maxrate", fmt.Sprintf("%dk", int(videoBitrateKbps*1.5)),
				"-bufsize", fmt.Sprintf("%dk", int(videoBitrateKbps*2)),
			)
		default:
			args = append(args, "-crf", strconv.Itoa(crf))
			if preset != "" {
				args = append(args, "-preset", preset)
			}
		}

		if !keepDepth {
			args = append(args, "-pix_fmt", "yuv420p")
		}
		args = append(args, "-fps_mode", "cfr")
	}

	if outputIsHEVC {
		args = append(args, "-tag:v", "hvc1")
	}

	if p != nil && len(p.AudioStreams()) == 0 {
		args = append(args, "-an")
	} else {
		args = append(args,
			"-map", "0:a:0",
			"-c:a", "aac",
			"-b:a", cmp.Or(opts.AudioBitrate, "128k"),
			"-ar", strconv.Itoa(cmp.Or(opts.AudioRate, 48000)),
			"-af", AudioSyncFilter,
		)
		if !keepChannels {
			args = append(args, "-ac", "2")
		}
	}

	if opts.Subs == "all" && p != nil {
		subCount := len(p.SubtitleStreams())
		for i := range subCount {
			args = append(args, "-map", fmt.Sprintf("0:s:%d", i))
		}
		if subCount > 0 {
			args = append(args, "-c:s", "mov_text")
		}
	}

	args = append(args, "-movflags", "+faststart")

	suffix := "optimized"
	switch {
	case opts.CopyVideo:
		suffix = "repaired"
	case opts.Lossless:
		suffix = "lossless"
	}
	if opts.CustomSuffix != "" {
		suffix = opts.CustomSuffix
	}

	return &OpResult{
		Args:      args,
		Suffix:    suffix,
		TargetExt: cmp.Or(opts.TargetExt, "mp4"),
	}, nil
}
