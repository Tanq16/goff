package cmd

import (
	"github.com/spf13/cobra"

	"github.com/Tanq16/goff/internal/ops"
	"github.com/Tanq16/goff/internal/probe"
)

var remuxFlags struct {
	to string
}

var remuxCmd = &cobra.Command{
	Use:     "remux <files...>",
	GroupID: "video",
	Short:   "Switch container without re-encoding a single stream",
	Example: "  goff remux capture.mkv --to mp4",
	Args:    cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runFiles("remux", args, func(input string, p *probe.ProbeResult) (*ops.OpResult, error) {
			return ops.BuildVideoRemux(input, p, ops.VideoRemuxOpts{
				TargetExt: remuxFlags.to,
			})
		})
	},
}

func init() {
	remuxCmd.Flags().Var(newEnum(&remuxFlags.to, "mp4", "mp4", "mkv", "mov"), "to", "Target container")
}
