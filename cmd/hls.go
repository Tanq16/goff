package cmd

import (
	"os"

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
	Use:     "hls <files...>",
	GroupID: "video",
	Short:   "Package video into an HLS VoD playlist with segments",
	Long: `Package video into an HLS VoD playlist with segments.

Each input gets its own directory holding index.m3u8 and the segments, so two
packaged videos never share a segment name.`,
	Example: `  goff hls lecture.mp4
  goff hls lecture.mp4 --to ts --segment 4`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		format := ops.HLSFormatFMP4
		suffix := "hls-fmp4"
		if hlsFlags.to == "ts" {
			format = ops.HLSFormatMPEGTS
			suffix = "hls-ts"
		}

		runFiles("hls", args, func(input string, p *probe.ProbeResult) (*ops.OpResult, error) {
			dir, err := claimOutputDir(input, suffix)
			if err != nil {
				return nil, err
			}
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return nil, err
			}
			return ops.BuildVideoHLS(input, p, ops.VideoHLSOpts{
				Format:          format,
				SegmentDuration: hlsFlags.segment,
				CRF:             hlsFlags.crf,
				OutputDir:       dir,
			})
		})
	},
}

func init() {
	hlsCmd.Flags().Var(newEnum(&hlsFlags.to, "fmp4", "fmp4", "ts"), "to", "Segment type")
	hlsCmd.Flags().Var(newBoundedInt(&hlsFlags.segment, 6, 1, 60, "between 1 and 60"), "segment", "Segment length in seconds")
	hlsCmd.Flags().Var(newBoundedInt(&hlsFlags.crf, 21, 0, 51, "between 0 and 51"), "crf", "Quality factor, lower is better")
}
