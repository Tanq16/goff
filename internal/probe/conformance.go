package probe

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"slices"
	"strconv"
	"strings"
)

const DriftToleranceSeconds = 0.1

const gapPeriodTolerance = 1.5

type Conformance struct {
	CodecTag            string   `json:"codecTag"`
	VideoStartSeconds   float64  `json:"videoStartSeconds"`
	AudioStartSeconds   float64  `json:"audioStartSeconds"`
	StartOffsetSeconds  float64  `json:"startOffsetSeconds"`
	VideoContentSeconds float64  `json:"videoContentSeconds"`
	AudioContentSeconds float64  `json:"audioContentSeconds"`
	DriftSeconds        float64  `json:"driftSeconds"`
	VideoGapCount       int      `json:"videoGapCount"`
	VideoGapSeconds     float64  `json:"videoGapSeconds"`
	AudioGapCount       int      `json:"audioGapCount"`
	AudioGapSeconds     float64  `json:"audioGapSeconds"`
	BrowserSafe         bool     `json:"browserSafe"`
	Issues              []string `json:"issues"`
}

type timeline struct {
	Start      float64
	Content    float64
	GapCount   int
	GapSeconds float64
}

func (s *StreamInfo) TimeBaseSeconds() float64 {
	return ParseFPS(s.TimeBase)
}

func packetTimestamps(ctx context.Context, filePath string) (map[int][]int64, error) {
	args := []string{
		"-v", "error",
		"-show_entries", "packet=stream_index,pts",
		"-of", "csv=p=0",
		filePath,
	}

	cmd := exec.CommandContext(ctx, "ffprobe", args...)
	var stdout bytes.Buffer
	var stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if detail := strings.TrimSpace(stderr.String()); detail != "" {
			return nil, fmt.Errorf("ffprobe failed to read packets of %q: %s (%w)", filePath, detail, err)
		}
		return nil, fmt.Errorf("ffprobe failed to read packets of %q: %w", filePath, err)
	}

	byStream := make(map[int][]int64, 4)
	for line := range strings.SplitSeq(stdout.String(), "\n") {
		indexField, rest, ok := strings.Cut(strings.TrimSpace(line), ",")
		if !ok {
			continue
		}
		index, err := strconv.Atoi(indexField)
		if err != nil {
			continue
		}
		ptsField, _, _ := strings.Cut(rest, ",")
		pts, err := strconv.ParseInt(ptsField, 10, 64)
		if err != nil {
			continue
		}
		byStream[index] = append(byStream[index], pts)
	}
	for index := range byStream {
		slices.Sort(byStream[index])
	}
	return byStream, nil
}

func modalDelta(stamps []int64) float64 {
	if len(stamps) < 2 {
		return 0
	}
	counts := make(map[int64]int, 8)
	for i := range len(stamps) - 1 {
		counts[stamps[i+1]-stamps[i]]++
	}
	var best int64
	var bestCount int
	for delta, count := range counts {
		if count > bestCount || (count == bestCount && delta < best) {
			best, bestCount = delta, count
		}
	}
	return float64(best)
}

func measure(stamps []int64, expected, timeBase float64) timeline {
	var t timeline
	if len(stamps) < 2 || expected <= 0 || timeBase <= 0 {
		return t
	}
	t.Start = float64(stamps[0]) * timeBase
	t.Content = float64(len(stamps)) * expected * timeBase

	period := modalDelta(stamps)
	if period <= 0 {
		period = expected
	}
	for i := range len(stamps) - 1 {
		delta := float64(stamps[i+1] - stamps[i])
		if delta > period*gapPeriodTolerance {
			t.GapCount++
			t.GapSeconds += (delta - period) * timeBase
		}
	}
	return t
}

