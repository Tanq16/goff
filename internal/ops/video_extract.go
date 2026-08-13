package ops

import (
	"strings"

	"github.com/Tanq16/goff/internal/probe"
)

type VideoExtractOpts struct {
	Format  string
	Bitrate string
}

func BuildVideoExtract(inputPath string, p *probe.ProbeResult, opts VideoExtractOpts) (*OpResult, error) {
	format := strings.ToLower(opts.Format)
	if format == "" {
		format = "mp3"
	}

	bitrate := opts.Bitrate
	var codec string
	var ext string

	switch format {
	case "aac", "m4a":
		codec = "aac"
		ext = "m4a"
		if bitrate == "" {
			bitrate = "256k"
		}
	case "opus":
		codec = "libopus"
		ext = "opus"
		if bitrate == "" {
			bitrate = "160k"
		}
	case "flac":
		codec = "flac"
		ext = "flac"
	case "wav":
		codec = "pcm_s16le"
		ext = "wav"
	default:
		codec = "libmp3lame"
		ext = "mp3"
		if bitrate == "" {
			bitrate = "320k"
		}
	}

	args := []string{
		"-i", inputPath,
		"-vn",
		"-c:a", codec,
	}

	if bitrate != "" && codec != "flac" && codec != "pcm_s16le" {
		args = append(args, "-b:a", bitrate)
	}

	return &OpResult{
		Args:      args,
		Suffix:    "audio",
		TargetExt: ext,
	}, nil
}
