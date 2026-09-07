package cmd

import (
	"github.com/spf13/cobra"

	"github.com/Tanq16/goff/internal/ops"
	"github.com/Tanq16/goff/internal/probe"
)

var compressFlags struct {
	codec         string
	crf           int
	height        int
	size          string
	lossless      bool
	preset        string
	audioBitrate  string
	audioRate     int
	keep10Bit     bool
	keepHiFiAudio bool
	copyVideo     bool
	audioTracks   probe.TrackSelector
	subTracks     probe.TrackSelector
	sizeBudgetMB  float64
}

var compressCmd = &cobra.Command{
	Use:     "compress <files...>",
	GroupID: "video",
	Short:   "Re-encode video smaller, normalized for playback anywhere",
	Example: `  goff compress movie.mkv
  goff compress movie.mkv --codec av1 --crf 28
  goff compress clip.mp4 --size 25MB
  goff compress master.mov --lossless
  goff compress movie.mkv --keep-10bit --keep-hifi-audio
  goff compress drifting.mp4 --copy-video
  goff compress rip.mkv --audio-track eng --sub-track none`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runFiles("compress", args, func(input string, p *probe.ProbeResult) (*ops.OpResult, error) {
			return ops.BuildVideoOptimize(input, p, ops.VideoOptimizeOpts{
				Codec:         compressFlags.codec,
				CRF:           compressFlags.crf,
				Lossless:      compressFlags.lossless,
				MaxHeight:     compressFlags.height,
				TargetSizeMB:  compressFlags.sizeBudgetMB,
				CustomSuffix:  compressFlags.size,
				Preset:        compressFlags.preset,
				AudioBitrate:  compressFlags.audioBitrate,
				AudioRate:     compressFlags.audioRate,
				Keep10Bit:     compressFlags.keep10Bit,
				KeepHiFiAudio: compressFlags.keepHiFiAudio,
				CopyVideo:     compressFlags.copyVideo,
				AudioTracks:   compressFlags.audioTracks,
				SubTracks:     compressFlags.subTracks,
			})
		})
	},
}

func init() {
	compressCmd.Flags().VarP(newEnum(&compressFlags.codec, "hevc", "hevc", "av1", "h264"), "codec", "c", "Video codec")
	compressCmd.Flags().Var(newBoundedInt(&compressFlags.crf, 0, 1, 63), "crf", "Quality factor, lower is better (codec default when unset)")
	compressCmd.Flags().Var(newBoundedInt(&compressFlags.height, 1080, 144, 4320), "height", "Maximum output height in pixels")
	compressCmd.Flags().Var(newSize(&compressFlags.size, &compressFlags.sizeBudgetMB), "size", "Target file size budget, e.g. 25MB (overrides --crf)")
	compressCmd.Flags().BoolVar(&compressFlags.lossless, "lossless", false, "Re-encode with no video quality loss, keeping the source resolution, bit depth, and channel layout")
	compressCmd.Flags().StringVar(&compressFlags.preset, "preset", "", "Encoder speed preset, e.g. slow, or 6 for av1 (codec default when unset)")
	compressCmd.Flags().Var(newBitrate(&compressFlags.audioBitrate), "audio-bitrate", "Audio bitrate, e.g. 192k (128k when unset)")
	compressCmd.Flags().Var(newSampleRate(&compressFlags.audioRate, 48000), "audio-rate", "Audio sample rate")
	compressCmd.Flags().BoolVar(&compressFlags.keep10Bit, "keep-10bit", false, "Keep the source bit depth and skip HDR tone-mapping")
	compressCmd.Flags().BoolVar(&compressFlags.keepHiFiAudio, "keep-hifi-audio", false, "Keep the source channel layout instead of downmixing to stereo")
	compressCmd.Flags().BoolVar(&compressFlags.copyVideo, "copy-video", false, "Copy the video stream untouched and re-encode only audio, to repair sync in seconds")
	compressCmd.Flags().Var(newTrack(&compressFlags.audioTracks, probe.AllTracks()), "audio-track", "Audio tracks to keep, by stream index or language")
	compressCmd.Flags().Var(newTrack(&compressFlags.subTracks, probe.AllTracks()), "sub-track", "Subtitle tracks to keep, by stream index or language")

	compressCmd.MarkFlagsMutuallyExclusive("crf", "size", "lossless")
	compressCmd.MarkFlagsMutuallyExclusive("height", "lossless")
	for _, flag := range []string{"codec", "crf", "height", "size", "lossless", "preset", "keep-10bit"} {
		compressCmd.MarkFlagsMutuallyExclusive("copy-video", flag)
	}

	addOutputFlag(compressCmd)
	addJobsFlag(compressCmd)
}
