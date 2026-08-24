package cmd

import (
	"github.com/spf13/cobra"

	"github.com/Tanq16/goff/internal/ops"
	"github.com/Tanq16/goff/internal/probe"
	"github.com/Tanq16/goff/utils"
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
		var format ops.HLSFormat
		switch hlsFlags.to {
		case "fmp4":
			format = ops.HLSFormatFMP4
		case "ts", "mpegts":
			format = ops.HLSFormatMPEGTS
		default:
			utils.PrintFatal("--to must be fmp4 or ts, got "+hlsFlags.to, nil)
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
	hlsCmd.Flags().StringVar(&hlsFlags.to, "to", "fmp4", "Segment type: fmp4 or ts")
	hlsCmd.Flags().IntVar(&hlsFlags.segment, "segment", 6, "Segment length in seconds")
	hlsCmd.Flags().IntVar(&hlsFlags.crf, "crf", 21, "Quality factor, lower is better")

	rootCmd.AddCommand(hlsCmd)
}
