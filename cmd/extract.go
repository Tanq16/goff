package cmd

import (
	"cmp"
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/Tanq16/goff/internal/ops"
	"github.com/Tanq16/goff/internal/probe"
)

var extractFlags struct {
	to          string
	bitrate     string
	rate        int
	channels    int
	audioTracks probe.TrackSelector
	subTracks   probe.TrackSelector
}

var extractCmd = &cobra.Command{
	Use:     "extract <files...>",
	GroupID: "audio",
	Short:   "Pull audio or subtitle streams out of a video into their own files",
	Example: `  goff extract talk.mp4 --to mp3 --bitrate 320k
  goff extract talk.mp4 --to wav --rate 16000 --channels 1
  goff extract film.mkv --to vtt`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runFiles("extract", args, func(input string, p *probe.ProbeResult) (*ops.OpResult, error) {
			streams, ext, err := extractStreams(input, p)
			if err != nil {
				return nil, err
			}
			if sharedFlags.output != "" && len(streams) > 1 {
				return nil, fmt.Errorf("--output names a single destination, and %s has %d matching streams", filepath.Base(input), len(streams))
			}

			targets := make([]ops.ExtractTarget, 0, len(streams))
			for _, s := range streams {
				suffix := cmp.Or(s.Language(), defaultExtractSuffix(s.CodecType))
				path, err := claimOutput(input, suffix, ext)
				if err != nil {
					return nil, err
				}
				targets = append(targets, ops.ExtractTarget{StreamIndex: s.Index, Path: path})
			}

			return ops.BuildExtract(input, targets, ops.VideoExtractOpts{
				Format:     extractFlags.to,
				Bitrate:    extractFlags.bitrate,
				SampleRate: extractFlags.rate,
				Channels:   extractFlags.channels,
			})
		})
	},
}

func extractStreams(input string, p *probe.ProbeResult) ([]probe.StreamInfo, string, error) {
	ext, subtitles := ops.ExtractFormat(extractFlags.to)
	if !subtitles {
		streams, err := p.SelectAudio(extractFlags.audioTracks)
		if err != nil {
			return nil, "", err
		}
		if len(streams) == 0 {
			return nil, "", fmt.Errorf("%s has no audio stream", filepath.Base(input))
		}
		return streams, ext, nil
	}

	selected, err := p.SelectSubtitles(extractFlags.subTracks)
	if err != nil {
		return nil, "", err
	}
	var text []probe.StreamInfo
	for _, s := range selected {
		if probe.IsTextSubtitle(s.CodecName) {
			text = append(text, s)
		}
	}
	if len(text) == 0 {
		return nil, "", fmt.Errorf("%s has no text subtitle stream to extract", filepath.Base(input))
	}
	return text, ext, nil
}

func defaultExtractSuffix(codecType string) string {
	if codecType == "subtitle" {
		return "subs"
	}
	return "audio"
}

func init() {
	extractCmd.Flags().Var(newEnum(&extractFlags.to, "mp3", "mp3", "aac", "m4a", "flac", "opus", "wav", "srt", "vtt"), "to", "Output format")
	extractCmd.Flags().Var(newBitrate(&extractFlags.bitrate), "bitrate", "Audio bitrate, e.g. 320k (format default when unset)")
	extractCmd.Flags().Var(newBoundedInt(&extractFlags.rate, 0, 8000, 192000), "rate", "Audio sample rate in Hz (source rate when unset)")
	extractCmd.Flags().Var(newBoundedInt(&extractFlags.channels, 0, 1, 8), "channels", "Audio channel count (source layout when unset)")
	extractCmd.Flags().Var(newTrack(&extractFlags.audioTracks, probe.AllTracks()), "audio-track", "Audio tracks to pull, by stream index or language")
	extractCmd.Flags().Var(newTrack(&extractFlags.subTracks, probe.AllTracks()), "sub-track", "Subtitle tracks to pull, by stream index or language")

	addOutputFlag(extractCmd)
	addJobsFlag(extractCmd)
}
