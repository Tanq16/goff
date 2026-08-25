package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/Tanq16/goff/internal/ops"
	"github.com/Tanq16/goff/internal/probe"
	"github.com/Tanq16/goff/utils"
)

var watermarkFlags struct {
	logo    string
	at      string
	width   int
	margin  int
	opacity float64
}

var watermarkCmd = &cobra.Command{
	Use:     "watermark <video>",
	GroupID: "video",
	Short:   "Overlay a logo or image onto a video",
	Example: "  goff watermark clip.mp4 --logo logo.png --at bottom-right --width 12 --opacity 0.6",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		video := args[0]

		res, err := ops.BuildMultiWatermark(ops.MultiWatermarkOpts{
			VideoInput:     video,
			WatermarkInput: watermarkFlags.logo,
			Position:       ops.WatermarkPosition(watermarkFlags.at),
			ScalePercent:   watermarkFlags.width,
			MarginPercent:  watermarkFlags.margin,
			Opacity:        watermarkFlags.opacity,
		})
		if err != nil {
			utils.PrintFatal("failed to build watermark arguments", err)
		}

		var totalSec float64
		if p, probeErr := probe.RunProbe(context.Background(), video); probeErr == nil {
			totalSec = p.TotalDuration()
		}

		runComposed("watermark", video, res, totalSec)
	},
}

func init() {
	watermarkCmd.Flags().StringVar(&watermarkFlags.logo, "logo", "", "Image file to overlay (required)")
	watermarkCmd.MarkFlagRequired("logo")
	watermarkCmd.Flags().Var(newEnum(&watermarkFlags.at, "top-right", "top-right", "top-left", "bottom-right", "bottom-left", "center"), "at", "Overlay position")
	watermarkCmd.Flags().Var(newBoundedInt(&watermarkFlags.width, 15, 1, 100, "between 1 and 100"), "width", "Overlay width as a percent of video width")
	watermarkCmd.Flags().Var(newBoundedInt(&watermarkFlags.margin, 2, 0, 49, "between 0 and 49"), "margin", "Edge margin as a percent of video size")
	watermarkCmd.Flags().Var(newBoundedFloat(&watermarkFlags.opacity, 1.0, 0.01, 1.0, "between 0.01 and 1"), "opacity", "Overlay opacity")
}
