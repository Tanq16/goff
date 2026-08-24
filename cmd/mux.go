package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/Tanq16/goff/internal/ops"
	"github.com/Tanq16/goff/internal/probe"
	"github.com/Tanq16/goff/utils"
)

var muxFlags struct {
	audio   string
	subs    string
	replace bool
	burn    bool
}

var muxCmd = &cobra.Command{
	Use:   "mux <video>",
	Short: "Add an external audio or subtitle track to a video",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		video := args[0]
		if muxFlags.audio == "" && muxFlags.subs == "" {
			utils.PrintFatal("mux needs --audio or --subs", nil)
		}

		var res *ops.OpResult
		var err error
		if muxFlags.audio != "" {
			res, err = ops.BuildMultiMuxAudio(ops.MultiMuxAudioOpts{
				VideoInput:      video,
				AudioInput:      muxFlags.audio,
				ReplaceOriginal: muxFlags.replace,
			})
		} else {
			res, err = ops.BuildMultiMuxSubs(ops.MultiMuxSubsOpts{
				VideoInput: video,
				SubsInput:  muxFlags.subs,
				Hardburn:   muxFlags.burn,
			})
		}
		if err != nil {
			utils.PrintFatal("failed to build mux arguments", err)
		}

		var totalSec float64
		if p, probeErr := probe.RunProbe(context.Background(), video); probeErr == nil {
			totalSec = p.TotalDuration()
		}

		runComposed("mux", video, res, totalSec)
	},
}

func init() {
	muxCmd.Flags().StringVar(&muxFlags.audio, "audio", "", "External audio file to add")
	muxCmd.Flags().StringVar(&muxFlags.subs, "subs", "", "External subtitle file to add")
	muxCmd.Flags().BoolVar(&muxFlags.replace, "replace", false, "Replace the existing audio track instead of adding one")
	muxCmd.Flags().BoolVar(&muxFlags.burn, "burn", false, "Burn subtitles into the picture instead of embedding a track")
	muxCmd.MarkFlagsMutuallyExclusive("audio", "subs")
	muxCmd.MarkFlagsMutuallyExclusive("replace", "burn")

	rootCmd.AddCommand(muxCmd)
}
