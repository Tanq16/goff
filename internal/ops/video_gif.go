package ops

import (
	"fmt"
	"strings"

	"github.com/Tanq16/goff/internal/probe"
)

type VideoGIFOpts struct {
	Width    int
	FPS      int
	Format   string
	Start    string
	Duration string
}

func BuildVideoGIF(inputPath string, p *probe.ProbeResult, opts VideoGIFOpts) (*OpResult, error) {
	width := opts.Width
	if width <= 0 {
		width = 480
	}

	fps := opts.FPS
	if fps <= 0 {
		fps = 15
	}

	format := strings.ToLower(opts.Format)
	if format == "" {
		format = "gif"
	}

	var args []string
	if opts.Start != "" {
		args = append(args, "-ss", opts.Start)
	}

	args = append(args, "-i", inputPath)

	if opts.Duration != "" {
		args = append(args, "-t", opts.Duration)
	}

	if format == "webp" {
		filter := fmt.Sprintf("fps=%d,scale=%d:-1:flags=lanczos", fps, width)
		args = append(args,
			"-vf", filter,
			"-vcodec", "libwebp",
			"-lossless", "0",
			"-compression_level", "6",
			"-q:v", "75",
			"-loop", "0",
		)
		return &OpResult{
			Args:      args,
			Suffix:    "animated",
			TargetExt: "webp",
		}, nil
	}

	filterComplex := fmt.Sprintf("fps=%d,scale=%d:-1:flags=lanczos,split[s0][s1];[s0]palettegen[p];[s1][p]paletteuse", fps, width)
	args = append(args, "-filter_complex", filterComplex)

	return &OpResult{
		Args:      args,
		Suffix:    "animated",
		TargetExt: "gif",
	}, nil
}
