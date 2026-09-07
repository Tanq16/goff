package ops

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
)

type ScaleTier struct {
	Aliases []string
	Width   int
	Height  int
	Suffix  string
}

var ScaleTiers = []ScaleTier{
	{Aliases: []string{"4k", "2160p"}, Width: 3840, Height: 2160, Suffix: "4k"},
	{Aliases: []string{"1440p", "2k"}, Width: 2560, Height: 1440, Suffix: "1440p"},
	{Aliases: []string{"1080p", "fhd"}, Width: 1920, Height: 1080, Suffix: "1080p"},
	{Aliases: []string{"720p", "hd"}, Width: 1280, Height: 720, Suffix: "720p"},
	{Aliases: []string{"480p", "sd"}, Width: 854, Height: 480, Suffix: "480p"},
}

var scaleDimensions = regexp.MustCompile(`^[1-9][0-9]*x[1-9][0-9]*$`)

func ScaleAliases() []string {
	var aliases []string
	for _, t := range ScaleTiers {
		aliases = append(aliases, t.Aliases...)
	}
	return aliases
}

func LookupScaleTier(v string) (ScaleTier, bool) {
	lower := strings.ToLower(strings.TrimSpace(v))
	for _, t := range ScaleTiers {
		if slices.Contains(t.Aliases, lower) {
			return t, true
		}
	}
	return ScaleTier{}, false
}

func ValidScaleTarget(v string) bool {
	if _, ok := LookupScaleTier(v); ok {
		return true
	}
	return scaleDimensions.MatchString(strings.ToLower(strings.TrimSpace(v)))
}

func FitClause(maxHeight int) string {
	if maxHeight <= 0 {
		maxHeight = 1080
	}
	return fmt.Sprintf("scale='min(%d,iw)':'min(%d,ih)':force_original_aspect_ratio=decrease,scale=trunc(iw/2)*2:trunc(ih/2)*2", maxHeight*16/9, maxHeight)
}

func ToneMapFilter(maxHeight int) string {
	return FitClause(maxHeight) + ",format=gbrpf32le,tonemap=hable:desat=0.5,format=yuv420p"
}
