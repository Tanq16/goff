package cmd

import (
	"context"
	"fmt"
	"slices"

	"github.com/spf13/cobra"

	"github.com/Tanq16/goff/internal/ops"
	"github.com/Tanq16/goff/internal/probe"
	"github.com/Tanq16/goff/utils"
)

var mixFlags struct {
	to      string
	bitrate string
	fit     string
	sources []ops.AudioSource
}

var mixCmd = &cobra.Command{
	Use:     "mix <audio...>",
	GroupID: "audio",
	Short:   "Layer several audio tracks so they play at the same time",
	Long: `Layer several audio tracks so they play at the same time.

Every input starts at 0 unless it carries an offset. Append ":at=<time>" to delay
an input and ":vol=<factor>" to change its level, e.g. music.mp3:at=00:00:05:vol=0.3.

To join clips end to end instead of layering them, use concat.`,
	Example: `  goff mix voice.wav music.mp3
  goff mix voice.wav music.mp3:at=5:vol=0.3 --to m4a
  goff mix a.mp3 b.mp3 --fit shortest`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return fmt.Errorf("mix needs at least two audio inputs")
		}
		mixFlags.sources = mixFlags.sources[:0]
		for _, arg := range args {
			in, err := parseMediaInput(arg)
			if err != nil {
				return fmt.Errorf("%s: %w", arg, err)
			}
			mixFlags.sources = append(mixFlags.sources, ops.AudioSource{
				Path:    in.Path,
				DelayMS: in.DelayMS,
				Volume:  in.Volume,
			})
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		res, err := ops.BuildMultiMix(ops.MultiMixOpts{
			Sources: mixFlags.sources,
			Fit:     mixFlags.fit,
			Format:  mixFlags.to,
			Bitrate: mixFlags.bitrate,
		})
		if err != nil {
			utils.PrintFatal("failed to build mix arguments", err)
		}

		runComposed("mix", mixFlags.sources[0].Path, res, mixDuration(mixFlags.sources, mixFlags.fit))
	},
}

func mixDuration(sources []ops.AudioSource, fit string) float64 {
	var ends []float64
	for _, s := range sources {
		p, err := probe.RunProbe(context.Background(), s.Path)
		if err != nil {
			continue
		}
		ends = append(ends, float64(s.DelayMS)/1000.0+p.TotalDuration())
	}
	if len(ends) == 0 {
		return 0
	}
	if fit == "shortest" {
		return slices.Min(ends)
	}
	return slices.Max(ends)
}

func init() {
	mixCmd.Flags().Var(newEnum(&mixFlags.to, "mp3", "mp3", "aac", "m4a", "flac", "opus", "wav", "ogg"), "to", "Output audio format")
	mixCmd.Flags().Var(newBitrate(&mixFlags.bitrate), "bitrate", "Audio bitrate, e.g. 256k (format default when unset)")
	mixCmd.Flags().Var(newEnum(&mixFlags.fit, "longest", "longest", "shortest"), "fit", "Output length follows the longest or the shortest input")
}
