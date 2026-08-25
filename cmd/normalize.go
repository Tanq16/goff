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
	Use:     "normalize <files...>",
	GroupID: "audio",
	Short:   "Apply EBU R128 loudness normalization to audio or video",
	Example: `  goff normalize podcast.wav --to mp3
  goff normalize lecture.mp4 --lufs -14`,
	Args: cobra.MinimumNArgs(1),
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
	normalizeCmd.Flags().Var(newBoundedFloat(&normalizeFlags.lufs, -16.0, -70.0, -5.0, "between -70 and -5"), "lufs", "Integrated loudness target in LUFS")
	normalizeCmd.Flags().Var(newBoundedFloat(&normalizeFlags.peak, -1.5, -9.0, 0.0, "between -9 and 0"), "peak", "True peak ceiling in dBTP")
	normalizeCmd.Flags().Var(newBoundedFloat(&normalizeFlags.lra, 11.0, 1.0, 50.0, "between 1 and 50"), "range", "Loudness range target")
	normalizeCmd.Flags().Var(newEnum(&normalizeFlags.to, "", "mp3", "m4a", "aac", "flac", "wav", "ogg", "opus", "mp4", "mkv", "mov"), "to", "Output format (keeps the source container when unset)")
}
