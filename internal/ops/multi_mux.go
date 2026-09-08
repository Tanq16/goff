package ops

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Tanq16/goff/internal/probe"
)

const (
	MuxModeMix      = "mix"
	MuxModeReplace  = "replace"
	MuxModeSeparate = "separate"
)

type MultiMuxAudioOpts struct {
	VideoInput   string
	Sources      []AudioSource
	Mode         string
	Fit          string
	AudioBitrate string
}

func videoContainerFor(path string) string {
	switch strings.ToLower(strings.TrimPrefix(filepath.Ext(path), ".")) {
	case "mkv":
		return "mkv"
	case "mov":
		return "mov"
	}
	return "mp4"
}

func BuildMultiMuxAudio(p *probe.ProbeResult, opts MultiMuxAudioOpts) (*OpResult, error) {
	if opts.VideoInput == "" {
		return nil, fmt.Errorf("muxing audio requires a video input")
	}
	if len(opts.Sources) == 0 {
		return nil, fmt.Errorf("muxing audio requires at least one audio input")
	}

	bitrate := opts.AudioBitrate
	if bitrate == "" {
		bitrate = "192k"
	}

	videoHasAudio := p != nil && len(p.AudioStreams()) > 0

	args := []string{"-i", opts.VideoInput}
	for _, s := range opts.Sources {
		args = append(args, "-i", s.Path)
	}

	if opts.Mode == MuxModeSeparate {
		args = append(args, "-map", "0:v:0")
		if videoHasAudio {
			args = append(args, "-map", "0:a")
		}
		for i := range opts.Sources {
			args = append(args, "-map", fmt.Sprintf("%d:a:0", i+1))
		}
		args = append(args, "-c:v", "copy", "-c:a", "aac", "-b:a", bitrate)
	} else {
		labels := make([]string, 0, len(opts.Sources)+1)
		sources := make([]AudioSource, 0, len(opts.Sources)+1)
		if opts.Mode != MuxModeReplace && videoHasAudio {
			labels = append(labels, "[0:a]")
			sources = append(sources, AudioSource{Volume: 1.0})
		}
		for i, s := range opts.Sources {
			labels = append(labels, fmt.Sprintf("[%d:a]", i+1))
			sources = append(sources, s)
		}
		args = append(args,
			"-filter_complex", mixFilter(labels, sources, opts.Fit, "[aout]"),
			"-map", "0:v:0", "-map", "[aout]",
			"-c:v", "copy", "-c:a", "aac", "-b:a", bitrate,
		)
	}

	if opts.Fit != "longest" {
		args = append(args, "-shortest")
	}

	targetExt := videoContainerFor(opts.VideoInput)
	args = tagHEVC(args, copiedVideoIsHEVC(p), targetExt)
	if targetExt == "mp4" || targetExt == "mov" {
		args = append(args, "-movflags", "+faststart")
	}

	return &OpResult{
		Args:      args,
		Suffix:    "muxed",
		TargetExt: targetExt,
	}, nil
}
