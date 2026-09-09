package ops

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Tanq16/goff/internal/subs"
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

	paths := make([]string, 0, len(targets))
	for _, t := range targets {
		paths = append(paths, t.Path)
	}

	res := &OpResult{
		Args:        args,
		OutputPaths: paths,
	}
	if strings.EqualFold(opts.Format, "vtt") {
		res.PostProcess = cleanVTTOutputs
	}
	return res, nil
}

func cleanVTTOutputs(outputs []string) error {
	for _, out := range outputs {
		if err := subs.CleanVTT(out); err != nil {
			return err
		}
	}
	return nil
}
