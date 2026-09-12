package utils

import (
	"fmt"
	"os"

	"charm.land/lipgloss/v2"
	"github.com/rs/zerolog/log"
)

var (
	ColorBlue    = lipgloss.ANSIColor(12)
	ColorGreen   = lipgloss.ANSIColor(10)
	ColorRed     = lipgloss.ANSIColor(9)
	ColorYellow  = lipgloss.ANSIColor(11)
	ColorMagenta = lipgloss.ANSIColor(13)
	ColorCyan    = lipgloss.ANSIColor(14)
	ColorFg      = lipgloss.ANSIColor(15)
	ColorMuted   = lipgloss.ANSIColor(7)
	ColorChrome  = lipgloss.ANSIColor(8)
)

var (
	infoStyle    = lipgloss.NewStyle().Foreground(ColorBlue)
	successStyle = lipgloss.NewStyle().Foreground(ColorGreen)
	errorStyle   = lipgloss.NewStyle().Foreground(ColorRed)
	warnStyle    = lipgloss.NewStyle().Foreground(ColorYellow)
)

func PrintInfo(msg string) {
	if GlobalDebugFlag {
		log.Info().Msg(msg)
		return
	}
	emit(infoStyle.Render("→ " + msg))
}

func PrintSuccess(msg string) {
	if GlobalDebugFlag {
		log.Info().Msg(msg)
		return
	}
	emit(successStyle.Render("✓ " + msg))
}

func PrintError(msg string, err error) {
	if GlobalDebugFlag {
		if err != nil {
			log.Error().Err(err).Msg(msg)
		} else {
			log.Error().Msg(msg)
		}
		return
	}
	emit(errorStyle.Render("✗ " + msg))
}

func PrintFatal(msg string, err error) {
	PrintError(msg, err)
	ShowCursor()
	os.Exit(1)
}

func PrintWarn(msg string, err error) {
	if GlobalDebugFlag {
		if err != nil {
			log.Warn().Err(err).Msg(msg)
		} else {
			log.Warn().Msg(msg)
		}
		return
	}
	emit(warnStyle.Render("! " + msg))
}

func PrintGeneric(msg string) {
	emit(msg)
}

func PrintRunning(msg string) {
	if GlobalDebugFlag {
		log.Info().Msg(msg)
		return
	}
	emit(infoStyle.Render("↻ " + msg))
}

func PrintIndentedSuccess(msg string) {
	if GlobalDebugFlag {
		log.Info().Msg(msg)
		return
	}
	emit(successStyle.Render("  ✓ " + msg))
}

func PrintIndentedError(msg string, err error) {
	if GlobalDebugFlag {
		if err != nil {
			log.Error().Err(err).Msg(msg)
		} else {
			log.Error().Msg(msg)
		}
		return
	}
	emit(errorStyle.Render("  ✗ " + msg))
}

func PrintIndentedWarn(msg string, err error) {
	if GlobalDebugFlag {
		if err != nil {
			log.Warn().Err(err).Msg(msg)
		} else {
			log.Warn().Msg(msg)
		}
		return
	}
	emit(warnStyle.Render("  ! " + msg))
}

func PrintIndentedRunning(msg string) {
	if GlobalDebugFlag {
		log.Info().Msg(msg)
		return
	}
	emit(infoStyle.Render("  ↻ " + msg))
}

func OutputPersists() bool {
	return GlobalDebugFlag || !StdoutIsTerminal
}

func ClearLines(n int) {
	if OutputPersists() {
		return
	}
	for range n {
		fmt.Print("\033[A\033[2K")
	}
}

func ClearPreviousLine() {
	ClearLines(1)
}
