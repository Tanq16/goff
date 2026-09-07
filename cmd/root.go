package cmd

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/Tanq16/goff/utils"
)

var AppVersion = "dev-build"
var debugFlag bool

var rootCmd = &cobra.Command{
	Use:     "goff",
	Short:   "Standalone terminal media suite & FFmpeg CLI harness",
	Version: AppVersion,
	Long: `Standalone terminal media suite & FFmpeg CLI harness.

Every per-file command takes one or more inputs and writes one output per input,
next to the source, numbering the name on a collision so nothing is overwritten.
Only -o overwrites, since it names the destination outright.

concat joins clips end to end, while mix layers audio tracks so they play at the
same time.`,
	CompletionOptions: cobra.CompletionOptions{
		HiddenDefaultCmd: true,
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func setupLogs() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	var out io.Writer = os.Stdout
	if utils.StdoutIsTerminal {
		out = zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.DateTime}
	}
	log.Logger = zerolog.New(out).With().Timestamp().Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if debugFlag {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		utils.GlobalDebugFlag = true
	}
}

func init() {
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})
	rootCmd.SetHelpCommandGroupID("info")

	rootCmd.PersistentFlags().BoolVar(&debugFlag, "debug", false, "Enable debug logging")

	cobra.OnInitialize(setupLogs)

	rootCmd.AddGroup(
		&cobra.Group{ID: "video", Title: "Video:"},
		&cobra.Group{ID: "audio", Title: "Audio:"},
		&cobra.Group{ID: "segments", Title: "Segments:"},
		&cobra.Group{ID: "combine", Title: "Combining:"},
		&cobra.Group{ID: "info", Title: "Info:"},
	)

	rootCmd.AddCommand(
		compressCmd, remuxCmd, hlsCmd, transformCmd, watermarkCmd,
		extractCmd, convertCmd, normalizeCmd, mixCmd,
		trimCmd, gifCmd, thumbnailCmd,
		concatCmd, muxCmd, subsCmd,
		inspectCmd,
	)
}
