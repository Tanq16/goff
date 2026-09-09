package ops

import (
	"cmp"
	"strings"

	"github.com/Tanq16/goff/internal/probe"
)

type VideoRemuxOpts struct {
	TargetExt     string
	FixTimestamps bool
	AudioTracks   probe.TrackSelector
	SubTracks     probe.TrackSelector
	DefaultAudio  probe.TrackSelector
}

func BuildVideoRemux(inputPath string, p *probe.ProbeResult, opts VideoRemuxOpts) (*OpResult, error) {
	targetExt := cmp.Or(strings.ToLower(strings.TrimPrefix(opts.TargetExt, ".")), "mp4")

	plan, err := planTracks(p, opts.AudioTracks, opts.SubTracks, targetExt)
	if err != nil {
		return nil, err
	}

	promoted, err := promoteDefaultAudio(&plan, p, opts.DefaultAudio)
	if err != nil {
		return nil, err
	}

	args := []string{"-i", inputPath}
	if p != nil {
		args = mapStreams(args, p.VideoStreams())
	}
	args = mapStreams(args, plan.Audio)
	args = mapStreams(args, plan.Subtitles)
	args = mapStreams(args, plan.Attachments)

	args = append(args, "-c", "copy")
	if promoted {
		args = audioDispositions(args, len(plan.Audio))
	}
	args = tagHEVC(args, copiedVideoIsHEVC(p), targetExt)
	if len(plan.Subtitles) > 0 && plan.SubEncoder != "copy" {
		args = append(args, "-c:s", plan.SubEncoder)
	}
	if opts.FixTimestamps {
		args = append(args, "-avoid_negative_ts", "make_zero")
	}
	if containerUsesMOVMuxer(targetExt) {
		args = append(args, "-movflags", "+faststart")
	}

	return &OpResult{
		Args:      args,
		Suffix:    "remux",
		TargetExt: targetExt,
		Notes:     plan.Notes,
	}, nil
}
