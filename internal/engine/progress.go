package engine

import (
	"bufio"
	"io"
	"strconv"
	"strings"
)

type ProgressUpdate struct {
	Percent       int
	CurrentMicros int64
	TotalSeconds  float64
	Speed         float64
	FPS           float64
	OutBytes      int64
	Done          bool
}

type ProgressCallback func(p ProgressUpdate)

func ScanProgress(r io.Reader, totalDurationSec float64, onProgress ProgressCallback) error {
	scanner := bufio.NewScanner(r)
	var current ProgressUpdate
	current.TotalSeconds = totalDurationSec

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)

		switch key {
		case "out_time_us":
			if us, err := strconv.ParseInt(val, 10, 64); err == nil {
				current.CurrentMicros = us
				if totalDurationSec > 0 {
					pct := int(float64(us) / (totalDurationSec * 1_000_000.0) * 100)
					current.Percent = min(max(pct, 0), 100)
				}
			}
		case "total_size":
			if sz, err := strconv.ParseInt(val, 10, 64); err == nil {
				current.OutBytes = sz
			}
		case "speed":
			if s, err := strconv.ParseFloat(strings.TrimSuffix(val, "x"), 64); err == nil {
				current.Speed = s
			}
		case "fps":
			if f, err := strconv.ParseFloat(val, 64); err == nil {
				current.FPS = f
			}
		case "progress":
			if val == "end" {
				current.Done = true
				current.Percent = 100
			}
			if onProgress != nil {
				onProgress(current)
			}
		}
	}

	return scanner.Err()
}
