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
		ext = strings.TrimPrefix(filepath.Ext(inputPath), ".")
		if ext == "" {
			ext = "mp3"
		}
	}

	args := []string{
		"-i", inputPath,
		"-af", filter,
	}

	if p != nil && p.IsVideo() {
		args = append(args, "-c:v", "copy", "-c:a", "aac", "-b:a", "192k")
		if ext == "" {
			ext = "mp4"
		}
	} else {
		if ext == "mp3" {
			args = append(args, "-c:a", "libmp3lame", "-b:a", "320k")
		} else if ext == "m4a" || ext == "aac" {
			args = append(args, "-c:a", "aac", "-b:a", "256k")
		} else if ext == "flac" {
			args = append(args, "-c:a", "flac")
		}
	}

	return &OpResult{
		Args:      args,
		Suffix:    "normalized",
		TargetExt: ext,
	}, nil
}
