package ops

import (
	"fmt"
	"strconv"

	"github.com/Tanq16/goff/internal/probe"
)

type ThumbnailOpts struct {
	At        string
	AtSeconds float64
	Width     int
}

func BuildThumbnail(inputPath string, p *probe.ProbeResult, opts ThumbnailOpts) (*OpResult, error) {
	if !p.IsVideo() {
		return nil, fmt.Errorf("no video stream to capture a frame from")
	}

	total := p.TotalDuration()
	at := opts.At
	if at == "" {
		at = strconv.FormatFloat(total/2, 'f', 3, 64)
	} else if total > 0 && opts.AtSeconds >= total {
		return nil, fmt.Errorf("--at %s is past the end of the file (%s)", at, probe.FormatDuration(total))
	}

	width := opts.Width
	if width <= 0 {
		width = 640
	}

	args := []string{
		"-ss", at,
		"-i", inputPath,
		"-map", "0:v:0",
		"-frames:v", "1",
		"-vf", fmt.Sprintf("scale=%d:-2", width),
		"-q:v", "2",
		"-f", "image2",
	}

	return &OpResult{
		Args:      args,
		Suffix:    "thumbnail",
		TargetExt: "jpg",
	}, nil
}
