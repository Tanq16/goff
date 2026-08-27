package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/Tanq16/goff/internal/probe"
	"github.com/Tanq16/goff/utils"
)

var inspectFlags struct {
	json bool
}

var inspectCmd = &cobra.Command{
	Use:     "inspect <file>",
	GroupID: "info",
	Short:   "Inspect media streams, codecs, bitrates, and HDR parameters",
	Example: `  goff inspect movie.mkv
  goff inspect movie.mkv --json`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filePath := args[0]
		p, err := probe.RunProbe(context.Background(), filePath)
		if err != nil {
			utils.PrintFatal(fmt.Sprintf("failed to probe %q", filePath), err)
		}

		summary := p.Summarize(filePath)

		if inspectFlags.json {
			encoded, err := json.MarshalIndent(summary, "", "  ")
			if err != nil {
				utils.PrintFatal("failed to encode the inspection as json", err)
			}
			utils.PrintGeneric(string(encoded))
			return
		}

		utils.PrintInfo(fmt.Sprintf("File: %s", summary.File))
		utils.PrintInfo(fmt.Sprintf("Duration: %s | Size: %s | Bitrate: %s | Format: %s",
			summary.Duration, summary.Size, summary.Bitrate, summary.Format))

		rows := make([][]string, 0, len(summary.Streams))
		for _, s := range summary.Streams {
			isDefault := "No"
			if s.Default {
				isDefault = "Yes"
			}
			rows = append(rows, []string{strconv.Itoa(s.Index), s.Type, s.Codec, s.Details, s.Bitrate, isDefault})
		}
		utils.PrintTable([]string{"#", "Type", "Codec", "Details", "Bitrate", "Default"}, rows)
	},
}

func init() {
	inspectCmd.Flags().BoolVar(&inspectFlags.json, "json", false, "Emit the inspection as JSON instead of a table")
}
