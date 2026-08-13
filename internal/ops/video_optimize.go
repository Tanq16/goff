package ops

import (
	"fmt"
	"strings"

	"github.com/Tanq16/goff/internal/probe"
)

type VideoOptimizeOpts struct {
	Codec        string
	CRF          int
	Preset       string
	MaxHeight    int
	TargetSizeMB float64
	AudioBitrate string
	CustomSuffix string
	TargetExt    string
}

func BuildVideoOptimize(inputPath string, p *probe.ProbeResult, opts VideoOptimizeOpts) (*OpResult, error) {
	codec := strings.ToLower(opts.Codec)
	if codec == "" {
		codec = "hevc"
	}

	crf := opts.CRF
	preset := opts.Preset
	var videoEncoder string

	switch codec {
	case "av1", "libsvtav1", "svtav1":
		videoEncoder = "libsvtav1"
		if crf == 0 {
			crf = 32
		}
		if preset == "" {
			preset = "6"
		}
	case "h264", "libx264", "x264":
		videoEncoder = "libx264"
		if crf == 0 {
			crf = 23
		}
		if preset == "" {
			preset = "medium"
		}
	default:
		videoEncoder = "libx265"
		if crf == 0 {
			crf = 30
		}
		if preset == "" {
			preset = "medium"
		}
	}

	var vfFilters []string
	if p != nil && p.IsHDR() {
		vfFilters = append(vfFilters, probe.ToneMapFilter())
	} else {
		maxH := opts.MaxHeight
		if maxH == 0 {
			maxH = 1080
		}
		vfFilters = append(vfFilters, fmt.Sprintf("scale='min(1920,iw)':'min(%d,ih)':force_original_aspect_ratio=decrease,scale=trunc(iw/2)*2:trunc(ih/2)*2", maxH))
	}

	var args []string
	args = append(args, "-i", inputPath)

	if len(vfFilters) > 0 {
		args = append(args, "-vf", strings.Join(vfFilters, ","))
	}

	args = append(args, "-c:v", videoEncoder)

	if opts.TargetSizeMB > 0 && p != nil && p.TotalDuration() > 0 {
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
	} else {
		args = append(args, "-crf", fmt.Sprintf("%d", crf))
		if preset != "" {
			args = append(args, "-preset", preset)
		}
	}

	audioBitrate := opts.AudioBitrate
	if audioBitrate == "" {
		audioBitrate = "128k"
	}
	args = append(args, "-c:a", "aac", "-b:a", audioBitrate)
	args = append(args, "-movflags", "+faststart")

	suffix := "optimized"
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
