package cmd

import (
	"github.com/spf13/cobra"

	"github.com/Tanq16/goff/internal/ops"
	"github.com/Tanq16/goff/internal/probe"
)

var convertFlags struct {
	to       string
	bitrate  string
	rate     int
	channels int
}

var convertCmd = &cobra.Command{
	Use:     "convert <files...>",
	GroupID: "audio",
	Short:   "Transcode audio between formats, sample rates, and channel counts",
	Example: `  goff convert song.wav --to opus --rate 48000
  goff convert *.flac --to mp3 --bitrate 320k`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runFiles("convert", args, func(input string, p *probe.ProbeResult) (*ops.OpResult, error) {
			return ops.BuildAudioConvert(input, p, ops.AudioConvertOpts{
				Format:     convertFlags.to,
				Bitrate:    convertFlags.bitrate,
				SampleRate: convertFlags.rate,
				Channels:   convertFlags.channels,
			})
		})
	},
}

func init() {
	convertCmd.Flags().Var(newEnum(&convertFlags.to, "mp3", "mp3", "aac", "m4a", "flac", "opus", "wav", "ogg"), "to", "Audio format")
	convertCmd.Flags().Var(newBitrate(&convertFlags.bitrate), "bitrate", "Audio bitrate, e.g. 256k (format default when unset)")
	convertCmd.Flags().Var(newBoundedInt(&convertFlags.rate, 0, 8000, 192000), "rate", "Sample rate in Hz, e.g. 48000 (source rate when unset)")
	convertCmd.Flags().Var(newBoundedInt(&convertFlags.channels, 0, 1, 8), "channels", "Channel count, 2 downmixes to stereo (source layout when unset)")

	addOutputFlag(convertCmd)
	addJobsFlag(convertCmd)
}
