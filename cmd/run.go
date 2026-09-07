package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/Tanq16/goff/internal/engine"
	"github.com/Tanq16/goff/internal/ops"
	"github.com/Tanq16/goff/internal/probe"
	"github.com/Tanq16/goff/utils"
)

type buildFunc func(input string, p *probe.ProbeResult) (*ops.OpResult, error)

type buildError struct{ error }

type fileResult struct {
	Input    string
	Outputs  []string
	OrigSize int64
	OutSize  int64
	Duration time.Duration
	Notes    []string
	Err      error
}

func printNotes(notes []string) {
	for _, note := range notes {
		utils.PrintWarn(note, nil)
	}
}

func requireInputs(args []string) {
	if sharedFlags.output != "" && len(args) > 1 {
		utils.PrintFatal("--output names a single destination; drop it to write one output per input", nil)
	}
}

func runContext() (context.Context, context.CancelFunc) {
	return signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
}

func outcome(err error) string {
	if errors.Is(err, context.Canceled) {
		return "cancelled"
	}
	return "failed"
}

func failure(verb, label string, err error) (string, error) {
	var invalid buildError
	if errors.As(err, &invalid) {
		return fmt.Sprintf("cannot %s %s: %v", verb, label, err), nil
	}
	return fmt.Sprintf("%s %s for %s", verb, outcome(err), label), err
}

func discard(res *ops.OpResult, outputs []string) {
	if res != nil && res.Cleanup != nil {
		res.Cleanup()
	}
	for _, out := range outputs {
		os.Remove(out)
	}
}

func encodeArgs(res *ops.OpResult, outputs []string) []string {
	return append(res.Args, outputs[len(outputs)-1])
}

var claimed struct {
	sync.Mutex
	paths map[string]bool
}

