package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

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
	Use:     "goff",
	Short:   "Standalone terminal media suite & FFmpeg CLI harness",
	Version: AppVersion,
	Long: `Standalone terminal media suite & FFmpeg CLI harness.

Every per-file command takes one or more inputs and writes one output per input,
next to the source, so nothing is overwritten without -y.

concat joins clips end to end, while mix layers audio tracks so they play at the
same time.`,
	CompletionOptions: cobra.CompletionOptions{
		HiddenDefaultCmd: true,
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
	rootCmd.SetHelpCommandGroupID("info")

	rootCmd.PersistentFlags().BoolVar(&debugFlag, "debug", false, "Enable debug logging")
	rootCmd.PersistentFlags().BoolVar(&forAIFlag, "for-ai", false, "AI-friendly output (plain text, piped input)")
	rootCmd.MarkFlagsMutuallyExclusive("debug", "for-ai")

	rootCmd.PersistentFlags().StringVarP(&rootFlags.output, "output", "o", "", "Explicit output path (single input only)")
	rootCmd.PersistentFlags().BoolVarP(&rootFlags.overwrite, "yes", "y", false, "Allow overwriting existing files")
	rootCmd.PersistentFlags().IntVarP(&rootFlags.jobs, "jobs", "j", 2, "Concurrent encodes when several inputs are given")

	cobra.OnInitialize(setupLogs)

	rootCmd.AddGroup(
		&cobra.Group{ID: "video", Title: "Video:"},
		&cobra.Group{ID: "audio", Title: "Audio:"},
		&cobra.Group{ID: "segments", Title: "Segments:"},
		&cobra.Group{ID: "combine", Title: "Combining:"},
		&cobra.Group{ID: "info", Title: "Info:"},
	)

	rootCmd.AddCommand(
		compressCmd, remuxCmd, hlsCmd, rotateCmd, scaleCmd, speedCmd, cropCmd, muteCmd, watermarkCmd,
		extractCmd, convertCmd, normalizeCmd, mixCmd,
		trimCmd, gifCmd,
		concatCmd, muxCmd, subsCmd,
		inspectCmd,
	)
}
