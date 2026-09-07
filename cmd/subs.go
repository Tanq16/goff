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
	language  string
	makeDflt  bool
}

var subsCmd = &cobra.Command{
	Use:     "subs <video>",
	GroupID: "combine",
	Short:   "Add a subtitle track to a video",
	Example: `  goff subs film.mkv --subtitles film.srt
  goff subs film.mp4 --subtitles film.eng.srt --lang eng --default`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		video := args[0]

		p, err := probe.RunProbe(context.Background(), video)
		if err != nil {
			utils.PrintFatal(fmt.Sprintf("cannot read %s", filepath.Base(video)), err)
		}

		res, err := ops.BuildMultiMuxSubs(p, ops.MultiMuxSubsOpts{
			VideoInput: video,
			SubsInput:  subsFlags.subtitles,
			Language:   subsFlags.language,
			Default:    subsFlags.makeDflt,
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
	subsCmd.Flags().StringVar(&subsFlags.language, "lang", "", "Language code to tag the added track with, e.g. eng")
	subsCmd.Flags().BoolVar(&subsFlags.makeDflt, "default", false, "Mark the added track as the default subtitle track")

	addOutputFlag(subsCmd)
}
