package ops

import (
	"fmt"
	"strings"

	"github.com/Tanq16/goff/internal/probe"
)

type AudioConvertOpts struct {
	Format     string
	Bitrate    string
	SampleRate int
	Channels   int
}

func BuildAudioConvert(inputPath string, p *probe.ProbeResult, opts AudioConvertOpts) (*OpResult, error) {
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
	case "ogg":
		codec = "libvorbis"
		ext = "ogg"
		if bitrate == "" {
			bitrate = "192k"
		}
	default:
		codec = "libmp3lame"
		ext = "mp3"
		if bitrate == "" {
			bitrate = "320k"
		}
	}

	args := []string{
		"-i", inputPath,
		"-c:a", codec,
	}

	if bitrate != "" && codec != "flac" && codec != "pcm_s16le" {
		args = append(args, "-b:a", bitrate)
	}

	if opts.SampleRate > 0 {
		args = append(args, "-ar", fmt.Sprintf("%d", opts.SampleRate))
	}

	if opts.Channels > 0 {
		args = append(args, "-ac", fmt.Sprintf("%d", opts.Channels))
	}

	return &OpResult{
		Args:      args,
		Suffix:    "converted",
		TargetExt: ext,
	}, nil
}
