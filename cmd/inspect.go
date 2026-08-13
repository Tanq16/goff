package cmd

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/Tanq16/goff/internal/probe"
	"github.com/Tanq16/goff/utils"
	"github.com/spf13/cobra"
)

var inspectCmd = &cobra.Command{
	Use:   "inspect <file>",
	Short: "Inspect media streams, codecs, bitrates, and HDR parameters",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filePath := args[0]
		p, err := probe.RunProbe(context.Background(), filePath)
		if err != nil {
			utils.PrintFatal(fmt.Sprintf("failed to probe %q", filePath), err)
		}

		utils.PrintInfo(fmt.Sprintf("File: %s", filepath.Base(filePath)))
		utils.PrintInfo(fmt.Sprintf("Duration: %s | Size: %s | Format: %s", p.HumanDuration(), p.HumanSize(), p.Format.FormatLongName))

		var headers = []string{"#", "Type", "Codec", "Details", "Bitrate", "Default"}
		var rows [][]string

		for _, s := range p.Streams {
			idx := strconv.Itoa(s.Index)
			codecType := s.CodecType
			codecName := s.CodecName
			if s.Profile != "" {
				codecName = fmt.Sprintf("%s (%s)", s.CodecName, s.Profile)
			}

			var details string
			switch s.CodecType {
			case "video":
				fps := probe.ParseFPS(s.AvgFrameRate)
				hdrTag := ""
				if p.IsHDR() {
					hdrTag = fmt.Sprintf(" [%s]", p.HDRType())
				}
				details = fmt.Sprintf("%dx%d @ %.2ffps, %s%s", s.Width, s.Height, fps, s.PixFmt, hdrTag)
			case "audio":
				details = fmt.Sprintf("%s, %d ch (%s)", s.SampleRate+"Hz", s.Channels, s.ChannelLayout)
			case "subtitle":
				lang := s.Tags["language"]
				if lang == "" {
					lang = "und"
				}
				title := s.Tags["title"]
				if title != "" {
					details = fmt.Sprintf("%s (%s)", lang, title)
				} else {
					details = lang
				}
			default:
				details = s.CodecLongName
			}

			bitrateStr := s.BitRate
			if bitrateStr != "" {
				if br, err := strconv.ParseInt(bitrateStr, 10, 64); err == nil {
					bitrateStr = fmt.Sprintf("%d kbps", br/1000)
				}
			} else {
				bitrateStr = "-"
			}

			isDefault := "No"
			if s.Disposition["default"] == 1 {
				isDefault = "Yes"
			}

			rows = append(rows, []string{
				idx,
				codecType,
				codecName,
				details,
				bitrateStr,
				isDefault,
			})
		}

		utils.PrintTable(headers, rows)
	},
}

func init() {
	rootCmd.AddCommand(inspectCmd)
}
