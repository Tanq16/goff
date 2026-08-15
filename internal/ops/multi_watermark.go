package ops

import (
	"fmt"
)

type WatermarkPosition string

const (
	PosTopRight    WatermarkPosition = "top-right"
	PosTopLeft     WatermarkPosition = "top-left"
	PosBottomRight WatermarkPosition = "bottom-right"
	PosBottomLeft  WatermarkPosition = "bottom-left"
	PosCenter      WatermarkPosition = "center"
)

type MultiWatermarkOpts struct {
	VideoInput     string
	WatermarkInput string
	Position       WatermarkPosition
	ScalePercent   int
	MarginPercent  int
	Opacity        float64
}

func BuildMultiWatermark(opts MultiWatermarkOpts) (*OpResult, error) {
	if opts.VideoInput == "" || opts.WatermarkInput == "" {
		return nil, fmt.Errorf("watermark operation requires both video input and watermark input")
	}

	pos := opts.Position
	if pos == "" {
		pos = PosTopRight
	}

	scalePct := opts.ScalePercent
	if scalePct <= 0 {
		scalePct = 15
	}

	marginPct := opts.MarginPercent
	if marginPct < 0 {
		marginPct = 2
	}
	mFrac := float64(marginPct) / 100.0
	sFrac := float64(scalePct) / 100.0

	var xExpr, yExpr string
	switch pos {
	case PosTopLeft:
		xExpr = fmt.Sprintf("main_w*%f", mFrac)
		yExpr = fmt.Sprintf("main_h*%f", mFrac)
	case PosBottomLeft:
		xExpr = fmt.Sprintf("main_w*%f", mFrac)
		yExpr = fmt.Sprintf("main_h-overlay_h-(main_h*%f)", mFrac)
	case PosBottomRight:
		xExpr = fmt.Sprintf("main_w-overlay_w-(main_w*%f)", mFrac)
		yExpr = fmt.Sprintf("main_h-overlay_h-(main_h*%f)", mFrac)
	case PosCenter:
		xExpr = "(main_w-overlay_w)/2"
		yExpr = "(main_h-overlay_h)/2"
	case PosTopRight:
		fallthrough
	default:
		xExpr = fmt.Sprintf("main_w-overlay_w-(main_w*%f)", mFrac)
		yExpr = fmt.Sprintf("main_h*%f", mFrac)
	}

	var filterComplex string
	if opts.Opacity > 0 && opts.Opacity < 1.0 {
		filterComplex = fmt.Sprintf("[1:v]format=rgba,colorchannelmixer=aa=%.2f[wm];[wm][0:v]scale2ref=w='main_w*%.4f':h=-1[logo][vid];[vid][logo]overlay=x='%s':y='%s'[outv]", opts.Opacity, sFrac, xExpr, yExpr)
	} else {
		filterComplex = fmt.Sprintf("[1:v][0:v]scale2ref=w='main_w*%.4f':h=-1[logo][vid];[vid][logo]overlay=x='%s':y='%s'[outv]", sFrac, xExpr, yExpr)
	}

	args := []string{
		"-i", opts.VideoInput,
		"-i", opts.WatermarkInput,
		"-filter_complex", filterComplex,
		"-map", "[outv]",
		"-map", "0:a?",
		"-c:v", "libx264",
		"-crf", "20",
		"-preset", "medium",
		"-c:a", "copy",
		"-movflags", "+faststart",
	}

	return &OpResult{
		Args:      args,
		Suffix:    "watermarked",
		TargetExt: "mp4",
	}, nil
}
