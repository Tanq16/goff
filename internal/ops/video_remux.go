package ops

import (
	"strings"

	"github.com/Tanq16/goff/internal/probe"
)

type VideoRemuxOpts struct {
	TargetExt string
}

func BuildVideoRemux(inputPath string, p *probe.ProbeResult, opts VideoRemuxOpts) (*OpResult, error) {
	targetExt := strings.TrimPrefix(opts.TargetExt, ".")
	if targetExt == "" {
		targetExt = "mp4"
	}

	args := []string{
		"-i", inputPath,
		"-c", "copy",
		"-movflags", "+faststart",
	}

	return &OpResult{
		Args:      args,
		Suffix:    "remux",
		TargetExt: targetExt,
	}, nil
}
