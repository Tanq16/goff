package cmd

import (
	"context"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/Tanq16/goff/internal/probe"
	"github.com/Tanq16/goff/utils"
)

var inspectFlags struct {
	json  bool
	check bool
}

var inspectCmd = &cobra.Command{
	Use:     "inspect <file>",
	GroupID: "info",
	Short:   "Inspect media streams, codecs, bitrates, and HDR parameters",
	Example: `  goff inspect movie.mkv
  goff inspect movie.mkv --json
  goff inspect movie.mp4 --check`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filePath := args[0]
		ctx := context.Background()

		p, err := probe.RunProbe(ctx, filePath)
		if err != nil {
			utils.PrintFatal(fmt.Sprintf("failed to probe %q", filePath), err)
		}

		summary := p.Summarize(filePath)

		if inspectFlags.check {
			conformance, err := probe.Conform(ctx, filePath, p)
			if err != nil {
				utils.PrintFatal(fmt.Sprintf("failed to check %q", filePath), err)
			}
			summary.Conformance = conformance
		}

		if inspectFlags.json {
			encoded, err := json.Marshal(summary, jsontext.WithIndent("  "))
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

		if summary.Conformance != nil {
			printConformance(summary.Conformance)
		}
	},
}

func printConformance(c *probe.Conformance) {
	utils.PrintInfo(fmt.Sprintf("Content: video %.3fs | audio %.3fs | drift %+.3fs",
		c.VideoContentSeconds, c.AudioContentSeconds, c.DriftSeconds))
	utils.PrintInfo(fmt.Sprintf("Start: video %.3fs | audio %.3fs | offset %+.3fs",
		c.VideoStartSeconds, c.AudioStartSeconds, c.StartOffsetSeconds))
	utils.PrintInfo(fmt.Sprintf("Timeline gaps: video %d (%.3fs) | audio %d (%.3fs)",
		c.VideoGapCount, c.VideoGapSeconds, c.AudioGapCount, c.AudioGapSeconds))

	if c.BrowserSafe {
		utils.PrintSuccess("Browser-safe")
		return
	}
	utils.PrintWarn("Not browser-safe", nil)
	for _, issue := range c.Issues {
		utils.PrintIndentedWarn(issue, nil)
	}
}

func init() {
	inspectCmd.Flags().BoolVar(&inspectFlags.json, "json", false, "Emit the inspection as JSON instead of a table")
	inspectCmd.Flags().BoolVar(&inspectFlags.check, "check", false, "Scan every packet and report timeline gaps, A/V drift, and browser playability")
}
