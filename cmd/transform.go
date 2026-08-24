package cmd

import (
	"strconv"

	"github.com/spf13/cobra"

	"github.com/Tanq16/goff/internal/ops"
	"github.com/Tanq16/goff/internal/probe"
	"github.com/Tanq16/goff/utils"
)

var transformFlags struct {
	by     string
	rotate string
	scale  string
	speed  float64
	crop   bool
	mute   bool
}

func runTransform(verb string, args []string) {
	opts := ops.VideoTransformOpts{
		Rotate:       transformFlags.rotate,
		Scale:        transformFlags.scale,
		Speed:        transformFlags.speed,
		CropVertical: transformFlags.crop,
		StripAudio:   transformFlags.mute,
	}

	set := 0
	for _, on := range []bool{opts.Rotate != "", opts.Scale != "", opts.Speed > 0, opts.CropVertical, opts.StripAudio} {
		if on {
			set++
		}
	}
	if set > 1 {
		opts.CustomSuffix = verb
	}

	runFiles(verb, args, func(input string, p *probe.ProbeResult) (*ops.OpResult, error) {
		return ops.BuildVideoTransform(input, p, opts)
	})
}

func addTransformFlags(cmd *cobra.Command, own string) {
	if own != "rotate" {
		cmd.Flags().StringVar(&transformFlags.rotate, "rotate", "", "Also rotate: 90, 180, 270, hflip, vflip")
	}
	if own != "scale" {
		cmd.Flags().StringVar(&transformFlags.scale, "scale", "", "Also scale: 720p, 1080p, 4k, or WxH")
	}
	if own != "speed" {
		cmd.Flags().Float64Var(&transformFlags.speed, "speed", 0, "Also change speed, e.g. 1.5 (audio pitch corrected)")
	}
	if own != "crop" {
		cmd.Flags().BoolVar(&transformFlags.crop, "crop", false, "Also center-crop to 9:16 vertical")
	}
	if own != "mute" {
		cmd.Flags().BoolVar(&transformFlags.mute, "mute", false, "Also drop every audio track")
	}
}

var rotateCmd = &cobra.Command{
	Use:   "rotate <files...>",
	Short: "Rotate or flip video",
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if transformFlags.by == "" {
			utils.PrintFatal("rotate needs --by (90, 180, 270, hflip, or vflip)", nil)
		}
		transformFlags.rotate = transformFlags.by
		runTransform("rotate", args)
	},
}

var scaleCmd = &cobra.Command{
	Use:   "scale <files...>",
	Short: "Resize video to a resolution tier or explicit dimensions",
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if transformFlags.by == "" {
			utils.PrintFatal("scale needs --by (720p, 1080p, 4k, or WxH)", nil)
		}
		transformFlags.scale = transformFlags.by
		runTransform("scale", args)
	},
}

var speedCmd = &cobra.Command{
	Use:   "speed <files...>",
	Short: "Change playback speed with pitch-corrected audio",
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if transformFlags.by == "" {
			utils.PrintFatal("speed needs --by, e.g. 1.5", nil)
		}
		factor, err := strconv.ParseFloat(transformFlags.by, 64)
		if err != nil || factor <= 0 {
			utils.PrintFatal("--by must be a positive multiplier, e.g. 1.5", err)
		}
		transformFlags.speed = factor
		runTransform("speed", args)
	},
}

var cropCmd = &cobra.Command{
	Use:   "crop <files...>",
	Short: "Center-crop video to 9:16 vertical",
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		transformFlags.crop = true
		runTransform("crop", args)
	},
}

var muteCmd = &cobra.Command{
	Use:   "mute <files...>",
	Short: "Drop every audio track from video",
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		transformFlags.mute = true
		runTransform("mute", args)
	},
}

func init() {
	rotateCmd.Flags().StringVar(&transformFlags.by, "by", "", "Rotation: 90, 180, 270, hflip, vflip")
	scaleCmd.Flags().StringVar(&transformFlags.by, "by", "", "Target: 720p, 1080p, 4k, or WxH")
	speedCmd.Flags().StringVar(&transformFlags.by, "by", "", "Speed multiplier, e.g. 1.5")

	addTransformFlags(rotateCmd, "rotate")
	addTransformFlags(scaleCmd, "scale")
	addTransformFlags(speedCmd, "speed")
	addTransformFlags(cropCmd, "crop")
	addTransformFlags(muteCmd, "mute")

	rootCmd.AddCommand(rotateCmd, scaleCmd, speedCmd, cropCmd, muteCmd)
}
