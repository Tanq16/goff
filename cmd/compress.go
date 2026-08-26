package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/Tanq16/goff/internal/ops"
	"github.com/Tanq16/goff/internal/probe"
	"github.com/Tanq16/goff/utils"
)

var compressFlags struct {
	codec    string
	crf      int
	height   int
	size     string
	lossless bool
	compat   bool
}

func parseSizeBudget(s string) (float64, error) {
	trimmed := strings.TrimSpace(strings.ToLower(s))
	trimmed = strings.TrimSuffix(trimmed, "b")
	trimmed = strings.TrimSuffix(trimmed, "m")
	trimmed = strings.TrimSpace(trimmed)
	if trimmed == "" {
		return 0, fmt.Errorf("empty size budget")
	}
	limit, err := strconv.ParseFloat(trimmed, 64)
	if err != nil {
		return 0, fmt.Errorf("size must look like 25MB, got %q", s)
	}
	if limit <= 0 {
		return 0, fmt.Errorf("size must be greater than zero, got %q", s)
	}
	return limit - max(limit*0.02, 0.5), nil
}

var compressCmd = &cobra.Command{
	Use:     "compress <files...>",
	GroupID: "video",
	Short:   "Re-encode video smaller, with optional codec, quality, and size targets",
	Example: `  goff compress movie.mkv
  goff compress movie.mkv --codec av1 --crf 28
  goff compress clip.mp4 --size 25MB
  goff compress master.mov --lossless`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var targetMB float64
		suffix := ""
		if compressFlags.size != "" {
			budget, err := parseSizeBudget(compressFlags.size)
			if err != nil {
				utils.PrintFatal("invalid --size", err)
			}
			targetMB = budget
			suffix = strings.ToLower(strings.TrimSpace(compressFlags.size))
		}

		runFiles("compress", args, func(input string, p *probe.ProbeResult) (*ops.OpResult, error) {
			return ops.BuildVideoOptimize(input, p, ops.VideoOptimizeOpts{
				Codec:        compressFlags.codec,
				CRF:          compressFlags.crf,
				Lossless:     compressFlags.lossless,
				MaxHeight:    compressFlags.height,
				TargetSizeMB: targetMB,
				CustomSuffix: suffix,
				Compat:       compressFlags.compat,
			})
		})
	},
}

func init() {
	compressCmd.Flags().VarP(newEnum(&compressFlags.codec, "hevc", "hevc", "av1", "h264"), "codec", "c", "Video codec")
	compressCmd.Flags().Var(newBoundedInt(&compressFlags.crf, 0, 1, 63), "crf", "Quality factor, lower is better (codec default when unset)")
	compressCmd.Flags().Var(newBoundedInt(&compressFlags.height, 1080, 144, 4320), "height", "Maximum output height in pixels")
	compressCmd.Flags().StringVar(&compressFlags.size, "size", "", "Target file size budget, e.g. 25MB (overrides --crf)")
	compressCmd.Flags().BoolVar(&compressFlags.lossless, "lossless", false, "Re-encode with no video quality loss, keeping the source resolution")
	compressCmd.Flags().BoolVar(&compressFlags.compat, "compat", false, "Normalize for wide playback: first video and audio track, 8-bit, constant frame rate, 48kHz stereo")
	compressCmd.MarkFlagsMutuallyExclusive("crf", "size", "lossless")
	compressCmd.MarkFlagsMutuallyExclusive("height", "lossless")
	compressCmd.MarkFlagsMutuallyExclusive("compat", "lossless")
}
