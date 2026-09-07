package probe

import (
	"fmt"
	"strings"
	"time"
)

func (p *ProbeResult) VideoStreams() []StreamInfo {
	var list []StreamInfo
	for _, s := range p.Streams {
		if strings.EqualFold(s.CodecType, "video") {
			list = append(list, s)
		}
	}
	return list
}

func (p *ProbeResult) AudioStreams() []StreamInfo {
	var list []StreamInfo
	for _, s := range p.Streams {
		if strings.EqualFold(s.CodecType, "audio") {
			list = append(list, s)
		}
	}
	return list
}

func (p *ProbeResult) SubtitleStreams() []StreamInfo {
	var list []StreamInfo
	for _, s := range p.Streams {
		if strings.EqualFold(s.CodecType, "subtitle") {
			list = append(list, s)
		}
	}
	return list
}

func (p *ProbeResult) AttachmentStreams() []StreamInfo {
	var list []StreamInfo
	for _, s := range p.Streams {
		if strings.EqualFold(s.CodecType, "attachment") {
			list = append(list, s)
		}
	}
	return list
}

func (p *ProbeResult) PrimaryVideoStream() *StreamInfo {
	videos := p.VideoStreams()
	if len(videos) == 0 {
		return nil
	}
	for i := range videos {
		if videos[i].Disposition["default"] == 1 {
			return &videos[i]
		}
	}
	return &videos[0]
}

func (p *ProbeResult) PrimaryAudioStream() *StreamInfo {
	audios := p.AudioStreams()
	if len(audios) == 0 {
		return nil
	}
	for i := range audios {
		if audios[i].Disposition["default"] == 1 {
			return &audios[i]
		}
	}
	return &audios[0]
}

func (p *ProbeResult) IsVideo() bool {
	return len(p.VideoStreams()) > 0
}

func (p *ProbeResult) TotalDuration() float64 {
	if d := p.Format.Duration(); d > 0 {
		return d
	}
	if v := p.PrimaryVideoStream(); v != nil {
		if d := v.DurationSeconds(); d > 0 {
			return d
		}
	}
	if a := p.PrimaryAudioStream(); a != nil {
		if d := a.DurationSeconds(); d > 0 {
			return d
		}
	}
	return 0
}

func (s *StreamInfo) DurationSeconds() float64 {
	if s.Duration == "" {
		return 0
	}
	var sec float64
	fmt.Sscanf(s.Duration, "%f", &sec)
	return sec
}

func (p *ProbeResult) HumanSize() string {
	return FormatBytes(p.Format.Size())
}

func (p *ProbeResult) HumanDuration() string {
	return FormatDuration(p.TotalDuration())
}

func (p *ProbeResult) HumanBitRate() string {
	bps := p.Format.BitRate()
	if bps <= 0 {
		return "-"
	}
	return fmt.Sprintf("%d kbps", bps/1000)
}

func FormatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func FormatDuration(seconds float64) string {
	if seconds <= 0 {
		return "00:00"
	}
	d := time.Duration(seconds * float64(time.Second))
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	if h > 0 {
		return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%02d:%02d", m, s)
}
