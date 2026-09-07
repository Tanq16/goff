package probe

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
)

const (
	SelectAll  = "all"
	SelectNone = "none"
)

var nonTextSubtitleCodecs = []string{
	"hdmv_pgs_subtitle",
	"hdmv_text_subtitle",
	"dvd_subtitle",
	"dvb_subtitle",
	"dvb_teletext",
	"xsub",
	"arib_caption",
	"ivtv_vbi",
	"eia_608",
}

type TrackSelector struct {
	raw      string
	all      bool
	none     bool
	hasIndex bool
	index    int
	language string
}

func ParseTrackSelector(v string) (TrackSelector, error) {
	trimmed := strings.ToLower(strings.TrimSpace(v))
	switch trimmed {
	case SelectAll:
		return TrackSelector{raw: trimmed, all: true}, nil
	case SelectNone:
		return TrackSelector{raw: trimmed, none: true}, nil
	}

	malformed := fmt.Errorf("must be all, none, a stream index, or a language code like eng")
	if trimmed == "" {
		return TrackSelector{}, malformed
	}
	if index, err := strconv.Atoi(trimmed); err == nil {
		if index < 0 {
			return TrackSelector{}, fmt.Errorf("a stream index cannot be negative")
		}
		return TrackSelector{raw: trimmed, hasIndex: true, index: index}, nil
	}
	if len(trimmed) < 2 || len(trimmed) > 3 || strings.Trim(trimmed, "abcdefghijklmnopqrstuvwxyz") != "" {
		return TrackSelector{}, malformed
	}
	return TrackSelector{raw: trimmed, language: trimmed}, nil
}

func AllTracks() TrackSelector { return TrackSelector{raw: SelectAll, all: true} }

func (t TrackSelector) String() string { return t.raw }

func (p *ProbeResult) SelectAudio(sel TrackSelector) ([]StreamInfo, error) {
	if p == nil {
		return nil, nil
	}
	return selectStreams(p.AudioStreams(), "audio", sel)
}

func (p *ProbeResult) SelectSubtitles(sel TrackSelector) ([]StreamInfo, error) {
	if p == nil {
		return nil, nil
	}
	return selectStreams(p.SubtitleStreams(), "subtitle", sel)
}

func selectStreams(pool []StreamInfo, kind string, sel TrackSelector) ([]StreamInfo, error) {
	switch {
	case sel.none:
		return nil, nil
	case sel.all:
		return pool, nil
	case sel.hasIndex:
		for _, s := range pool {
			if s.Index == sel.index {
				return []StreamInfo{s}, nil
			}
		}
		return nil, fmt.Errorf("stream %d is not one of this file's %s streams", sel.index, kind)
	}

	var picked []StreamInfo
	for _, s := range pool {
		if strings.EqualFold(s.Language(), sel.language) {
			picked = append(picked, s)
		}
	}
	if len(picked) == 0 {
		return nil, fmt.Errorf("no %s stream is tagged %s", kind, sel.language)
	}
	return picked, nil
}

func DefaultFirst(streams []StreamInfo) []StreamInfo {
	for i, s := range streams {
		if s.Disposition["default"] != 1 {
			continue
		}
		ordered := make([]StreamInfo, 0, len(streams))
		ordered = append(ordered, s)
		ordered = append(ordered, streams[:i]...)
		return append(ordered, streams[i+1:]...)
	}
	return streams
}

func IsTextSubtitle(codecName string) bool {
	return !slices.Contains(nonTextSubtitleCodecs, strings.ToLower(codecName))
}

func (s *StreamInfo) Language() string {
	return tagValue(s.Tags, "language")
}

func (s *StreamInfo) Title() string {
	return tagValue(s.Tags, "title")
}

func tagValue(tags map[string]string, key string) string {
	for k, v := range tags {
		if strings.EqualFold(k, key) {
			return v
		}
	}
	return ""
}
