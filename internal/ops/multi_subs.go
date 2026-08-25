package ops

import (
	"path/filepath"
	"strings"
)

type MultiMuxSubsOpts struct {
	VideoInput string
	SubsInput  string
	Hardburn   bool
}

func BuildMultiMuxSubs(opts MultiMuxSubsOpts) (*OpResult, error) {
	if opts.Hardburn {
		absSubs, err := filepath.Abs(opts.SubsInput)
		if err != nil {
			absSubs = opts.SubsInput
		}
		escaped := strings.ReplaceAll(absSubs, ":", "\\:")
		escaped = strings.ReplaceAll(escaped, "'", "'\\''")

		args := []string{
			"-i", opts.VideoInput,
			"-vf", "subtitles='" + escaped + "'",
			"-c:v", "libx264",
			"-crf", "20",
			"-preset", "medium",
			"-c:a", "copy",
			"-movflags", "+faststart",
		}
		return &OpResult{
			Args:      args,
			Suffix:    "hardsub",
			TargetExt: "mp4",
		}, nil
	}

	ext := strings.ToLower(filepath.Ext(opts.VideoInput))
	subExt := strings.ToLower(filepath.Ext(opts.SubsInput))

	args := []string{
		"-i", opts.VideoInput,
		"-i", opts.SubsInput,
		"-map", "0",
		"-map", "1",
		"-c", "copy",
	}

	targetExt := "mp4"
	if ext == ".mkv" {
		targetExt = "mkv"
	} else if subExt == ".srt" || subExt == ".vtt" {
		args = append(args, "-c:s", "mov_text")
	}

	return &OpResult{
		Args:      args,
		Suffix:    "subtitled",
		TargetExt: targetExt,
	}, nil
}
