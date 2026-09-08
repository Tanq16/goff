package ops

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Tanq16/goff/internal/probe"
)

type AudioLoudnormOpts struct {
	IntegratedLoudness float64
	TruePeak           float64
	LoudnessRange      float64
	OutputExt          string
}

func BuildAudioLoudnorm(inputPath string, p *probe.ProbeResult, opts AudioLoudnormOpts) (*OpResult, error) {
	i := opts.IntegratedLoudness
	if i == 0 {
		i = -16.0
	}

	tp := opts.TruePeak
	if tp == 0 {
		tp = -1.5
	}

	lra := opts.LoudnessRange
	if lra == 0 {
		lra = 11.0
	}

	filter := fmt.Sprintf("loudnorm=I=%.1f:TP=%.1f:LRA=%.1f", i, tp, lra)

	ext := strings.TrimPrefix(opts.OutputExt, ".")
	if ext == "" {
		if p != nil && p.IsVideo() {
			ext = "mp4"
		} else {
			ext = strings.TrimPrefix(filepath.Ext(inputPath), ".")
			if ext == "" {
				ext = "mp3"
			}
		}
	}

	args := []string{
		"-i", inputPath,
		"-af", filter,
	}

	isAudioExt := ext == "mp3" || ext == "m4a" || ext == "aac" || ext == "flac" || ext == "wav" || ext == "ogg" || ext == "opus"

	if p != nil && p.IsVideo() && !isAudioExt {
		args = append(args, "-c:v", "copy", "-c:a", "aac", "-b:a", "192k")
		args = tagHEVC(args, copiedVideoIsHEVC(p), ext)
		if containerUsesMOVMuxer(ext) {
			args = append(args, "-movflags", "+faststart")
		}
	} else {
		if p != nil && p.IsVideo() {
			args = append(args, "-vn")
		}
		switch ext {
		case "mp3":
			args = append(args, "-c:a", "libmp3lame", "-b:a", "320k")
		case "m4a", "aac":
			args = append(args, "-c:a", "aac", "-b:a", "256k")
		case "flac":
			args = append(args, "-c:a", "flac")
		case "opus":
			args = append(args, "-c:a", "libopus", "-b:a", "160k")
		case "wav":
			args = append(args, "-c:a", "pcm_s16le")
		case "ogg":
			args = append(args, "-c:a", "libvorbis", "-b:a", "192k")
		default:
			args = append(args, "-c:a", "libmp3lame", "-b:a", "320k")
		}
	}

	return &OpResult{
		Args:      args,
		Suffix:    "normalized",
		TargetExt: ext,
	}, nil
}
