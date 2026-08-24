package cmd

import (
	"math"

	"github.com/spf13/cobra"

	"github.com/Tanq16/goff/internal/ops"
	"github.com/Tanq16/goff/internal/probe"
)

var transformFlags struct {
	rotate string
	scale  string
	speed  float64
	crop   bool
	mute   bool
}

var rotations = []string{"90", "180", "270", "hflip", "vflip"}

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
		cmd.Flags().Var(newEnum(&transformFlags.rotate, "", rotations...), "rotate", "Also rotate")
	}
	if own != "scale" {
		cmd.Flags().Var(newScale(&transformFlags.scale), "scale", "Also scale")
	}
	if own != "speed" {
		cmd.Flags().Var(newBoundedFloat(&transformFlags.speed, 0, 0, math.MaxFloat64, "greater than 0"), "speed", "Also change speed, e.g. 1.5 (audio pitch corrected)")
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
		runTransform("rotate", args)
	},
}

var scaleCmd = &cobra.Command{
	Use:   "scale <files...>",
	Short: "Resize video to a resolution tier or explicit dimensions",
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		runTransform("scale", args)
	},
}

var speedCmd = &cobra.Command{
	Use:   "speed <files...>",
	Short: "Change playback speed with pitch-corrected audio",
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
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
	rotateCmd.Flags().Var(newEnum(&transformFlags.rotate, "", rotations...), "by", "Rotation to apply")
	rotateCmd.MarkFlagRequired("by")
	scaleCmd.Flags().Var(newScale(&transformFlags.scale), "by", "Resolution to scale to")
	scaleCmd.MarkFlagRequired("by")
	speedCmd.Flags().Var(newBoundedFloat(&transformFlags.speed, 0, 0, math.MaxFloat64, "greater than 0"), "by", "Speed multiplier, e.g. 1.5")
	speedCmd.MarkFlagRequired("by")

	addTransformFlags(rotateCmd, "rotate")
	addTransformFlags(scaleCmd, "scale")
	addTransformFlags(speedCmd, "speed")
	addTransformFlags(cropCmd, "crop")
	addTransformFlags(muteCmd, "mute")

	rootCmd.AddCommand(rotateCmd, scaleCmd, speedCmd, cropCmd, muteCmd)
}
