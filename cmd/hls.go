package cmd

import (
	"github.com/spf13/cobra"

	"github.com/Tanq16/goff/internal/ops"
	"github.com/Tanq16/goff/internal/probe"
)

var hlsFlags struct {
	to      string
	segment int
	crf     int
}

var hlsCmd = &cobra.Command{
	Use:   "hls <files...>",
	Short: "Package video into an HLS VoD playlist with segments",
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		format := ops.HLSFormatFMP4
		if hlsFlags.to == "ts" {
			format = ops.HLSFormatMPEGTS
		}

		runFiles("hls", args, func(input string, p *probe.ProbeResult) (*ops.OpResult, error) {
			return ops.BuildVideoHLS(input, p, ops.VideoHLSOpts{
				Format:          format,
				SegmentDuration: hlsFlags.segment,
				CRF:             hlsFlags.crf,
			})
		})
	},
}

func init() {
	hlsCmd.Flags().Var(newEnum(&hlsFlags.to, "fmp4", "fmp4", "ts"), "to", "Segment type")
	hlsCmd.Flags().IntVar(&hlsFlags.segment, "segment", 6, "Segment length in seconds")
	hlsCmd.Flags().IntVar(&hlsFlags.crf, "crf", 21, "Quality factor, lower is better")

	rootCmd.AddCommand(hlsCmd)
}
