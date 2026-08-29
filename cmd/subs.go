package cmd

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/Tanq16/goff/internal/ops"
	"github.com/Tanq16/goff/internal/probe"
	"github.com/Tanq16/goff/utils"
)

var subsFlags struct {
	subtitles string
	burn      bool
}

var subsCmd = &cobra.Command{
	Use:     "subs <video>",
	GroupID: "combine",
	Short:   "Add a subtitle track to a video, embedded or burned in",
	Example: `  goff subs film.mkv --subtitles film.srt
  goff subs clip.mp4 --subtitles clip.srt --burn`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		video := args[0]

		p, err := probe.RunProbe(context.Background(), video)
		if err != nil {
			utils.PrintFatal(fmt.Sprintf("cannot read %s", filepath.Base(video)), err)
		}

		res, err := ops.BuildMultiMuxSubs(ops.MultiMuxSubsOpts{
			VideoInput: video,
			SubsInput:  subsFlags.subtitles,
			Hardburn:   subsFlags.burn,
		})
		if err != nil {
			utils.PrintFatal("failed to build subtitle arguments", err)
		}

		runComposed("subs", video, res, p.TotalDuration())
	},
}

func init() {
	subsCmd.Flags().StringVar(&subsFlags.subtitles, "subtitles", "", "Subtitle file to add (required)")
	subsCmd.MarkFlagRequired("subtitles")
	subsCmd.Flags().BoolVar(&subsFlags.burn, "burn", false, "Render subtitles into the picture instead of embedding a track")

	addOutputFlag(subsCmd)
}
