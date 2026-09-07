package cmd

import (
	"github.com/spf13/cobra"

	"github.com/Tanq16/goff/internal/ops"
	"github.com/Tanq16/goff/internal/probe"
)

var remuxFlags struct {
	to            string
	fixTimestamps bool
	audioTracks   probe.TrackSelector
	subTracks     probe.TrackSelector
}

var remuxCmd = &cobra.Command{
	Use:     "remux <files...>",
	GroupID: "video",
	Short:   "Switch container without re-encoding a single stream",
	Example: `  goff remux capture.mkv --to mp4
  goff remux film.mkv --to mp4 --audio-track eng`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runFiles("remux", args, func(input string, p *probe.ProbeResult) (*ops.OpResult, error) {
			return ops.BuildVideoRemux(input, p, ops.VideoRemuxOpts{
				TargetExt:     remuxFlags.to,
				FixTimestamps: remuxFlags.fixTimestamps,
				AudioTracks:   remuxFlags.audioTracks,
				SubTracks:     remuxFlags.subTracks,
			})
		})
	},
}

func init() {
	remuxCmd.Flags().Var(newEnum(&remuxFlags.to, "mp4", "mp4", "mkv", "mov"), "to", "Target container")
	remuxCmd.Flags().BoolVar(&remuxFlags.fixTimestamps, "fix-timestamps", false, "Shift negative start timestamps to zero")
	remuxCmd.Flags().Var(newTrack(&remuxFlags.audioTracks, probe.AllTracks()), "audio-track", "Audio tracks to keep, by stream index or language")
	remuxCmd.Flags().Var(newTrack(&remuxFlags.subTracks, probe.AllTracks()), "sub-track", "Subtitle tracks to keep, by stream index or language")

	addOutputFlag(remuxCmd)
	addJobsFlag(remuxCmd)
}
