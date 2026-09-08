package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Tanq16/goff/internal/ops"
	"github.com/Tanq16/goff/internal/probe"
	"github.com/Tanq16/goff/utils"
)

var muxFlags struct {
	audio   []mediaInput
	mode    string
	fit     string
	bitrate string
}

var muxCmd = &cobra.Command{
	Use:     "mux <video>",
	GroupID: "combine",
	Short:   "Put external audio tracks onto a video",
	Long: `Put external audio tracks onto a video.

--audio is repeatable, and every track starts at 0 unless it carries an offset.
Append ":at=<time>" to delay a track and ":vol=<factor>" to change its level,
e.g. --audio music.mp3:at=00:00:05:vol=0.3.

--audio-mode decides what happens to the audio the video already has: "mix"
layers it with the new tracks into one, "replace" drops it, and "separate"
keeps every track selectable instead of combining them.`,
	Example: `  goff mux talk.mp4 --audio music.mp3:vol=0.3
  goff mux talk.mp4 --audio dub.m4a --audio-mode replace
  goff mux film.mkv --audio en.m4a --audio fr.m4a --audio-mode separate`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		video := args[0]
		rejectInertMuxOptions()

		p, err := probe.RunProbe(context.Background(), video)
		if err != nil {
			utils.PrintFatal("cannot read "+video, err)
		}
		if !p.IsVideo() {
			utils.PrintFatal(video+" has no video stream; use mix to combine audio files", nil)
		}

		sources := make([]ops.AudioSource, 0, len(muxFlags.audio))
		for _, in := range muxFlags.audio {
			sources = append(sources, ops.AudioSource{Path: in.Path, DelayMS: in.DelayMS, Volume: in.Volume})
		}

		res, err := ops.BuildMultiMuxAudio(p, ops.MultiMuxAudioOpts{
			VideoInput:   video,
			Sources:      sources,
			Mode:         muxFlags.mode,
			Fit:          muxFlags.fit,
			AudioBitrate: muxFlags.bitrate,
		})
		if err != nil {
			utils.PrintFatal("failed to build mux arguments", err)
		}

		runComposed("mux", video, res, p.TotalDuration())
	},
}

func rejectInertMuxOptions() {
	if muxFlags.mode != ops.MuxModeSeparate {
		return
	}
	for _, in := range muxFlags.audio {
		if in.DelayMS != 0 || in.Volume != 1.0 {
			utils.PrintFatal(fmt.Sprintf("at= and vol= do not apply to --audio-mode separate (%s)", in.Path), nil)
		}
	}
}

func init() {
	muxCmd.Flags().Var(newMediaInputs(&muxFlags.audio), "audio", "Audio file to add, repeatable (required)")
	muxCmd.MarkFlagRequired("audio")
	muxCmd.Flags().Var(newEnum(&muxFlags.mode, ops.MuxModeMix, ops.MuxModeMix, ops.MuxModeReplace, ops.MuxModeSeparate), "audio-mode", "What happens to the audio the video already has")
	muxCmd.Flags().Var(newEnum(&muxFlags.fit, "video", "video", "longest", "shortest"), "fit", "Output length follows the video, the longest input, or the shortest")
	muxCmd.Flags().Var(newBitrate(&muxFlags.bitrate), "bitrate", "Audio bitrate for the muxed track (default 192k)")

	addOutputFlag(muxCmd)
}
