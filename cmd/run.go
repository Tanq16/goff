package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/Tanq16/goff/internal/engine"
	"github.com/Tanq16/goff/internal/ops"
	"github.com/Tanq16/goff/internal/probe"
	"github.com/Tanq16/goff/utils"
)

type buildFunc func(input string, p *probe.ProbeResult) (*ops.OpResult, error)

type fileResult struct {
	Input    string
	Output   string
	OrigSize int64
	OutSize  int64
	Duration time.Duration
	Err      error
}

func requireInputs(verb string, args []string) {
	if len(args) == 0 {
		utils.PrintFatal(fmt.Sprintf("%s needs at least one input file", verb), nil)
	}
	if rootFlags.output != "" && len(args) > 1 {
		utils.PrintFatal("--output names a single destination; drop it to write one output per input", nil)
	}
}

var claimed struct {
	sync.Mutex
	paths map[string]bool
}

func claimOutput(input, suffix, targetExt string) (string, error) {
	if rootFlags.overwrite || rootFlags.output != "" {
		return engine.ResolveOutputName(input, suffix, targetExt, rootFlags.output, rootFlags.overwrite)
	}

	claimed.Lock()
	defer claimed.Unlock()
	if claimed.paths == nil {
		claimed.paths = make(map[string]bool)
	}
	for attempt := range 1000 {
		candidate := suffix
		if attempt > 0 {
			candidate = fmt.Sprintf("%s.%d", suffix, attempt)
		}
		out, err := engine.ResolveOutputName(input, candidate, targetExt, "", false)
		if err != nil {
			return "", err
		}
		if !claimed.paths[out] {
			claimed.paths[out] = true
			return out, nil
		}
	}
	return "", fmt.Errorf("unable to find a free output name for %q", input)
}

func prepare(ctx context.Context, input string, build buildFunc) (*ops.OpResult, *probe.ProbeResult, string, error) {
	p, err := probe.RunProbe(ctx, input)
	if err != nil {
		return nil, nil, "", err
	}
	res, err := build(input, p)
	if err != nil {
		return nil, nil, "", err
	}
	outPath, err := claimOutput(input, res.Suffix, res.TargetExt)
	if err != nil {
		return nil, nil, "", err
	}
	return res, p, outPath, nil
}

func encodeWithProgress(ctx context.Context, verb string, label string, args []string, totalSec float64) (time.Duration, error) {
	utils.PrintRunning(fmt.Sprintf("%s %s", verb, label))

	var curPercent atomic.Int32
	var printed atomic.Bool
	done := make(chan struct{})
	var bar sync.WaitGroup

	bar.Go(func() {
		t := time.NewTicker(250 * time.Millisecond)
		defer t.Stop()
		firstTick := true
		for {
			select {
			case <-done:
				return
			case <-t.C:
				if !firstTick {
					utils.ClearPreviousLine()
				}
				firstTick = false
				printed.Store(true)
				utils.PrintProgress(label, int(curPercent.Load()))
			}
		}
	})

	start := time.Now()
	err := engine.RunFFmpeg(ctx, args, totalSec, func(prog engine.ProgressUpdate) {
		curPercent.Store(int32(prog.Percent))
	})
	elapsed := time.Since(start)

	close(done)
	bar.Wait()
	if printed.Load() {
		utils.ClearPreviousLine()
	}
	utils.ClearLines(1)

	return elapsed, err
}

func fileSize(path string) int64 {
	fi, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return fi.Size()
}

func runFiles(verb string, files []string, build buildFunc) {
	requireInputs(verb, files)
	if len(files) == 1 {
		runSingle(verb, files[0], build)
		return
	}
	runMany(verb, files, build)
}

func runSingle(verb string, input string, build buildFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	res, p, outPath, err := prepare(ctx, input, build)
	if err != nil {
		utils.PrintFatal(fmt.Sprintf("cannot %s %s", verb, filepath.Base(input)), err)
	}

	label := filepath.Base(input)
	elapsed, err := encodeWithProgress(ctx, verb, label, append(res.Args, outPath), p.TotalDuration())
	if err != nil {
		utils.PrintFatal(fmt.Sprintf("%s failed for %s", verb, label), err)
	}

	utils.PrintSuccess(fmt.Sprintf("%s → %s (%s, %s → %s)",
		label, outPath, elapsed.Round(100*time.Millisecond),
		probe.FormatBytes(p.Format.Size()), probe.FormatBytes(fileSize(outPath))))
}

