package ops

import (
	"fmt"
	"strconv"
	"strings"
)

type ExtractTarget struct {
	StreamIndex int
	Path        string
}

type VideoExtractOpts struct {
	Format     string
	Bitrate    string
	SampleRate int
	Channels   int
}

func subtitleEncoder(format string) (string, bool) {
	switch strings.ToLower(format) {
	case "srt":
		return "srt", true
	case "vtt":
		return "webvtt", true
	}
	return "", false
}

func ExtractFormat(format string) (string, bool) {
	if _, ok := subtitleEncoder(format); ok {
		return strings.ToLower(format), true
	}
	_, ext, _ := audioCodecFor(format, "")
	return ext, false
}

func BuildExtract(inputPath string, targets []ExtractTarget, opts VideoExtractOpts) (*OpResult, error) {
	if len(targets) == 0 {
		return nil, fmt.Errorf("no stream to extract")
	}

	var perTarget []string
	if encoder, ok := subtitleEncoder(opts.Format); ok {
		perTarget = []string{"-c:s", encoder}
	} else {
		encodeArgs, _ := audioEncodeArgs(opts.Format, opts.Bitrate)
		perTarget = append(encodeArgs, "-af", AudioSyncFilter)
		if opts.SampleRate > 0 {
			perTarget = append(perTarget, "-ar", strconv.Itoa(opts.SampleRate))
		}
		if opts.Channels > 0 {
			perTarget = append(perTarget, "-ac", strconv.Itoa(opts.Channels))
		}
	}

	args := []string{"-i", inputPath}
	for i, t := range targets {
		args = append(args, "-map", fmt.Sprintf("0:%d", t.StreamIndex))
		args = append(args, perTarget...)
		if i < len(targets)-1 {
			args = append(args, t.Path)
		}
	}

	return &OpResult{
		Args:       args,
		OutputPath: targets[len(targets)-1].Path,
	}, nil
}
