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

var transformDimensions = []string{"rotate", "scale", "speed", "crop", "mute"}

const (
	speedMin = 0.5
	speedMax = 100.0
)

var transformCmd = &cobra.Command{
	Use:     "transform <files...>",
	GroupID: "video",
	Short:   "Rotate, scale, retime, crop, or mute video in one encode",
	Long: `Rotate, scale, retime, crop, or mute video in one encode.

Every dimension given is applied in a single pass, so combining them costs one
generation of quality rather than one per change.`,
	Example: `  goff transform clip.mp4 --rotate 90
  goff transform clip.mp4 --scale 720p --mute
  goff transform lecture.mp4 --speed 1.5
  goff transform clip.mp4 --rotate 90 --scale 720p --mute`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		opts := ops.VideoTransformOpts{
			Rotate:       transformFlags.rotate,
			Scale:        transformFlags.scale,
			Speed:        transformFlags.speed,
			CropVertical: transformFlags.crop,
			StripAudio:   transformFlags.mute,
		}

		set := 0
		for _, name := range transformDimensions {
			if cmd.Flags().Changed(name) {
				set++
			}
		}
		if set > 1 {
			opts.CustomSuffix = "transform"
		}

		runFiles("transform", args, func(input string, p *probe.ProbeResult) (*ops.OpResult, error) {
			return ops.BuildVideoTransform(input, p, opts)
		})
	},
}

func init() {
	transformCmd.Flags().Var(newEnum(&transformFlags.rotate, "", rotations...), "rotate", "Rotation or flip to apply")
	transformCmd.Flags().Var(newScale(&transformFlags.scale), "scale", "Resolution to scale to")
	transformCmd.Flags().Var(newBoundedFloat(&transformFlags.speed, 0, speedMin, speedMax), "speed", "Speed multiplier, e.g. 1.5 (audio pitch corrected)")
	transformCmd.Flags().BoolVar(&transformFlags.crop, "crop", false, "Center-crop to 9:16 vertical")
	transformCmd.Flags().BoolVar(&transformFlags.mute, "mute", false, "Drop every audio track")
	transformCmd.MarkFlagsOneRequired(transformDimensions...)

	addOutputFlag(transformCmd)
	addJobsFlag(transformCmd)
}
