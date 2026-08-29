package ops

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Tanq16/goff/internal/probe"
)

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
	Codec        string
	CRF          int
	Lossless     bool
	Preset       string
	MaxHeight    int
	TargetSizeMB float64
	AudioBitrate string
	CustomSuffix string
	TargetExt    string
	Compat       bool
	Subs         string
}

func BuildVideoOptimize(inputPath string, p *probe.ProbeResult, opts VideoOptimizeOpts) (*OpResult, error) {
	codec := strings.ToLower(opts.Codec)
	if codec == "" {
		codec = "hevc"
	}

	crf := opts.CRF
	preset := opts.Preset
	var videoEncoder string
	var losslessArgs []string
	crfMax := 63

	switch codec {
	case "av1", "libsvtav1", "svtav1":
		videoEncoder = "libsvtav1"
		losslessArgs = []string{"-svtav1-params", "lossless=1"}
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
		crfMax = 51
		if crf == 0 {
			crf = 30
		}
		if preset == "" {
			preset = "medium"
		}
	}
	crf = min(crf, crfMax)

	var vfFilters []string
	if !opts.Lossless {
		if p != nil && p.IsHDR() {
			vfFilters = append(vfFilters, probe.ToneMapFilter())
		} else {
			maxH := opts.MaxHeight
			if maxH == 0 {
				maxH = 1080
			}
			vfFilters = append(vfFilters, fmt.Sprintf("scale='min(1920,iw)':'min(%d,ih)':force_original_aspect_ratio=decrease,scale=trunc(iw/2)*2:trunc(ih/2)*2", maxH))
		}
	}

	var args []string
	args = append(args, "-i", inputPath)
	if opts.Compat {
		args = append(args, "-map", "0:v:0")
	}

	if len(vfFilters) > 0 {
		args = append(args, "-vf", strings.Join(vfFilters, ","))
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
		videoBitrateKbps := totalBitrateKbps - audioKbps
		if videoBitrateKbps < 64 {
			videoBitrateKbps = 64
		}
		args = append(args,
			"-b:v", fmt.Sprintf("%dk", int(videoBitrateKbps)),
			"-maxrate", fmt.Sprintf("%dk", int(videoBitrateKbps*1.5)),
			"-bufsize", fmt.Sprintf("%dk", int(videoBitrateKbps*2)),
		)
	default:
		args = append(args, "-crf", fmt.Sprintf("%d", crf))
		if preset != "" {
			args = append(args, "-preset", preset)
		}
	}

	if opts.Compat {
		args = append(args, "-pix_fmt", "yuv420p", "-fps_mode", "cfr")
	}

	audioBitrate := opts.AudioBitrate
	if audioBitrate == "" {
		audioBitrate = "128k"
	}
	switch {
	case opts.Compat && p != nil && len(p.AudioStreams()) == 0:
		args = append(args, "-an")
	case opts.Compat:
		args = append(args, "-map", "0:a:0", "-c:a", "aac", "-b:a", audioBitrate, "-ac", "2", "-ar", "48000")
	default:
		args = append(args, "-c:a", "aac", "-b:a", audioBitrate)
	}

	switch opts.Subs {
	case "all":
		if p != nil {
			subCount := len(p.SubtitleStreams())
			for i := range subCount {
				args = append(args, "-map", fmt.Sprintf("0:s:%d", i))
			}
			if subCount > 0 {
				args = append(args, "-c:s", "mov_text")
			}
		}
	case "none":
		args = append(args, "-sn")
	}

	args = append(args, "-movflags", "+faststart")

	suffix := "optimized"
	if opts.Lossless {
		suffix = "lossless"
	}
	if opts.CustomSuffix != "" {
		suffix = opts.CustomSuffix
	}

	targetExt := opts.TargetExt
	if targetExt == "" {
		targetExt = "mp4"
	}

	return &OpResult{
		Args:      args,
		Suffix:    suffix,
		TargetExt: targetExt,
	}, nil
}
