package probe

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type ProbeResult struct {
	Format  FormatInfo   `json:"format"`
	Streams []StreamInfo `json:"streams"`
}

type FormatInfo struct {
	Filename       string            `json:"filename"`
	NbStreams      int               `json:"nb_streams"`
	FormatName     string            `json:"format_name"`
	FormatLongName string            `json:"format_long_name"`
	DurationStr    string            `json:"duration"`
	SizeStr        string            `json:"size"`
	BitRateStr     string            `json:"bit_rate"`
	Tags           map[string]string `json:"tags"`
}

type StreamInfo struct {
	Index            int               `json:"index"`
	CodecType        string            `json:"codec_type"`
	CodecName        string            `json:"codec_name"`
	CodecLongName    string            `json:"codec_long_name"`
	Profile          string            `json:"profile"`
	Width            int               `json:"width"`
	Height           int               `json:"height"`
	PixFmt           string            `json:"pix_fmt"`
	AvgFrameRate     string            `json:"avg_frame_rate"`
	RFrameRate       string            `json:"r_frame_rate"`
	SampleRate       string            `json:"sample_rate"`
	Channels         int               `json:"channels"`
	ChannelLayout    string            `json:"channel_layout"`
	BitsPerSample    int               `json:"bits_per_sample"`
	BitsPerRawSample string            `json:"bits_per_raw_sample"`
	ColorSpace       string            `json:"color_space"`
	ColorTransfer    string            `json:"color_transfer"`
	ColorPrimaries   string            `json:"color_primaries"`
	Duration         string            `json:"duration"`
	BitRate          string            `json:"bit_rate"`
	Tags             map[string]string `json:"tags"`
	Disposition      map[string]int    `json:"disposition"`
}

func RunProbe(ctx context.Context, filePath string) (*ProbeResult, error) {
	args := []string{
		"-v", "error",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		filePath,
	}

	cmd := exec.CommandContext(ctx, "ffprobe", args...)
	var stdout bytes.Buffer
	var stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if detail := strings.TrimSpace(stderr.String()); detail != "" {
			return nil, fmt.Errorf("ffprobe failed on %q: %s (%w)", filePath, detail, err)
		}
		return nil, fmt.Errorf("ffprobe failed on %q: %w", filePath, err)
	}

	var result ProbeResult
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		return nil, fmt.Errorf("failed to decode ffprobe json: %w", err)
	}

	return &result, nil
}

func (f *FormatInfo) Duration() float64 {
	d, _ := strconv.ParseFloat(f.DurationStr, 64)
	return d
}

func (f *FormatInfo) Size() int64 {
	s, _ := strconv.ParseInt(f.SizeStr, 10, 64)
	return s
}

func (f *FormatInfo) BitRate() int64 {
	if br, err := strconv.ParseInt(f.BitRateStr, 10, 64); err == nil && br > 0 {
		return br
	}
	if d := f.Duration(); d > 0 {
		return int64(float64(f.Size()) * 8 / d)
	}
	return 0
}

func ParseFPS(fpsStr string) float64 {
	if fpsStr == "" || fpsStr == "0/0" {
		return 0
	}
	if num, den, ok := strings.Cut(fpsStr, "/"); ok {
		n, err1 := strconv.ParseFloat(num, 64)
		d, err2 := strconv.ParseFloat(den, 64)
		if err1 == nil && err2 == nil && d > 0 {
			return n / d
		}
	}
	val, _ := strconv.ParseFloat(fpsStr, 64)
	return val
}
