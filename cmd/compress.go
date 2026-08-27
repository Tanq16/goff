package cmd

import (
	"github.com/spf13/cobra"

	"github.com/Tanq16/goff/internal/ops"
	"github.com/Tanq16/goff/internal/probe"
)

var compressFlags struct {
	codec        string
	crf          int
	height       int
	size         string
	lossless     bool
	preset       string
	audioBitrate string
	compat       bool
	subs         string
	sizeBudgetMB float64
}

var compressCmd = &cobra.Command{
	Use:     "compress <files...>",
	GroupID: "video",
	Short:   "Re-encode video smaller, with optional codec, quality, and size targets",
	Example: `  goff compress movie.mkv
  goff compress movie.mkv --codec av1 --crf 28
  goff compress clip.mp4 --size 25MB
  goff compress master.mov --lossless
  goff compress movie.mkv --compat --subs all`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runFiles("compress", args, func(input string, p *probe.ProbeResult) (*ops.OpResult, error) {
			return ops.BuildVideoOptimize(input, p, ops.VideoOptimizeOpts{
				Codec:        compressFlags.codec,
				CRF:          compressFlags.crf,
				Lossless:     compressFlags.lossless,
				MaxHeight:    compressFlags.height,
				TargetSizeMB: compressFlags.sizeBudgetMB,
				CustomSuffix: compressFlags.size,
				Preset:       compressFlags.preset,
				AudioBitrate: compressFlags.audioBitrate,
				Compat:       compressFlags.compat,
				Subs:         compressFlags.subs,
			})
		})
	},
}

func init() {
	compressCmd.Flags().VarP(newEnum(&compressFlags.codec, "hevc", "hevc", "av1", "h264"), "codec", "c", "Video codec")
	compressCmd.Flags().Var(newBoundedInt(&compressFlags.crf, 0, 1, 63), "crf", "Quality factor, lower is better (codec default when unset)")
	compressCmd.Flags().Var(newBoundedInt(&compressFlags.height, 1080, 144, 4320), "height", "Maximum output height in pixels")
	compressCmd.Flags().Var(newSize(&compressFlags.size, &compressFlags.sizeBudgetMB), "size", "Target file size budget, e.g. 25MB (overrides --crf)")
	compressCmd.Flags().BoolVar(&compressFlags.lossless, "lossless", false, "Re-encode with no video quality loss, keeping the source resolution")
	compressCmd.Flags().StringVar(&compressFlags.preset, "preset", "", "Encoder speed preset, e.g. slow, or 6 for av1 (codec default when unset)")
	compressCmd.Flags().Var(newBitrate(&compressFlags.audioBitrate), "audio-bitrate", "Audio bitrate, e.g. 192k (128k when unset)")
	compressCmd.Flags().BoolVar(&compressFlags.compat, "compat", false, "Normalize for wide playback: first video and audio track, 8-bit, constant frame rate, 48kHz stereo")
	compressCmd.Flags().Var(newEnum(&compressFlags.subs, "auto", "auto", "all", "none"), "subs", "Subtitle handling")
	compressCmd.MarkFlagsMutuallyExclusive("crf", "size", "lossless")
	compressCmd.MarkFlagsMutuallyExclusive("height", "lossless")
	compressCmd.MarkFlagsMutuallyExclusive("compat", "lossless")

	addOutputFlag(compressCmd)
	addJobsFlag(compressCmd)
}
