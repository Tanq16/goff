package cmd

import (
	"github.com/spf13/cobra"

	"github.com/Tanq16/goff/internal/ops"
	"github.com/Tanq16/goff/internal/probe"
)

var thumbnailFlags struct {
	at        string
	atSeconds float64
	width     int
}

var thumbnailCmd = &cobra.Command{
	Use:     "thumbnail <files...>",
	GroupID: "segments",
	Short:   "Capture one frame from video as a JPEG still",
	Example: `  goff thumbnail movie.mp4
  goff thumbnail movie.mp4 --at 00:01:30 --width 1280`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runFiles("thumbnail", args, func(input string, p *probe.ProbeResult) (*ops.OpResult, error) {
			return ops.BuildThumbnail(input, p, ops.ThumbnailOpts{
				At:        thumbnailFlags.at,
				AtSeconds: thumbnailFlags.atSeconds,
				Width:     thumbnailFlags.width,
			})
		})
	},
}

func init() {
	thumbnailCmd.Flags().Var(newTimestampSeconds(&thumbnailFlags.at, &thumbnailFlags.atSeconds), "at", "Position to capture, e.g. 00:01:30 (midpoint when unset)")
	thumbnailCmd.Flags().Var(newBoundedInt(&thumbnailFlags.width, 640, 16, 3840), "width", "Output width in pixels, height follows the aspect ratio")

	addOutputFlag(thumbnailCmd)
	addJobsFlag(thumbnailCmd)
}
