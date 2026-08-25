package cmd

import (
	"github.com/spf13/cobra"

	"github.com/Tanq16/goff/internal/ops"
	"github.com/Tanq16/goff/internal/probe"
)

var gifFlags struct {
	to       string
	width    int
	fps      int
	start    string
	duration string
}

var gifCmd = &cobra.Command{
	Use:     "gif <files...>",
	GroupID: "segments",
	Short:   "Render an animated GIF or WebP loop from video",
	Example: `  goff gif screencast.mp4 --width 640 --fps 20
  goff gif screencast.mp4 --start 5 --duration 8 --to webp`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runFiles("gif", args, func(input string, p *probe.ProbeResult) (*ops.OpResult, error) {
			return ops.BuildVideoGIF(input, p, ops.VideoGIFOpts{
				Format:   gifFlags.to,
				Width:    gifFlags.width,
				FPS:      gifFlags.fps,
				Start:    gifFlags.start,
				Duration: gifFlags.duration,
			})
		})
	},
}

func init() {
	gifCmd.Flags().Var(newEnum(&gifFlags.to, "gif", "gif", "webp"), "to", "Output format")
	gifCmd.Flags().Var(newBoundedInt(&gifFlags.width, 480, 16, 3840), "width", "Output width in pixels, height follows the aspect ratio")
	gifCmd.Flags().Var(newBoundedInt(&gifFlags.fps, 15, 1, 60), "fps", "Frames per second")
	gifCmd.Flags().Var(newTimestamp(&gifFlags.start), "start", "Start position, e.g. 00:00:05")
	gifCmd.Flags().Var(newTimestamp(&gifFlags.duration), "duration", "Length to capture from --start, e.g. 5")
}
