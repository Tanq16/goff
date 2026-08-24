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
	Use:   "remux <files...>",
	Short: "Switch container without re-encoding a single stream",
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		requireOneOf("to", remuxFlags.to, "mp4", "mkv", "mov")

		runFiles("remux", args, func(input string, p *probe.ProbeResult) (*ops.OpResult, error) {
			return ops.BuildVideoRemux(input, p, ops.VideoRemuxOpts{
				TargetExt: remuxFlags.to,
			})
		})
	},
}

func init() {
	remuxCmd.Flags().StringVar(&remuxFlags.to, "to", "mp4", "Target container: mp4, mkv, mov")

	rootCmd.AddCommand(remuxCmd)
}
