package ops

import (
	"github.com/Tanq16/goff/internal/probe"
)

type VideoTrimOpts struct {
	Start    string
	End      string
	Duration string
	Accurate bool
}

func BuildVideoTrim(inputPath string, p *probe.ProbeResult, opts VideoTrimOpts) (*OpResult, error) {
	var args []string

	if !opts.Accurate && opts.Start != "" {
		args = append(args, "-ss", opts.Start)
	}

	args = append(args, "-i", inputPath)

	if opts.Accurate && opts.Start != "" {
		args = append(args, "-ss", opts.Start)
	}

	if opts.End != "" {
		args = append(args, "-to", opts.End)
	} else if opts.Duration != "" {
		args = append(args, "-t", opts.Duration)
	}

	if opts.Accurate {
		args = append(args, "-c:v", "libx264", "-crf", "18", "-preset", "fast", "-c:a", "aac", "-b:a", "192k")
	} else {
		args = append(args, "-c", "copy")
	}

	args = append(args, "-movflags", "+faststart")

	return &OpResult{
		Args:      args,
		Suffix:    "trimmed",
		TargetExt: "mp4",
	}, nil
}
