package ops

import (
	"fmt"

	"github.com/Tanq16/goff/internal/probe"
)

type MultiMuxSubsOpts struct {
	VideoInput string
	SubsInput  string
	Language   string
	Default    bool
}

func BuildMultiMuxSubs(p *probe.ProbeResult, opts MultiMuxSubsOpts) (*OpResult, error) {
	targetExt := videoContainerFor(opts.VideoInput)

	plan, err := planTracks(p, probe.AllTracks(), probe.AllTracks(), targetExt)
	if err != nil {
		return nil, err
	}

	args := []string{"-i", opts.VideoInput, "-i", opts.SubsInput}
	if p != nil {
		args = mapStreams(args, p.VideoStreams())
	}
	args = mapStreams(args, plan.Audio)
	args = mapStreams(args, plan.Subtitles)
	args = mapStreams(args, plan.Attachments)
	args = append(args, "-map", "1:0", "-c", "copy")
	args = tagHEVC(args, copiedVideoIsHEVC(p), targetExt)

	added := len(plan.Subtitles)
	if !containerHoldsEverything(targetExt) {
		args = append(args, "-c:s", "mov_text")
	}
	if opts.Language != "" {
		args = append(args, fmt.Sprintf("-metadata:s:s:%d", added), "language="+opts.Language)
	}
	if opts.Default {
		for i := range added {
			args = append(args, fmt.Sprintf("-disposition:s:%d", i), "0")
		}
		args = append(args, fmt.Sprintf("-disposition:s:%d", added), "default")
	}
	if containerUsesMOVMuxer(targetExt) {
		args = append(args, "-movflags", "+faststart")
	}

	return &OpResult{
		Args:      args,
		Suffix:    "subtitled",
		TargetExt: targetExt,
		Notes:     plan.Notes,
	}, nil
}
