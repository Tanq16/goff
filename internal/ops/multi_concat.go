package ops

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type MultiConcatOpts struct {
	Inputs   []string
	Reencode bool
}

func BuildMultiConcat(inputs []string, opts MultiConcatOpts) (*OpResult, string, error) {
	if len(inputs) < 2 {
		return nil, "", fmt.Errorf("concatenation requires at least 2 input files")
	}

	ext := strings.TrimPrefix(filepath.Ext(inputs[0]), ".")
	if ext == "" {
		ext = "mp4"
	}

	if !opts.Reencode {
		tmpList, err := os.CreateTemp("", "goff_concat_*.txt")
		if err != nil {
			return nil, "", fmt.Errorf("failed to create temp concat list: %w", err)
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
			return nil, "", err
		}
		tmpList.Close()

		args := []string{
			"-f", "concat",
			"-safe", "0",
			"-i", tmpList.Name(),
			"-c", "copy",
		}
		if ext == "mp4" || ext == "mov" {
			args = append(args, "-movflags", "+faststart")
		}

		return &OpResult{
			Args:      args,
			Suffix:    "merged",
			TargetExt: ext,
		}, tmpList.Name(), nil
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
	}, "", nil
}
