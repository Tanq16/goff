package probe

import (
	"cmp"
	"fmt"
	"path/filepath"
	"strconv"
)

type Summary struct {
	File        string          `json:"file"`
	Duration    string          `json:"duration"`
	Size        string          `json:"size"`
	Bitrate     string          `json:"bitrate"`
	Format      string          `json:"format"`
	Streams     []StreamSummary `json:"streams"`
	Conformance *Conformance    `json:"conformance,omitempty"`
}

type StreamSummary struct {
	Index    int    `json:"index"`
	Type     string `json:"type"`
	Codec    string `json:"codec"`
	Details  string `json:"details"`
	Bitrate  string `json:"bitrate"`
	Language string `json:"language"`
	Title    string `json:"title"`
	Default  bool   `json:"default"`
	Forced   bool   `json:"forced"`
}

func (p *ProbeResult) Summarize(path string) Summary {
	s := Summary{
		File:     filepath.Base(path),
		Duration: p.HumanDuration(),
		Size:     p.HumanSize(),
		Bitrate:  p.HumanBitRate(),
		Format:   p.Format.FormatLongName,
	}
	for _, stream := range p.Streams {
		s.Streams = append(s.Streams, p.summarizeStream(stream))
	}
	return s
}

func (p *ProbeResult) summarizeStream(s StreamInfo) StreamSummary {
	codec := s.CodecName
	if s.Profile != "" {
		codec = fmt.Sprintf("%s (%s)", s.CodecName, s.Profile)
	}

	var details string
	switch s.CodecType {
	case "video":
		hdrTag := ""
		if p.IsHDR() {
			hdrTag = fmt.Sprintf(" [%s]", p.HDRType())
		}
		details = fmt.Sprintf("%dx%d @ %.2ffps, %s%s", s.Width, s.Height, ParseFPS(s.AvgFrameRate), s.PixFmt, hdrTag)
	case "audio":
		details = fmt.Sprintf("%sHz, %d ch (%s), %s", s.SampleRate, s.Channels, s.ChannelLayout, trackLabel(s))
	case "subtitle":
		details = trackLabel(s)
	default:
		details = s.CodecLongName
	}

	bitrate := "-"
	if s.BitRate != "" {
		bitrate = s.BitRate
		if br, err := strconv.ParseInt(s.BitRate, 10, 64); err == nil {
			bitrate = fmt.Sprintf("%d kbps", br/1000)
		}
	}

	return StreamSummary{
		Index:    s.Index,
		Type:     s.CodecType,
		Codec:    codec,
		Details:  details,
		Bitrate:  bitrate,
		Language: s.Language(),
		Title:    s.Title(),
		Default:  s.Disposition["default"] == 1,
		Forced:   s.Disposition["forced"] == 1,
	}
}

func trackLabel(s StreamInfo) string {
	label := cmp.Or(s.Language(), "und")
	if title := s.Title(); title != "" {
		label = fmt.Sprintf("%s (%s)", label, title)
	}
	if s.Disposition["forced"] == 1 {
		label += ", forced"
	}
	return label
}
