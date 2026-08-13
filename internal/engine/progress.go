package engine

import (
	"bufio"
	"io"
	"strconv"
	"strings"
)

type ProgressUpdate struct {
	Percent        int
	CurrentSeconds float64
	TotalSeconds   float64
	Speed          string
	FPS            float64
	OutBytes       int64
	Done           bool
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
			if us, err := strconv.ParseFloat(val, 64); err == nil {
				current.CurrentSeconds = us / 1_000_000.0
				if totalDurationSec > 0 {
					pct := int((current.CurrentSeconds / totalDurationSec) * 100)
					if pct > 100 {
						pct = 100
					} else if pct < 0 {
						pct = 0
					}
					current.Percent = pct
				}
			}
		case "total_size":
			if sz, err := strconv.ParseInt(val, 10, 64); err == nil {
				current.OutBytes = sz
			}
		case "speed":
			current.Speed = val
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