func runComposed(verb string, namingInput string, res *ops.OpResult, totalSec float64) {
	outPath, err := engine.ResolveOutputName(namingInput, res.Suffix, res.TargetExt, rootFlags.output, rootFlags.overwrite)
	if err != nil {
		utils.PrintFatal("failed to resolve output path", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	label := filepath.Base(namingInput)
	elapsed, err := encodeWithProgress(ctx, verb, label, append(res.Args, outPath), totalSec)
	if err != nil {
		utils.PrintFatal(fmt.Sprintf("%s failed", verb), err)
	}

	utils.PrintSuccess(fmt.Sprintf("%s → %s (%s, %s)",
		label, outPath, elapsed.Round(100*time.Millisecond), probe.FormatBytes(fileSize(outPath))))
}

func runMany(verb string, files []string, build buildFunc) {
	workers := max(rootFlags.jobs, 1)
	utils.PrintRunning(fmt.Sprintf("%s: %d files, %d workers", verb, len(files), workers))

	g, ctx := errgroup.WithContext(context.Background())
	g.SetLimit(workers)

	var mu sync.Mutex
	var results []fileResult
	lineCount := 0

	record := func(r fileResult) {
		mu.Lock()
		defer mu.Unlock()
		results = append(results, r)
		lineCount++
		if r.Err != nil {
			utils.PrintIndentedError(filepath.Base(r.Input), r.Err)
			return
		}
		utils.PrintIndentedSuccess(fmt.Sprintf("%s → %s (%s)",
			filepath.Base(r.Input), filepath.Base(r.Output), probe.FormatBytes(r.OutSize)))
	}

	for _, input := range files {
		g.Go(func() error {
			start := time.Now()
			res, p, outPath, err := prepare(ctx, input, build)
			if err != nil {
				record(fileResult{Input: input, Err: err})
				return nil
			}
			if err := engine.RunFFmpeg(ctx, append(res.Args, outPath), p.TotalDuration(), nil); err != nil {
				record(fileResult{Input: input, OrigSize: p.Format.Size(), Duration: time.Since(start), Err: err})
				return nil
			}
			record(fileResult{
				Input:    input,
				Output:   outPath,
				OrigSize: p.Format.Size(),
				OutSize:  fileSize(outPath),
				Duration: time.Since(start),
			})
			return nil
		})
	}

	_ = g.Wait()
	utils.ClearLines(lineCount + 1)

	var failed []fileResult
	for _, r := range results {
		if r.Err != nil {
			failed = append(failed, r)
		}
	}

	if len(failed) > 0 {
		utils.PrintError(fmt.Sprintf("%s: %d of %d files failed", verb, len(failed), len(files)), nil)
		for _, r := range failed {
			utils.PrintIndentedError(filepath.Base(r.Input), r.Err)
		}
	} else {
		utils.PrintSuccess(fmt.Sprintf("%s: %d files completed", verb, len(files)))
	}

	printSummary(results)
}

func printSummary(results []fileResult) {
	headers := []string{"File", "Status", "Original", "Output", "Savings", "Time"}
	rows := make([][]string, 0, len(results))
	for _, r := range results {
		status := "OK"
		savings := "-"
		if r.Err != nil {
			status = "FAILED"
		} else if r.OrigSize > 0 && r.OutSize > 0 {
			savings = fmt.Sprintf("%.1f%%", float64(r.OrigSize-r.OutSize)/float64(r.OrigSize)*100.0)
		}
		rows = append(rows, []string{
			filepath.Base(r.Input),
			status,
			probe.FormatBytes(r.OrigSize),
			probe.FormatBytes(r.OutSize),
			savings,
			r.Duration.Round(100 * time.Millisecond).String(),
		})
	}
	utils.PrintTable(headers, rows)
}
