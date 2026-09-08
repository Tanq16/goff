package subs

import (
	"os"
	"regexp"
	"strings"
)

var (
	markupTag     = regexp.MustCompile(`<[^>]*>`)
	overrideBlock = regexp.MustCompile(`\{[^}]*\}`)
)

func CleanVTT(path string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	blocks := strings.Split(strings.ReplaceAll(string(raw), "\r\n", "\n"), "\n\n")
	kept := make([]string, 0, len(blocks))
	seen := make(map[string]bool, len(blocks))
	for _, block := range blocks {
		trimmed := strings.TrimSpace(block)
		if trimmed == "" {
			continue
		}
		if !strings.Contains(trimmed, "-->") {
			kept = append(kept, trimmed)
			continue
		}
		cue, ok := cleanCue(trimmed)
		if !ok || seen[cue] {
			continue
		}
		seen[cue] = true
		kept = append(kept, cue)
	}

	return os.WriteFile(path, []byte(strings.Join(kept, "\n\n")+"\n\n"), 0o644)
}

func cleanCue(block string) (string, bool) {
	var out []string
	var hasText, timed bool
	for line := range strings.SplitSeq(block, "\n") {
		switch {
		case !timed && strings.Contains(line, "-->"):
			timed = true
			out = append(out, normalizeCueTiming(line))
		case !timed:
			out = append(out, line)
		default:
			stripped := overrideBlock.ReplaceAllString(markupTag.ReplaceAllString(line, ""), "")
			if stripped != line {
				stripped = strings.TrimSpace(stripped)
			}
			if stripped == "" {
				continue
			}
			hasText = true
			out = append(out, stripped)
		}
	}
	return strings.Join(out, "\n"), hasText
}

func normalizeCueTiming(line string) string {
	start, rest, ok := strings.Cut(line, "-->")
	if !ok {
		return line
	}
	end, settings, hasSettings := strings.Cut(strings.TrimSpace(rest), " ")
	timing := normalizeStamp(strings.TrimSpace(start)) + " --> " + normalizeStamp(end)
	if hasSettings {
		return timing + " " + settings
	}
	return timing
}

func normalizeStamp(stamp string) string {
	if strings.Count(stamp, ":") == 1 {
		return "00:" + stamp
	}
	return stamp
}