func Conform(ctx context.Context, filePath string, p *ProbeResult) (*Conformance, error) {
	c := &Conformance{}
	video := p.PrimaryVideoStream()
	audio := p.PrimaryAudioStream()

	var stamps map[int][]int64
	if video != nil || audio != nil {
		var err error
		stamps, err = packetTimestamps(ctx, filePath)
		if err != nil {
			return nil, err
		}
	}

	if video != nil {
		c.CodecTag = video.CodecTagString
		timeBase := video.TimeBaseSeconds()
		expected := modalDelta(stamps[video.Index])
		if fps := ParseFPS(video.RFrameRate); fps > 0 && timeBase > 0 {
			expected = 1 / (fps * timeBase)
		}
		t := measure(stamps[video.Index], expected, timeBase)
		c.VideoStartSeconds = t.Start
		c.VideoContentSeconds = t.Content
		c.VideoGapCount = t.GapCount
		c.VideoGapSeconds = t.GapSeconds
	}

	if audio != nil {
		t := measure(stamps[audio.Index], modalDelta(stamps[audio.Index]), audio.TimeBaseSeconds())
		c.AudioStartSeconds = t.Start
		c.AudioContentSeconds = t.Content
		c.AudioGapCount = t.GapCount
		c.AudioGapSeconds = t.GapSeconds
	}

	if c.VideoContentSeconds > 0 && c.AudioContentSeconds > 0 {
		c.DriftSeconds = c.VideoContentSeconds - c.AudioContentSeconds
		c.StartOffsetSeconds = c.VideoStartSeconds - c.AudioStartSeconds
	}

	c.Issues = p.browserIssues(c)
	c.BrowserSafe = len(c.Issues) == 0
	return c, nil
}

func (p *ProbeResult) browserIssues(c *Conformance) []string {
	issues := []string{}

	format := strings.ToLower(p.Format.FormatName)
	if !strings.Contains(format, "mp4") && !strings.Contains(format, "mov") {
		issues = append(issues, fmt.Sprintf("container is %s, not MP4", p.Format.FormatName))
	}

	video := p.PrimaryVideoStream()
	if video == nil {
		issues = append(issues, "no video stream")
	} else {
		codec := strings.ToLower(video.CodecName)
		if codec != "h264" && codec != "hevc" {
			issues = append(issues, fmt.Sprintf("video codec %s is not widely playable in browsers", video.CodecName))
		}
		if codec == "hevc" && !strings.EqualFold(video.CodecTagString, "hvc1") {
			issues = append(issues, fmt.Sprintf("HEVC is tagged %s, and Safari plays only hvc1", video.CodecTagString))
		}
		if !strings.EqualFold(video.PixFmt, "yuv420p") {
			issues = append(issues, fmt.Sprintf("pixel format %s is not 8-bit 4:2:0", video.PixFmt))
		}
		if c.VideoGapCount > 0 {
			issues = append(issues, fmt.Sprintf("video timeline has %d gaps totalling %.3fs, so it is not constant frame rate", c.VideoGapCount, c.VideoGapSeconds))
		}
	}

	audios := p.AudioStreams()
	for i, a := range audios {
		if a.Disposition["default"] != 1 {
			continue
		}
		if i > 0 {
			issues = append(issues, fmt.Sprintf("audio track %d carries the default flag while track %d comes first, so Chrome and Firefox play different tracks", a.Index, audios[0].Index))
		}
		break
	}

	if audio := p.PrimaryAudioStream(); audio != nil {
		if !strings.EqualFold(audio.CodecName, "aac") {
			issues = append(issues, fmt.Sprintf("audio codec %s is not AAC", audio.CodecName))
		}
		if audio.Channels > 2 {
			issues = append(issues, fmt.Sprintf("audio has %d channels rather than stereo", audio.Channels))
		}
		if audio.SampleRate != "48000" {
			issues = append(issues, fmt.Sprintf("audio sample rate is %sHz rather than 48000Hz", audio.SampleRate))
		}
		if c.AudioGapCount > 0 {
			issues = append(issues, fmt.Sprintf("audio timeline has %d gaps totalling %.3fs, which plays back as drift", c.AudioGapCount, c.AudioGapSeconds))
		}
	}

	if drift := c.DriftSeconds; drift > DriftToleranceSeconds || drift < -DriftToleranceSeconds {
		issues = append(issues, fmt.Sprintf("video and audio content differ by %.3fs", drift))
	}

	if offset := c.StartOffsetSeconds; offset > DriftToleranceSeconds || offset < -DriftToleranceSeconds {
		issues = append(issues, fmt.Sprintf("video and audio start %+.3fs apart, so playback is offset from the first frame", offset))
	}

	return issues
}
