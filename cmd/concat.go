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
		var probes []*probe.ProbeResult
		var totalSec float64
		for _, in := range args {
			p, err := probe.RunProbe(context.Background(), in)
			if err != nil {
				utils.PrintFatal(fmt.Sprintf("cannot read %s", filepath.Base(in)), err)
			}
			probes = append(probes, p)
			totalSec += p.TotalDuration()
		}
		if err := ops.CheckConcatInputs(probes, concatFlags.reencode); err != nil {
			utils.PrintFatal(err.Error(), nil)
		}

		res, err := ops.BuildMultiConcat(args, ops.MultiConcatOpts{
			Reencode: concatFlags.reencode,
		})
		if err != nil {
			utils.PrintFatal("failed to build concat arguments", err)
		}

		runComposed("concat", args[0], res, totalSec)
	},
}

func init() {
	concatCmd.Flags().BoolVar(&concatFlags.reencode, "reencode", false, "Re-encode to a uniform stream instead of copying (needed for mismatched sources)")

	addOutputFlag(concatCmd)
}
