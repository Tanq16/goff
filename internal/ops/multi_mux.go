package ops

import (
	"path/filepath"
	"strings"
)

type MultiMuxAudioOpts struct {
	VideoInput      string
	AudioInput      string
	ReplaceOriginal bool
	AudioBitrate    string
}

func BuildMultiMuxAudio(opts MultiMuxAudioOpts) (*OpResult, error) {
	bitrate := opts.AudioBitrate
	if bitrate == "" {
		bitrate = "192k"
	}

	args := []string{
		"-i", opts.VideoInput,
		"-i", opts.AudioInput,
	}

	if opts.ReplaceOriginal {
		args = append(args,
			"-map", "0:v:0",
			"-map", "1:a:0",
			"-c:v", "copy",
			"-c:a", "aac",
			"-b:a", bitrate,
			"-movflags", "+faststart",
		)
	} else {
		args = append(args,
			"-map", "0:v:0",
			"-map", "0:a?",
			"-map", "1:a:0",
			"-c:v", "copy",
			"-c:a", "aac",
			"-b:a", bitrate,
			"-movflags", "+faststart",
		)
	}

	return &OpResult{
		Args:      args,
		Suffix:    "muxed",
		TargetExt: "mp4",
	}, nil
}

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
