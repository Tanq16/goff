package ops

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Tanq16/goff/internal/probe"
)

type MultiConcatOpts struct {
	Reencode bool
}

func CheckConcatInputs(probes []*probe.ProbeResult, reencode bool) error {
	withVideo, withAudio := 0, 0
	for _, p := range probes {
		if p.IsVideo() {
			withVideo++
		}
		if len(p.AudioStreams()) > 0 {
			withAudio++
		}
	}
	if withVideo != 0 && withVideo != len(probes) {
		return fmt.Errorf("concat needs every input to be the same kind; these mix video files with audio-only files")
	}
	if reencode && (withVideo != len(probes) || withAudio != len(probes)) {
		return fmt.Errorf("--reencode needs every input to carry both a video and an audio stream")
	}
	return nil
}

func BuildMultiConcat(inputs []string, probes []*probe.ProbeResult, opts MultiConcatOpts) (*OpResult, error) {
	if len(inputs) < 2 {
		return nil, fmt.Errorf("concatenation requires at least 2 input files")
	}

	ext := strings.TrimPrefix(filepath.Ext(inputs[0]), ".")
	if ext == "" {
		ext = "mp4"
	}

	if !opts.Reencode {
		tmpList, err := os.CreateTemp("", "goff_concat_*.txt")
		if err != nil {
			return nil, fmt.Errorf("failed to create temp concat list: %w", err)
		}
		var content strings.Builder
		for _, in := range inputs {
			abs, err := filepath.Abs(in)
			if err != nil {
				abs = in
			}
			escaped := strings.ReplaceAll(abs, "'", "'\\''")
			fmt.Fprintf(&content, "file '%s'\n", escaped)
		}
		if _, err := tmpList.WriteString(content.String()); err != nil {
			tmpList.Close()
			os.Remove(tmpList.Name())
			return nil, err
		}
		tmpList.Close()

		args := []string{
			"-f", "concat",
			"-safe", "0",
			"-i", tmpList.Name(),
			"-c", "copy",
		}
		args = tagHEVC(args, copiedVideoIsHEVC(probes...), ext)
		if ext == "mp4" || ext == "mov" {
			args = append(args, "-movflags", "+faststart")
		}

		listPath := tmpList.Name()
		return &OpResult{
			Args:      args,
			Suffix:    "merged",
			TargetExt: ext,
			Cleanup:   func() { os.Remove(listPath) },
		}, nil
	}

	var args []string
	for _, in := range inputs {
		args = append(args, "-i", in)
	}

	var filter strings.Builder
	for i := range inputs {
		fmt.Fprintf(&filter, "[%d:v][%d:a]", i, i)
	}
	fmt.Fprintf(&filter, "concat=n=%d:v=1:a=1[v][a]", len(inputs))

	args = append(args,
		"-filter_complex", filter.String(),
		"-map", "[v]",
		"-map", "[a]",
		"-c:v", "libx264",
		"-crf", "20",
		"-preset", "fast",
		"-c:a", "aac",
		"-b:a", "192k",
		"-movflags", "+faststart",
	)

	return &OpResult{
		Args:      args,
		Suffix:    "merged",
		TargetExt: "mp4",
	}, nil
}
