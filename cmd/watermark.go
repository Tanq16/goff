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
	Use:   "watermark <video>",
	Short: "Overlay a logo or image onto a video",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		video := args[0]

		var pos ops.WatermarkPosition
		switch watermarkFlags.at {
		case "top-right":
			pos = ops.PosTopRight
		case "top-left":
			pos = ops.PosTopLeft
		case "bottom-right":
			pos = ops.PosBottomRight
		case "bottom-left":
			pos = ops.PosBottomLeft
		case "center":
			pos = ops.PosCenter
		default:
			utils.PrintFatal("--at must be top-right, top-left, bottom-right, bottom-left, or center, got "+watermarkFlags.at, nil)
		}

		res, err := ops.BuildMultiWatermark(ops.MultiWatermarkOpts{
			VideoInput:     video,
			WatermarkInput: watermarkFlags.logo,
			Position:       pos,
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
	watermarkCmd.Flags().StringVar(&watermarkFlags.at, "at", "top-right", "Corner: top-right, top-left, bottom-right, bottom-left, center")
	watermarkCmd.Flags().IntVar(&watermarkFlags.width, "width", 15, "Overlay width as a percent of video width")
	watermarkCmd.Flags().IntVar(&watermarkFlags.margin, "margin", 2, "Edge margin as a percent of video size")
	watermarkCmd.Flags().Float64Var(&watermarkFlags.opacity, "opacity", 1.0, "Overlay opacity from 0.0 to 1.0")

	rootCmd.AddCommand(watermarkCmd)
}
