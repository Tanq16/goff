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

func lastFrameStart(p *probe.ProbeResult, total float64) float64 {
	if total <= 0 {
		return 0
	}
	if v := p.PrimaryVideoStream(); v != nil {
		if fps := probe.ParseFPS(v.RFrameRate); fps > 0 {
			return total - 1/fps
		}
	}
	return total
}

func BuildThumbnail(inputPath string, p *probe.ProbeResult, opts ThumbnailOpts) (*OpResult, error) {
	if !p.IsVideo() {
		return nil, fmt.Errorf("no video stream to capture a frame from")
	}

	total := p.TotalDuration()
	at := opts.At
	if at == "" {
		at = strconv.FormatFloat(total/2, 'f', 3, 64)
	} else if last := lastFrameStart(p, total); last > 0 && opts.AtSeconds > last {
		return nil, fmt.Errorf("--at %s is at or past the last frame of the file (%s)", at, probe.FormatDuration(total))
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
