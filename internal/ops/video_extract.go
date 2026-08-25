package ops

import (
	"github.com/Tanq16/goff/internal/probe"
)

type VideoExtractOpts struct {
	Format  string
	Bitrate string
}

func BuildVideoExtract(inputPath string, p *probe.ProbeResult, opts VideoExtractOpts) (*OpResult, error) {
	args := []string{"-i", inputPath, "-vn"}

	encodeArgs, ext := audioEncodeArgs(opts.Format, opts.Bitrate)
	args = append(args, encodeArgs...)

	return &OpResult{
		Args:      args,
		Suffix:    "audio",
		TargetExt: ext,
	}, nil
}
