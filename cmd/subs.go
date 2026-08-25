package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/Tanq16/goff/internal/ops"
	"github.com/Tanq16/goff/internal/probe"
	"github.com/Tanq16/goff/utils"
)

var subsFlags struct {
	file string
	burn bool
}

var subsCmd = &cobra.Command{
	Use:     "subs <video>",
	GroupID: "combine",
	Short:   "Add a subtitle track to a video, embedded or burned in",
	Example: `  goff subs film.mkv --file film.srt
  goff subs clip.mp4 --file clip.srt --burn`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		video := args[0]

		res, err := ops.BuildMultiMuxSubs(ops.MultiMuxSubsOpts{
			VideoInput: video,
			SubsInput:  subsFlags.file,
			Hardburn:   subsFlags.burn,
		})
		if err != nil {
			utils.PrintFatal("failed to build subtitle arguments", err)
		}

		var totalSec float64
		if p, probeErr := probe.RunProbe(context.Background(), video); probeErr == nil {
			totalSec = p.TotalDuration()
		}

		runComposed("subs", video, res, totalSec)
	},
}

func init() {
	subsCmd.Flags().StringVar(&subsFlags.file, "file", "", "Subtitle file to add (required)")
	subsCmd.MarkFlagRequired("file")
	subsCmd.Flags().BoolVar(&subsFlags.burn, "burn", false, "Render subtitles into the picture instead of embedding a track")
}
