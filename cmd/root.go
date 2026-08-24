package cmd

import (
	"fmt"
	"os"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/Tanq16/goff/internal/tui"
	"github.com/Tanq16/goff/utils"
)

var AppVersion = "dev-build"
var debugFlag bool
var forAIFlag bool

var rootFlags struct {
	output    string
	overwrite bool
	jobs      int
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
		if utils.GlobalForAIFlag {
			utils.PrintFatal("--for-ai needs an explicit command, e.g. goff compress <file> --for-ai", nil)
		}

		app, err := tui.NewAppModel(args, rootFlags.overwrite)
		if err != nil {
			utils.PrintFatal("failed to initialize media suite", err)
		}
		if _, err := tea.NewProgram(app).Run(); err != nil {
			utils.PrintFatal("error running media suite", err)
		}
	},
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

	rootCmd.PersistentFlags().StringVarP(&rootFlags.output, "output", "o", "", "Explicit output path (single input only)")
	rootCmd.PersistentFlags().BoolVarP(&rootFlags.overwrite, "yes", "y", false, "Allow overwriting existing files")
	rootCmd.PersistentFlags().IntVarP(&rootFlags.jobs, "jobs", "j", 2, "Concurrent encodes when several inputs are given")

	cobra.OnInitialize(setupLogs)
}