func claimOutput(input, suffix, targetExt string) (string, error) {
	if sharedFlags.output != "" {
		return sharedFlags.output, nil
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
		out, err := engine.ResolveOutputName(input, candidate, targetExt)
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

func claimOutputDir(input, suffix string) (string, error) {
	if sharedFlags.output != "" {
		return sharedFlags.output, nil
	}

	dir := filepath.Dir(input)
	base := strings.TrimSuffix(filepath.Base(input), filepath.Ext(input))

	claimed.Lock()
	defer claimed.Unlock()
	if claimed.paths == nil {
		claimed.paths = make(map[string]bool)
	}
	for attempt := range 1000 {
		name := fmt.Sprintf("%s.%s", base, suffix)
		if attempt > 0 {
			name = fmt.Sprintf("%s.%s.%d", base, suffix, attempt)
		}
		candidate := filepath.Join(dir, name)
		if claimed.paths[candidate] {
			continue
		}
		if _, err := os.Stat(candidate); err == nil {
			continue
		}
		claimed.paths[candidate] = true
		return candidate, nil
	}
	return "", fmt.Errorf("unable to find a free output directory for %q", input)
}

func prepare(ctx context.Context, input string, build buildFunc) (*ops.OpResult, *probe.ProbeResult, []string, error) {
	p, err := probe.RunProbe(ctx, input)
	if err != nil {
		return nil, nil, nil, err
	}
	res, err := build(input, p)
	if err != nil {
		return nil, nil, nil, buildError{err}
	}
	if len(res.OutputPaths) > 0 {
		return res, p, res.OutputPaths, nil
	}
	outPath, err := claimOutput(input, res.Suffix, res.TargetExt)
	if err != nil {
		return nil, nil, nil, err
	}
	return res, p, []string{outPath}, nil
}

func encodeWithProgress(ctx context.Context, verb string, label string, args []string, totalSec float64) (time.Duration, error) {
	utils.PrintRunning(fmt.Sprintf("%s %s", verb, label))

	var latest atomic.Pointer[engine.ProgressUpdate]
	latest.Store(&engine.ProgressUpdate{TotalSeconds: totalSec})
	var printed atomic.Bool
	done := make(chan struct{})
	var bar sync.WaitGroup

	show := func() {
		p := latest.Load()
		utils.PrintProgress(label, p.Percent, p.CurrentSeconds, p.TotalSeconds)
	}

	bar.Go(func() {
		t := time.NewTicker(1 * time.Second)
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
				show()
			}
		}
	})

	start := time.Now()
	err := engine.RunFFmpeg(ctx, args, totalSec, func(prog engine.ProgressUpdate) {
		latest.Store(&prog)
	})
	elapsed := time.Since(start)

	close(done)
	bar.Wait()
	if printed.Load() {
		utils.ClearPreviousLine()
	}
	if err == nil {
		show()
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

func totalSize(paths []string) int64 {
	var total int64
	for _, path := range paths {
		total += fileSize(path)
	}
	return total
}

func runFiles(verb string, files []string, build buildFunc) {
	requireInputs(files)
	if len(files) == 1 {
		runSingle(verb, files[0], build)
		return
	}
	runMany(verb, files, build)
}

func runSingle(verb string, input string, build buildFunc) {
	ctx, stop := runContext()
	defer stop()

	label := filepath.Base(input)
	res, p, outPaths, err := prepare(ctx, input, build)
	if err != nil {
		utils.PrintFatal(failure(verb, label, err))
	}

	printNotes(res.Notes)

	elapsed, err := encodeWithProgress(ctx, verb, label, encodeArgs(res, outPaths), p.TotalDuration())
	if err != nil {
		discard(res, outPaths)
		utils.PrintFatal(failure(verb, label, err))
	}

	rounded := elapsed.Round(100 * time.Millisecond)
	if len(outPaths) == 1 {
		utils.PrintSuccess(fmt.Sprintf("%s → %s (%s, %s → %s)",
			label, outPaths[0], rounded,
			probe.FormatBytes(p.Format.Size()), probe.FormatBytes(fileSize(outPaths[0]))))
		return
	}
	utils.PrintSuccess(fmt.Sprintf("%s → %d files (%s, %s → %s)",
		label, len(outPaths), rounded,
		probe.FormatBytes(p.Format.Size()), probe.FormatBytes(totalSize(outPaths))))
	for _, out := range outPaths {
		utils.PrintIndentedSuccess(fmt.Sprintf("%s (%s)", out, probe.FormatBytes(fileSize(out))))
	}
}

func runComposed(verb string, namingInput string, res *ops.OpResult, totalSec float64) {
	outPath, err := claimOutput(namingInput, res.Suffix, res.TargetExt)
	if err != nil {
		discard(res, nil)
		utils.PrintFatal("failed to resolve output path", err)
	}

	printNotes(res.Notes)

	ctx, stop := runContext()
	defer stop()

	label := filepath.Base(namingInput)
	elapsed, err := encodeWithProgress(ctx, verb, label, append(res.Args, outPath), totalSec)
	if res.Cleanup != nil {
		res.Cleanup()
	}
	if err != nil {
		os.Remove(outPath)
		utils.PrintFatal(failure(verb, label, err))
	}

	utils.PrintSuccess(fmt.Sprintf("%s → %s (%s, %s)",
		label, outPath, elapsed.Round(100*time.Millisecond), probe.FormatBytes(fileSize(outPath))))
}

func runMany(verb string, files []string, build buildFunc) {
	workers := max(sharedFlags.jobs, 1)
	utils.PrintRunning(fmt.Sprintf("%s: %d files, %d workers", verb, len(files), workers))

	ctx, stop := runContext()
	defer stop()

	sem := make(chan struct{}, workers)
	var wg sync.WaitGroup

	var mu sync.Mutex
	var results []fileResult
	lineCount := 0

	record := func(r fileResult) {
		mu.Lock()
		defer mu.Unlock()
		results = append(results, r)
		lineCount++
		if r.Err != nil {
			utils.PrintIndentedError(failure(verb, filepath.Base(r.Input), r.Err))
			return
		}
		utils.PrintIndentedSuccess(successLine(r))
	}

	for _, input := range files {
		sem <- struct{}{}
		wg.Go(func() {
			defer func() { <-sem }()
			start := time.Now()
			res, p, outPaths, err := prepare(ctx, input, build)
			if err != nil {
				record(fileResult{Input: input, Err: err})
				return
			}
			label := filepath.Base(input)
			args := encodeArgs(res, outPaths)
			if utils.OutputPersists() {
				_, err = encodeWithProgress(ctx, verb, label, args, p.TotalDuration())
			} else {
				err = engine.RunFFmpeg(ctx, args, p.TotalDuration(), nil)
			}
			if err != nil {
				discard(res, outPaths)
				record(fileResult{Input: input, OrigSize: p.Format.Size(), Duration: time.Since(start), Err: err})
				return
			}
			record(fileResult{
				Input:    input,
				Outputs:  outPaths,
				OrigSize: p.Format.Size(),
				OutSize:  totalSize(outPaths),
				Duration: time.Since(start),
				Notes:    res.Notes,
			})
		})
	}

	wg.Wait()
	utils.ClearLines(lineCount + 1)

	var failed []fileResult
	for _, r := range results {
		if r.Err != nil {
			failed = append(failed, r)
		}
	}

	if len(failed) > 0 {
		utils.PrintError(fmt.Sprintf("%s: %d of %d files failed", verb, len(failed), len(files)), nil)
		if !utils.OutputPersists() {
			for _, r := range failed {
				utils.PrintIndentedError(failure(verb, filepath.Base(r.Input), r.Err))
			}
		}
	} else {
		utils.PrintSuccess(fmt.Sprintf("%s: %d files completed", verb, len(files)))
	}

	for _, r := range results {
		for _, note := range r.Notes {
			utils.PrintIndentedWarn(fmt.Sprintf("%s: %s", filepath.Base(r.Input), note), nil)
		}
	}

	if !utils.GlobalDebugFlag {
		printSummary(results)
	}

	if len(failed) > 0 {
		os.Exit(1)
	}
}

func successLine(r fileResult) string {
	label := filepath.Base(r.Input)
	if len(r.Outputs) == 1 {
		return fmt.Sprintf("%s → %s (%s)", label, filepath.Base(r.Outputs[0]), probe.FormatBytes(r.OutSize))
	}
	return fmt.Sprintf("%s → %d files (%s)", label, len(r.Outputs), probe.FormatBytes(r.OutSize))
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
