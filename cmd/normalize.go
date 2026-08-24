package cmd

import (
	"github.com/spf13/cobra"

	"github.com/Tanq16/goff/internal/ops"
	"github.com/Tanq16/goff/internal/probe"
)

var normalizeFlags struct {
	lufs float64
	peak float64
	lra  float64
	to   string
}

var normalizeCmd = &cobra.Command{
	Use:   "normalize <files...>",
	Short: "Apply EBU R128 loudness normalization to audio or video",
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		runFiles("normalize", args, func(input string, p *probe.ProbeResult) (*ops.OpResult, error) {
			return ops.BuildAudioLoudnorm(input, p, ops.AudioLoudnormOpts{
				IntegratedLoudness: normalizeFlags.lufs,
				TruePeak:           normalizeFlags.peak,
				LoudnessRange:      normalizeFlags.lra,
				OutputExt:          normalizeFlags.to,
			})
		})
	},
}

func init() {
	normalizeCmd.Flags().Float64Var(&normalizeFlags.lufs, "lufs", -16.0, "Integrated loudness target in LUFS")
	normalizeCmd.Flags().Float64Var(&normalizeFlags.peak, "peak", -1.5, "True peak ceiling in dBTP")
	normalizeCmd.Flags().Float64Var(&normalizeFlags.lra, "range", 11.0, "Loudness range target")
	normalizeCmd.Flags().StringVar(&normalizeFlags.to, "to", "", "Output format, e.g. mp3 (keeps the source container when unset)")

	rootCmd.AddCommand(normalizeCmd)
}
