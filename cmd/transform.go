package cmd

import (
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

const (
	speedMin    = 0.5
	speedMax    = 100.0
	speedBounds = "between 0.5 and 100"
)

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
		cmd.Flags().Var(newBoundedFloat(&transformFlags.speed, 0, speedMin, speedMax, speedBounds), "speed", "Also change speed, e.g. 1.5 (audio pitch corrected)")
	}
	if own != "crop" {
		cmd.Flags().BoolVar(&transformFlags.crop, "crop", false, "Also center-crop to 9:16 vertical")
	}
	if own != "mute" {
		cmd.Flags().BoolVar(&transformFlags.mute, "mute", false, "Also drop every audio track")
	}
}

var rotateCmd = &cobra.Command{
	Use:     "rotate <files...>",
	GroupID: "video",
	Short:   "Rotate or flip video",
	Example: `  goff rotate clip.mp4 --by 90
  goff rotate clip.mp4 --by 90 --scale 720p --mute`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runTransform("rotate", args)
	},
}

var scaleCmd = &cobra.Command{
	Use:     "scale <files...>",
	GroupID: "video",
	Short:   "Resize video to a resolution tier or explicit dimensions",
	Example: `  goff scale clip.mp4 --by 720p
  goff scale clip.mp4 --by 1280x720`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runTransform("scale", args)
	},
}

var speedCmd = &cobra.Command{
	Use:     "speed <files...>",
	GroupID: "video",
	Short:   "Change playback speed with pitch-corrected audio",
	Example: "  goff speed lecture.mp4 --by 1.5",
	Args:    cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runTransform("speed", args)
	},
}

var cropCmd = &cobra.Command{
	Use:     "crop <files...>",
	GroupID: "video",
	Short:   "Center-crop video to 9:16 vertical for Shorts and Reels",
	Example: "  goff crop landscape.mp4",
	Args:    cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		transformFlags.crop = true
		runTransform("crop", args)
	},
}

var muteCmd = &cobra.Command{
	Use:     "mute <files...>",
	GroupID: "video",
	Short:   "Drop every audio track from video",
	Example: "  goff mute clip.mp4",
	Args:    cobra.MinimumNArgs(1),
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
	speedCmd.Flags().Var(newBoundedFloat(&transformFlags.speed, 0, speedMin, speedMax, speedBounds), "by", "Speed multiplier, e.g. 1.5")
	speedCmd.MarkFlagRequired("by")

	addTransformFlags(rotateCmd, "rotate")
	addTransformFlags(scaleCmd, "scale")
	addTransformFlags(speedCmd, "speed")
	addTransformFlags(cropCmd, "crop")
	addTransformFlags(muteCmd, "mute")
}
