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
	MediaSec float64
	Duration time.Duration
	Notes    []string
	Err      error
}

const (
	mediaUnit       = utils.Unit("s")
	perFileMeterMax = 15
)

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
	if invalid, ok := errors.AsType[buildError](err); ok {
		return fmt.Sprintf("cannot %s %s: %v", verb, label, invalid.error), err
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

func postProcess(res *ops.OpResult, outputs []string) error {
	if res.PostProcess == nil {
		return nil
	}
	return res.PostProcess(outputs)
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

func encode(ctx context.Context, m *utils.Meter, args []string, totalSec float64) error {
	return engine.RunFFmpeg(ctx, args, totalSec, func(prog engine.ProgressUpdate) {
		m.Set(int64(prog.CurrentSeconds))
	})
}

func encodeRate(mediaSec float64, elapsed time.Duration) string {
	return utils.FormatRate(mediaSec/max(elapsed.Seconds(), 0.001), mediaUnit)
}

func outputName(label string, outputs []string, base bool) string {
	if len(outputs) != 1 {
		return fmt.Sprintf("%s → %d files", label, len(outputs))
	}
	if base {
		return fmt.Sprintf("%s → %s", label, filepath.Base(outputs[0]))
	}
	return fmt.Sprintf("%s → %s", label, outputs[0])
}

func sizeChange(orig, out int64) string {
	return fmt.Sprintf("%s → %s", probe.FormatBytes(orig), probe.FormatBytes(out))
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

	mediaSec := p.TotalDuration()
	m := utils.NewMeter(verb, label, int64(mediaSec), mediaUnit)
	err = encode(ctx, m, encodeArgs(res, outPaths), mediaSec)
	elapsed := m.Close()
	if err == nil {
		err = postProcess(res, outPaths)
	}
	if err != nil {
		discard(res, outPaths)
		utils.PrintFatal(failure(verb, label, err))
	}

	utils.PrintSuccess(utils.SettledLine(
		outputName(label, outPaths, false),
		sizeChange(p.Format.Size(), totalSize(outPaths)),
		elapsed,
		encodeRate(mediaSec, elapsed),
	))
	if len(outPaths) == 1 {
		return
	}
	for _, out := range outPaths {
		utils.PrintIndentedSuccess(fmt.Sprintf("%s  %s", out, probe.FormatBytes(fileSize(out))))
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
	m := utils.NewMeter(verb, label, int64(totalSec), mediaUnit)
	err = encode(ctx, m, append(res.Args, outPath), totalSec)
	elapsed := m.Close()
	if res.Cleanup != nil {
		res.Cleanup()
	}
	if err == nil {
		err = postProcess(res, []string{outPath})
	}
	if err != nil {
		os.Remove(outPath)
		utils.PrintFatal(failure(verb, label, err))
	}

	utils.PrintSuccess(utils.SettledLine(
		outputName(label, []string{outPath}, false),
		probe.FormatBytes(fileSize(outPath)),
		elapsed,
		encodeRate(totalSec, elapsed),
	))
}

func runMany(verb string, files []string, build buildFunc) {
	workers := max(sharedFlags.jobs, 1)
	perFile := utils.OutputPersists() || (workers == 1 && len(files) <= perFileMeterMax)

	ctx, stop := runContext()
	defer stop()

	g := utils.NewGroup(verb, "files")
	if !perFile && len(files) > perFileMeterMax {
		g.Collapse()
	}
	var counter *utils.Meter
	if !perFile {
		counter = g.Count(verb, int64(len(files)))
		if workers > 1 {
			counter.Context(fmt.Sprintf("%d workers", workers))
		}
	}

	sem := make(chan struct{}, workers)
	var wg sync.WaitGroup

	var mu sync.Mutex
	var results []fileResult

	record := func(r fileResult) {
		mu.Lock()
		results = append(results, r)
		mu.Unlock()
		if r.Err != nil {
			g.Fail(filepath.Base(r.Input), r.Err)
			return
		}
		g.OK(
			outputName(filepath.Base(r.Input), r.Outputs, true),
			sizeChange(r.OrigSize, r.OutSize),
			r.Duration,
			encodeRate(r.MediaSec, r.Duration),
		)
	}

dispatch:
	for _, input := range files {
		select {
		case <-ctx.Done():
			break dispatch
		default:
		}
		sem <- struct{}{}
		wg.Go(func() {
			defer func() { <-sem }()
			select {
			case <-ctx.Done():
				return
			default:
			}
			label := filepath.Base(input)
			if counter != nil && workers == 1 {
				counter.Context(label)
			}
			start := time.Now()
			res, p, outPaths, err := prepare(ctx, input, build)
			if err != nil {
				record(fileResult{Input: input, Err: err})
				return
			}
			mediaSec := p.TotalDuration()
			args := encodeArgs(res, outPaths)
			if perFile {
				m := g.Meter(verb, label, int64(mediaSec), mediaUnit)
				err = encode(ctx, m, args, mediaSec)
				m.Close()
			} else {
				err = engine.RunFFmpeg(ctx, args, mediaSec, nil)
			}
			if err == nil {
				err = postProcess(res, outPaths)
			}
			if err != nil {
				discard(res, outPaths)
				record(fileResult{Input: input, OrigSize: p.Format.Size(), MediaSec: mediaSec, Duration: time.Since(start), Err: err})
				return
			}
			record(fileResult{
				Input:    input,
				Outputs:  outPaths,
				OrigSize: p.Format.Size(),
				OutSize:  totalSize(outPaths),
				MediaSec: mediaSec,
				Duration: time.Since(start),
				Notes:    res.Notes,
			})
		})
	}

	wg.Wait()

	var moved int64
	failed := 0
	for _, r := range results {
		moved += r.OutSize
		if r.Err != nil {
			failed++
		}
	}
	written := ""
	if moved > 0 {
		written = probe.FormatBytes(moved)
	}
	g.Done(written)

	if ctx.Err() != nil && len(results) < len(files) {
		utils.PrintWarn(fmt.Sprintf("%s cancelled: %d of %d files were never started", verb, len(files)-len(results), len(files)), nil)
	}

	for _, r := range results {
		for _, note := range r.Notes {
			utils.PrintIndentedWarn(fmt.Sprintf("%s: %s", filepath.Base(r.Input), note), nil)
		}
	}

	if !utils.GlobalDebugFlag {
		printSummary(results)
	}

	if failed > 0 {
		os.Exit(1)
	}
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
			utils.FormatElapsed(r.Duration),
		})
	}
	utils.PrintTable(headers, rows)
}
