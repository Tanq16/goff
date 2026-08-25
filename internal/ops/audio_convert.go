package ops

import (
	"cmp"
	"fmt"
	"strings"

	"github.com/Tanq16/goff/internal/probe"
)

func audioCodecFor(format, bitrate string) (string, string, string) {
	switch strings.ToLower(format) {
	case "aac", "m4a":
		return "aac", "m4a", cmp.Or(bitrate, "256k")
	case "opus":
		return "libopus", "opus", cmp.Or(bitrate, "160k")
	case "flac":
		return "flac", "flac", ""
	case "wav":
		return "pcm_s16le", "wav", ""
	case "ogg":
		return "libvorbis", "ogg", cmp.Or(bitrate, "192k")
	}
	return "libmp3lame", "mp3", cmp.Or(bitrate, "320k")
}

func audioEncodeArgs(format, bitrate string) ([]string, string) {
	codec, ext, resolved := audioCodecFor(format, bitrate)
	args := []string{"-c:a", codec}
	if resolved != "" {
		args = append(args, "-b:a", resolved)
	}
	return args, ext
}

type AudioConvertOpts struct {
	Format     string
	Bitrate    string
	SampleRate int
	Channels   int
}

func BuildAudioConvert(inputPath string, p *probe.ProbeResult, opts AudioConvertOpts) (*OpResult, error) {
	args := []string{"-i", inputPath}

	if p != nil && p.IsVideo() {
		args = append(args, "-vn")
	}

	encodeArgs, ext := audioEncodeArgs(opts.Format, opts.Bitrate)
	args = append(args, encodeArgs...)

	if opts.SampleRate > 0 {
		args = append(args, "-ar", fmt.Sprintf("%d", opts.SampleRate))
	}

	if opts.Channels > 0 {
		args = append(args, "-ac", fmt.Sprintf("%d", opts.Channels))
	}

	return &OpResult{
		Args:      args,
		Suffix:    "converted",
		TargetExt: ext,
	}, nil
}
