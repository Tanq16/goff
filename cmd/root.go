package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/Tanq16/goff/internal/engine"
	"github.com/Tanq16/goff/internal/ops"
	"github.com/Tanq16/goff/internal/presets"
	"github.com/Tanq16/goff/internal/probe"
	"github.com/Tanq16/goff/internal/tui"
	"github.com/Tanq16/goff/utils"
)

var AppVersion = "dev-build"
var debugFlag bool
var forAIFlag bool

var rootFlags struct {
	preset    string
	codec     string
	output    string
	overwrite bool
}

var rootCmd = &cobra.Command{
	Use:     "goff [files...]",
	Short:   "Standalone terminal media suite & FFmpeg CLI/TUI harness",
	Version: AppVersion,
	Args:    cobra.ArbitraryArgs,
	CompletionOptions: cobra.CompletionOptions{
		HiddenDefaultCmd: true,
	},
	Run: func(cmd *cobra.Command, args []string) {
		if rootFlags.preset != "" {
			if len(args) == 0 {
				utils.PrintFatal("at least one input file required when using --preset", nil)
			}
			runHeadlessPreset(args[0], rootFlags.preset)
			return
		}

		if utils.GlobalForAIFlag {
			if len(args) == 0 {
				input := utils.ReadPipedLine()
				if input != "" {
					args = []string{input}
				} else {
					utils.PrintFatal("no input file specified for AI mode", nil)
				}
			}
			runHeadlessPreset(args[0], "web-optimize")
			return
		}

		app, err := tui.NewAppModel(args, rootFlags.overwrite)
		if err != nil {
			utils.PrintFatal("failed to initialize media suite", err)
		}

		p := tea.NewProgram(app)
		if _, err := p.Run(); err != nil {
			utils.PrintFatal("error running media suite", err)
		}
	},
}

func runHeadlessPreset(inputFile string, presetID string) {
	p, err := probe.RunProbe(context.Background(), inputFile)
	if err != nil {
		utils.PrintFatal(fmt.Sprintf("failed to probe %q", inputFile), err)
	}

	preset, ok := presets.Get(presetID)
	if !ok {
		utils.PrintFatal(fmt.Sprintf("unknown preset %q. Available: %s", presetID, presets.AvailablePresetIDs()), nil)
	}

	var res *ops.OpResult
	if rootFlags.codec != "" && presetID == "web-optimize" {
		res, err = ops.BuildVideoOptimize(inputFile, p, ops.VideoOptimizeOpts{
			Codec: rootFlags.codec,
		})
	} else {
		res, err = preset.Build(inputFile, p)
	}
	if err != nil {
		utils.PrintFatal(fmt.Sprintf("failed to build args for preset %q", presetID), err)
	}

	outPath, err := engine.ResolveOutputName(inputFile, res.Suffix, res.TargetExt, rootFlags.output, rootFlags.overwrite)
	if err != nil {
		utils.PrintFatal("failed to resolve output path", err)
	}

	fullArgs := append(res.Args, outPath)
	totalSec := p.TotalDuration()

	utils.PrintRunning(fmt.Sprintf("processing %s with preset %s", filepath.Base(inputFile), preset.ID))

	var curPercent atomic.Int32
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan struct{})
	var printed atomic.Bool

	if !utils.GlobalForAIFlag && !utils.GlobalDebugFlag {
		go func() {
			ticker := time.NewTicker(250 * time.Millisecond)
			defer ticker.Stop()
			firstTick := true
			for {
				select {
				case <-done:
					return
				case <-ticker.C:
					if !firstTick {
						utils.ClearPreviousLine()
					}
					firstTick = false
					printed.Store(true)
					utils.PrintProgress(filepath.Base(inputFile), int(curPercent.Load()))
				}
			}
		}()
	}

	start := time.Now()
	err = engine.RunFFmpeg(ctx, fullArgs, totalSec, func(prog engine.ProgressUpdate) {
		curPercent.Store(int32(prog.Percent))
		if utils.GlobalForAIFlag {
			utils.PrintProgress(filepath.Base(inputFile), prog.Percent)
		}
	})

	close(done)
	if printed.Load() {
		utils.ClearPreviousLine()
	}
	utils.ClearLines(1)

	if err != nil {
		utils.PrintFatal(fmt.Sprintf("failed processing %s", inputFile), err)
	}

	origSize := p.Format.Size()
	var outSize int64
	if fi, statErr := os.Stat(outPath); statErr == nil {
		outSize = fi.Size()
	}

	elapsed := time.Since(start).Round(time.Millisecond * 100)
	utils.PrintSuccess(fmt.Sprintf("%s → %s (%s, %s → %s)", filepath.Base(inputFile), outPath, elapsed, probe.FormatBytes(origSize), probe.FormatBytes(outSize)))
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func setupLogs() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	output := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.DateTime,
		NoColor:    false,
	}
	log.Logger = zerolog.New(output).With().Timestamp().Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if debugFlag {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		utils.GlobalDebugFlag = true
	}
	if forAIFlag {
		utils.GlobalForAIFlag = true
		zerolog.SetGlobalLevel(zerolog.Disabled)
	}
}

func init() {
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})

	rootCmd.PersistentFlags().BoolVar(&debugFlag, "debug", false, "Enable debug logging")
	rootCmd.PersistentFlags().BoolVar(&forAIFlag, "for-ai", false, "AI-friendly output (plain text, piped input)")
	rootCmd.MarkFlagsMutuallyExclusive("debug", "for-ai")

	rootCmd.Flags().StringVarP(&rootFlags.preset, "preset", "p", "", "Directly execute preset ("+presets.AvailablePresetIDs()+")")
	rootCmd.Flags().StringVarP(&rootFlags.codec, "codec", "c", "", "Video codec override (hevc, av1, h264)")
	rootCmd.Flags().StringVarP(&rootFlags.output, "output", "o", "", "Explicit output destination path")
	rootCmd.Flags().BoolVarP(&rootFlags.overwrite, "yes", "y", false, "Allow overwriting existing files")

	cobra.OnInitialize(setupLogs)
}
