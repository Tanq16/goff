package cmd

import (
	"github.com/spf13/cobra"

	"github.com/Tanq16/goff/internal/ops"
	"github.com/Tanq16/goff/internal/probe"
	"github.com/Tanq16/goff/utils"
)

var trimFlags struct {
	start    string
	end      string
	duration string
	accurate bool
}

var trimCmd = &cobra.Command{
	Use:     "trim <files...>",
	GroupID: "segments",
	Short:   "Cut a time range out of a video or audio file",
	Example: `  goff trim video.mp4 --start 00:01:30 --end 00:02:00
  goff trim video.mp4 --start 90 --duration 30 --accurate`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if trimFlags.start == "" && trimFlags.end == "" && trimFlags.duration == "" {
			utils.PrintFatal("trim needs at least one of --start, --end, or --duration", nil)
		}

		runFiles("trim", args, func(input string, p *probe.ProbeResult) (*ops.OpResult, error) {
			return ops.BuildVideoTrim(input, p, ops.VideoTrimOpts{
				Start:    trimFlags.start,
				End:      trimFlags.end,
				Duration: trimFlags.duration,
				Accurate: trimFlags.accurate,
			})
		})
	},
}

func init() {
	trimCmd.Flags().Var(newTimestamp(&trimFlags.start), "start", "Start position, e.g. 00:01:30 or 90")
	trimCmd.Flags().Var(newTimestamp(&trimFlags.end), "end", "End position, e.g. 00:02:00")
	trimCmd.Flags().Var(newTimestamp(&trimFlags.duration), "duration", "Length to keep from --start, e.g. 30")
	trimCmd.Flags().BoolVar(&trimFlags.accurate, "accurate", false, "Re-encode for frame-accurate cuts instead of copying streams")
	trimCmd.MarkFlagsMutuallyExclusive("end", "duration")
}
