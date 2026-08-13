package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"

	"github.com/Tanq16/goff/internal/engine"
	"github.com/Tanq16/goff/internal/presets"
	"github.com/Tanq16/goff/internal/probe"
	"github.com/Tanq16/goff/utils"
)

var batchFlags struct {
	preset    string
	workers   int
	overwrite bool
}

type batchResult struct {
	Input    string
	Output   string
	OrigSize int64
	OutSize  int64
	Duration time.Duration
	Err      error
}

var batchCmd = &cobra.Command{
	Use:   "batch [files...]",
	Short: "Batch process multiple media files with a selected preset",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			utils.PrintFatal("at least one input file required for batch processing", nil)
		}

		if batchFlags.preset == "" {
			batchFlags.preset = "web-optimize"
		}

		preset, ok := presets.Get(batchFlags.preset)
		if !ok {
			utils.PrintFatal(fmt.Sprintf("unknown preset %q. Available: %s", batchFlags.preset, presets.AvailablePresetIDs()), nil)
		}

		workers := batchFlags.workers
		if workers <= 0 {
			workers = 2
		}

		utils.PrintInfo(fmt.Sprintf("Starting batch processing %d files (preset: %s, workers: %d)", len(args), preset.ID, workers))

		g, ctx := errgroup.WithContext(context.Background())
		g.SetLimit(workers)

		var mu sync.Mutex
		var results []batchResult

		for _, inputFile := range args {
			file := inputFile
			g.Go(func() error {
				start := time.Now()
				p, err := probe.RunProbe(ctx, file)
				if err != nil {
					mu.Lock()
					results = append(results, batchResult{Input: file, Err: err})
					mu.Unlock()
					utils.PrintIndentedError(filepath.Base(file), err)
					return nil
				}

				res, err := preset.Build(file, p)
				if err != nil {
					mu.Lock()
					results = append(results, batchResult{Input: file, Err: err})
					mu.Unlock()
					utils.PrintIndentedError(filepath.Base(file), err)
					return nil
				}

				outPath, err := engine.ResolveOutputName(file, res.Suffix, res.TargetExt, "", batchFlags.overwrite)
				if err != nil {
					mu.Lock()
					results = append(results, batchResult{Input: file, Err: err})
					mu.Unlock()
					utils.PrintIndentedError(filepath.Base(file), err)
					return nil
				}

				fullArgs := append(res.Args, outPath)
				totalSec := p.TotalDuration()

				err = engine.RunFFmpeg(ctx, fullArgs, totalSec, nil)
				elapsed := time.Since(start)

				var outSize int64
				if err == nil {
					if fi, statErr := os.Stat(outPath); statErr == nil {
						outSize = fi.Size()
					}
					utils.PrintIndentedSuccess(fmt.Sprintf("%s → %s (%s)", filepath.Base(file), filepath.Base(outPath), probe.FormatBytes(outSize)))
				} else {
					utils.PrintIndentedError(filepath.Base(file), err)
				}

				mu.Lock()
				results = append(results, batchResult{
					Input:    file,
					Output:   outPath,
					OrigSize: p.Format.Size(),
					OutSize:  outSize,
					Duration: elapsed,
					Err:      err,
				})
				mu.Unlock()

				return nil
			})
		}

		_ = g.Wait()

		utils.PrintSuccess("Batch processing completed")

		var headers = []string{"File", "Status", "Original", "Output", "Savings", "Time"}
		var rows [][]string

		for _, r := range results {
			status := "OK"
			if r.Err != nil {
				status = "FAILED"
			}
			origStr := probe.FormatBytes(r.OrigSize)
			outStr := probe.FormatBytes(r.OutSize)
			savingsStr := "-"
			if r.Err == nil && r.OrigSize > 0 && r.OutSize > 0 {
				ratio := float64(r.OrigSize-r.OutSize) / float64(r.OrigSize) * 100.0
				savingsStr = fmt.Sprintf("%.1f%%", ratio)
			}
			rows = append(rows, []string{
				filepath.Base(r.Input),
				status,
				origStr,
				outStr,
				savingsStr,
				r.Duration.Round(time.Millisecond * 100).String(),
			})
		}

		utils.PrintTable(headers, rows)
	},
}

func init() {
	batchCmd.Flags().StringVarP(&batchFlags.preset, "preset", "p", "web-optimize", "Preset to apply across all files")
	batchCmd.Flags().IntVarP(&batchFlags.workers, "workers", "j", 2, "Concurrent encoding worker limit")
	batchCmd.Flags().BoolVarP(&batchFlags.overwrite, "yes", "y", false, "Allow overwriting existing files")

	rootCmd.AddCommand(batchCmd)
}
