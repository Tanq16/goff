package cmd

import (
	"context"
	"os"

	"github.com/spf13/cobra"

	"github.com/Tanq16/goff/internal/ops"
	"github.com/Tanq16/goff/internal/probe"
	"github.com/Tanq16/goff/utils"
)

var concatFlags struct {
	reencode bool
}

var concatCmd = &cobra.Command{
	Use:   "concat <files...>",
	Short: "Join several clips end to end into one file",
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			utils.PrintFatal("concat needs at least two input files", nil)
		}

		res, listFile, err := ops.BuildMultiConcat(args, ops.MultiConcatOpts{
			Inputs:   args,
			Reencode: concatFlags.reencode,
		})
		if err != nil {
			utils.PrintFatal("failed to build concat arguments", err)
		}
		if listFile != "" {
			defer os.Remove(listFile)
		}

		var totalSec float64
		for _, in := range args {
			if p, probeErr := probe.RunProbe(context.Background(), in); probeErr == nil {
				totalSec += p.TotalDuration()
			}
		}

		runComposed("concat", args[0], res, totalSec)
	},
}

func init() {
	concatCmd.Flags().BoolVar(&concatFlags.reencode, "reencode", false, "Re-encode to a uniform stream instead of copying (needed for mismatched sources)")

	rootCmd.AddCommand(concatCmd)
}
