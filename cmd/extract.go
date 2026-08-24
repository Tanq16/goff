package cmd

import (
	"github.com/spf13/cobra"

	"github.com/Tanq16/goff/internal/ops"
	"github.com/Tanq16/goff/internal/probe"
)

var extractFlags struct {
	to      string
	bitrate string
}

var extractCmd = &cobra.Command{
	Use:   "extract <files...>",
	Short: "Pull the audio stream out of a video into an audio file",
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		runFiles("extract", args, func(input string, p *probe.ProbeResult) (*ops.OpResult, error) {
			return ops.BuildVideoExtract(input, p, ops.VideoExtractOpts{
				Format:  extractFlags.to,
				Bitrate: extractFlags.bitrate,
			})
		})
	},
}

func init() {
	extractCmd.Flags().Var(newEnum(&extractFlags.to, "mp3", "mp3", "aac", "m4a", "flac", "opus", "wav"), "to", "Audio format")
	extractCmd.Flags().StringVar(&extractFlags.bitrate, "bitrate", "", "Audio bitrate, e.g. 320k (format default when unset)")

	rootCmd.AddCommand(extractCmd)
}
