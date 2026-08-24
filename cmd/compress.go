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
	codec  string
	crf    int
	height int
	size   string
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
	Use:   "compress <files...>",
	Short: "Re-encode video smaller, with optional codec, quality, and size targets",
	Args:  cobra.ArbitraryArgs,
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
				MaxHeight:    compressFlags.height,
				TargetSizeMB: targetMB,
				CustomSuffix: suffix,
			})
		})
	},
}

func init() {
	compressCmd.Flags().StringVarP(&compressFlags.codec, "codec", "c", "hevc", "Video codec: hevc, av1, h264")
	compressCmd.Flags().IntVar(&compressFlags.crf, "crf", 0, "Quality factor, lower is better (codec default when unset)")
	compressCmd.Flags().IntVar(&compressFlags.height, "height", 1080, "Maximum output height in pixels")
	compressCmd.Flags().StringVar(&compressFlags.size, "size", "", "Target file size budget, e.g. 25MB (overrides --crf)")
	compressCmd.MarkFlagsMutuallyExclusive("crf", "size")

	rootCmd.AddCommand(compressCmd)
}
