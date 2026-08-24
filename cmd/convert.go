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
	Use:   "convert <files...>",
	Short: "Transcode audio between formats, sample rates, and channel counts",
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		requireOneOf("to", convertFlags.to, "mp3", "aac", "m4a", "flac", "opus", "wav", "ogg")

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
	convertCmd.Flags().StringVar(&convertFlags.to, "to", "mp3", "Audio format: mp3, aac, flac, opus, wav, ogg")
	convertCmd.Flags().StringVar(&convertFlags.bitrate, "bitrate", "", "Audio bitrate, e.g. 256k (format default when unset)")
	convertCmd.Flags().IntVar(&convertFlags.rate, "rate", 0, "Sample rate in Hz, e.g. 48000")
	convertCmd.Flags().IntVar(&convertFlags.channels, "channels", 0, "Channel count, 2 downmixes to stereo")

	rootCmd.AddCommand(convertCmd)
}
