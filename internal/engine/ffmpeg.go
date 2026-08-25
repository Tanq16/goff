package engine

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
)

func RunFFmpeg(ctx context.Context, args []string, totalDurationSec float64, onProgress ProgressCallback) error {
	fullArgs := append([]string{"-hide_banner", "-loglevel", "error", "-nostats", "-progress", "pipe:1", "-y"}, args...)

	cmd := exec.CommandContext(ctx, "ffmpeg", fullArgs...)

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to open stdout pipe: %w", err)
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to open stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start ffmpeg: %w", err)
	}

	var stderrBuf bytes.Buffer
	var wg sync.WaitGroup

	wg.Go(func() {
		_ = ScanProgress(stdoutPipe, totalDurationSec, onProgress)
	})

	wg.Go(func() {
		_, _ = io.Copy(&stderrBuf, stderrPipe)
	})

	wg.Wait()

	if err := cmd.Wait(); err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		errDetail := strings.TrimSpace(stderrBuf.String())
		if errDetail != "" {
			return fmt.Errorf("ffmpeg error: %s (%w)", errDetail, err)
		}
		return fmt.Errorf("ffmpeg execution failed: %w", err)
	}

	return nil
}
