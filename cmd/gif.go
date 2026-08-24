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
	Use:   "gif <files...>",
	Short: "Render an animated GIF or WebP loop from video",
	Args:  cobra.ArbitraryArgs,
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
	gifCmd.Flags().IntVar(&gifFlags.width, "width", 480, "Output width in pixels, height follows the aspect ratio")
	gifCmd.Flags().IntVar(&gifFlags.fps, "fps", 15, "Frames per second")
	gifCmd.Flags().StringVar(&gifFlags.start, "start", "", "Start position, e.g. 00:00:05")
	gifCmd.Flags().StringVar(&gifFlags.duration, "duration", "", "Length to capture from --start, e.g. 5")

	rootCmd.AddCommand(gifCmd)
}
