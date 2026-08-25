package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/Tanq16/goff/internal/ops"
	"github.com/Tanq16/goff/internal/probe"
	"github.com/Tanq16/goff/utils"
)

var concatFlags struct {
	reencode bool
}

var concatCmd = &cobra.Command{
	Use:     "concat <files...>",
	GroupID: "combine",
	Short:   "Join several clips end to end into one file",
	Long: `Join several clips end to end into one file.

To layer audio tracks so they play at the same time, use mix.`,
	Example: `  goff concat part1.mp4 part2.mp4 part3.mp4
  goff concat a.mp4 b.mp4 --reencode`,
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		totalSec := checkConcatInputs(args)

		res, listFile, err := ops.BuildMultiConcat(args, ops.MultiConcatOpts{
			Inputs:   args,
			Reencode: concatFlags.reencode,
		})
		if err != nil {
			utils.PrintFatal("failed to build concat arguments", err)
		}
		cleanup := func() {}
		if listFile != "" {
			cleanup = func() { os.Remove(listFile) }
		}

		runComposed("concat", args[0], res, totalSec, cleanup)
	},
}

func checkConcatInputs(inputs []string) float64 {
	var totalSec float64
	withVideo, withAudio := 0, 0

	for _, in := range inputs {
		p, err := probe.RunProbe(context.Background(), in)
		if err != nil {
			utils.PrintFatal(fmt.Sprintf("cannot read %s", filepath.Base(in)), err)
		}
		totalSec += p.TotalDuration()
		if p.IsVideo() {
			withVideo++
		}
		if len(p.AudioStreams()) > 0 {
			withAudio++
		}
	}

	if withVideo != 0 && withVideo != len(inputs) {
		utils.PrintFatal("concat needs every input to be the same kind; these mix video files with audio-only files", nil)
	}
	if concatFlags.reencode && (withVideo != len(inputs) || withAudio != len(inputs)) {
		utils.PrintFatal("--reencode needs every input to carry both a video and an audio stream", nil)
	}

	return totalSec
}

func init() {
	concatCmd.Flags().BoolVar(&concatFlags.reencode, "reencode", false, "Re-encode to a uniform stream instead of copying (needed for mismatched sources)")
}
