package ops

import (
	"fmt"
	"slices"
	"strings"

	"github.com/Tanq16/goff/internal/probe"
)

type trackPlan struct {
	Audio       []probe.StreamInfo
	Subtitles   []probe.StreamInfo
	Attachments []probe.StreamInfo
	SubEncoder  string
	Notes       []string
}

func containerHoldsEverything(targetExt string) bool {
	return strings.EqualFold(strings.TrimPrefix(targetExt, "."), "mkv")
}

func copiedVideoIsHEVC(probes ...*probe.ProbeResult) bool {
	if len(probes) == 0 {
		return false
	}
	for _, p := range probes {
		if p == nil {
			return false
		}
		videos := p.VideoStreams()
		if len(videos) == 0 {
			return false
		}
		for _, v := range videos {
			if !strings.EqualFold(v.CodecName, "hevc") {
				return false
			}
		}
	}
	return true
}

func containerUsesMOVMuxer(targetExt string) bool {
	ext := strings.ToLower(strings.TrimPrefix(targetExt, "."))
	return ext == "mp4" || ext == "mov"
}

func tagHEVC(args []string, outputIsHEVC bool, targetExt string) []string {
	if !outputIsHEVC || !containerUsesMOVMuxer(targetExt) {
		return args
	}
	return append(args, "-tag:v", "hvc1")
}

func planTracks(p *probe.ProbeResult, audioSel, subSel probe.TrackSelector, targetExt string) (trackPlan, error) {
	var plan trackPlan
	if p == nil {
		return plan, nil
	}

	audio, err := p.SelectAudio(audioSel)
	if err != nil {
		return plan, err
	}
	plan.Audio = audio

	subs, err := p.SelectSubtitles(subSel)
	if err != nil {
		return plan, err
	}

	if containerHoldsEverything(targetExt) {
		plan.Subtitles = subs
		plan.SubEncoder = "copy"
		plan.Attachments = p.AttachmentStreams()
		return plan, nil
	}

	dropped := 0
	allMovText := true
	for _, s := range subs {
		if !probe.IsTextSubtitle(s.CodecName) {
			dropped++
			continue
		}
		if !strings.EqualFold(s.CodecName, "mov_text") {
			allMovText = false
		}
		plan.Subtitles = append(plan.Subtitles, s)
	}
	if dropped > 0 {
		plan.Notes = append(plan.Notes, fmt.Sprintf("dropped %d image-based subtitle track(s), which an %s cannot carry", dropped, strings.ToUpper(strings.TrimPrefix(targetExt, "."))))
	}
	if attachments := p.AttachmentStreams(); len(attachments) > 0 {
		plan.Notes = append(plan.Notes, fmt.Sprintf("dropped %d attachment(s), which an %s cannot carry", len(attachments), strings.ToUpper(strings.TrimPrefix(targetExt, "."))))
	}

	plan.SubEncoder = "mov_text"
	if allMovText {
		plan.SubEncoder = "copy"
	}
	return plan, nil
}

func mapStreams(args []string, streams []probe.StreamInfo) []string {
	for _, s := range streams {
		args = append(args, "-map", fmt.Sprintf("0:%d", s.Index))
	}
	return args
}

func promoteDefaultAudio(plan *trackPlan, p *probe.ProbeResult, sel probe.TrackSelector) (bool, error) {
	if sel.String() == "" || p == nil {
		return false, nil
	}
	picked, err := p.SelectAudio(sel)
	if err != nil {
		return false, err
	}
	if len(picked) != 1 {
		return false, fmt.Errorf("%s matches %d audio tracks, so it cannot name the default", sel, len(picked))
	}
	index := slices.IndexFunc(plan.Audio, func(s probe.StreamInfo) bool { return s.Index == picked[0].Index })
	if index < 0 {
		return false, fmt.Errorf("audio stream %d is not among the kept tracks, so it cannot be the default", picked[0].Index)
	}
	ordered := make([]probe.StreamInfo, 0, len(plan.Audio))
	ordered = append(ordered, plan.Audio[index])
	ordered = append(ordered, plan.Audio[:index]...)
	plan.Audio = append(ordered, plan.Audio[index+1:]...)
	return true, nil
}

func audioDispositions(args []string, count int) []string {
	for i := range count {
		state := "0"
		if i == 0 {
			state = "default"
		}
		args = append(args, fmt.Sprintf("-disposition:a:%d", i), state)
	}
	return args
}
