package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/Tanq16/goff/internal/ops"
	"github.com/Tanq16/goff/internal/probe"
)

var hlsFlags struct {
	segmentType     string
	segmentDuration int
	crf             int
}

var hlsCmd = &cobra.Command{
	Use:     "hls <files...>",
	GroupID: "video",
	Short:   "Package video into an HLS VoD playlist with segments",
	Long: `Package video into an HLS VoD playlist with segments.

Each input gets its own directory holding index.m3u8 and the segments, so two
packaged videos never share a segment name.`,
	Example: `  goff hls lecture.mp4
  goff hls lecture.mp4 --segment-type ts --segment-duration 4`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		format := ops.HLSFormat(hlsFlags.segmentType)

		runFiles("hls", args, func(input string, p *probe.ProbeResult) (*ops.OpResult, error) {
			dir, err := claimOutputDir(input, format.Suffix())
			if err != nil {
				return nil, err
			}
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return nil, err
			}
			res, err := ops.BuildVideoHLS(input, p, ops.VideoHLSOpts{
				Format:          format,
				SegmentDuration: hlsFlags.segmentDuration,
				CRF:             hlsFlags.crf,
				OutputDir:       dir,
			})
			if err != nil {
				os.RemoveAll(dir)
				return nil, err
			}
			res.Cleanup = func() { os.RemoveAll(dir) }
			return res, nil
		})
	},
}

func init() {
	hlsCmd.Flags().Var(newEnum(&hlsFlags.segmentType, string(ops.HLSFormatFMP4), string(ops.HLSFormatFMP4), string(ops.HLSFormatMPEGTS)), "segment-type", "Segment type")
	hlsCmd.Flags().Var(newBoundedInt(&hlsFlags.segmentDuration, 6, 1, 60), "segment-duration", "Segment length in seconds")
	hlsCmd.Flags().Var(newBoundedInt(&hlsFlags.crf, 21, 1, 51), "crf", "Quality factor, lower is better")

	addOutputFlag(hlsCmd)
	addJobsFlag(hlsCmd)
}
