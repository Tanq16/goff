package cmd

import (
	"fmt"
	"math"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/Tanq16/goff/internal/ops"
	"github.com/Tanq16/goff/internal/probe"
)

var sharedFlags struct {
	output string
	jobs   int
}

func addOutputFlag(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&sharedFlags.output, "output", "o", "", "Explicit output path, overwritten if it exists (single input only)")
}

func addJobsFlag(cmd *cobra.Command) {
	cmd.Flags().IntVarP(&sharedFlags.jobs, "jobs", "j", 2, "Concurrent encodes when several inputs are given")
}

type enumFlag struct {
	target  *string
	allowed []string
}

func newEnum(target *string, def string, allowed ...string) *enumFlag {
	*target = def
	return &enumFlag{target: target, allowed: allowed}
}

func (e *enumFlag) String() string { return *e.target }

func (e *enumFlag) Set(v string) error {
	lower := strings.ToLower(v)
	if !slices.Contains(e.allowed, lower) {
		return fmt.Errorf("must be one of: %s", strings.Join(e.allowed, ", "))
	}
	*e.target = lower
	return nil
}

func (e *enumFlag) Type() string { return strings.Join(e.allowed, "|") }

type trackFlag struct {
	target *probe.TrackSelector
}

func newTrack(target *probe.TrackSelector, def probe.TrackSelector) *trackFlag {
	*target = def
	return &trackFlag{target: target}
}

func (t *trackFlag) String() string { return t.target.String() }

func (t *trackFlag) Set(v string) error {
	sel, err := probe.ParseTrackSelector(v)
	if err != nil {
		return err
	}
	*t.target = sel
	return nil
}

func (t *trackFlag) Type() string { return "all|none|index|lang" }

type oneTrackFlag struct {
	target *probe.TrackSelector
}

func newOneTrack(target *probe.TrackSelector) *oneTrackFlag {
	return &oneTrackFlag{target: target}
}

func (t *oneTrackFlag) String() string { return t.target.String() }

func (t *oneTrackFlag) Set(v string) error {
	sel, err := probe.ParseTrackSelector(v)
	if err != nil {
		return err
	}
	if raw := sel.String(); raw == probe.SelectAll || raw == probe.SelectNone {
		return fmt.Errorf("must name a single track, so %s is not allowed", raw)
	}
	*t.target = sel
	return nil
}

func (t *oneTrackFlag) Type() string { return "index|lang" }

type scaleFlag struct {
	target *string
}

func newScale(target *string) *scaleFlag {
	return &scaleFlag{target: target}
}

func (s *scaleFlag) String() string { return *s.target }

func (s *scaleFlag) Set(v string) error {
	if !ops.ValidScaleTarget(v) {
		return fmt.Errorf("must be a tier (%s) or WxH like 1280x720", strings.Join(ops.ScaleAliases(), ", "))
	}
	*s.target = strings.ToLower(v)
	return nil
}

func (s *scaleFlag) Type() string { return "tier|WxH" }

type boundedFloat struct {
	target *float64
	min    float64
	max    float64
}

func newBoundedFloat(target *float64, def, lo, hi float64) *boundedFloat {
	*target = def
	return &boundedFloat{target: target, min: lo, max: hi}
}

func formatFloat(f float64) string { return strconv.FormatFloat(f, 'g', -1, 64) }

func (b *boundedFloat) String() string { return formatFloat(*b.target) }

func (b *boundedFloat) Set(v string) error {
	f, err := strconv.ParseFloat(v, 64)
	if err != nil || math.IsNaN(f) {
		return fmt.Errorf("must be a number")
	}
	if f < b.min || f > b.max {
		return fmt.Errorf("must be between %s and %s", formatFloat(b.min), formatFloat(b.max))
	}
	*b.target = f
	return nil
}

func (b *boundedFloat) Type() string { return formatFloat(b.min) + ".." + formatFloat(b.max) }

type boundedInt struct {
	target *int
	min    int
	max    int
}

func newBoundedInt(target *int, def, lo, hi int) *boundedInt {
	*target = def
	return &boundedInt{target: target, min: lo, max: hi}
}

func (b *boundedInt) String() string { return strconv.Itoa(*b.target) }

func (b *boundedInt) Set(v string) error {
	n, err := strconv.Atoi(v)
	if err != nil {
		return fmt.Errorf("must be a whole number")
	}
	if n < b.min || n > b.max {
		return fmt.Errorf("must be between %d and %d", b.min, b.max)
	}
	*b.target = n
	return nil
}

func (b *boundedInt) Type() string { return strconv.Itoa(b.min) + ".." + strconv.Itoa(b.max) }

type sizeFlag struct {
	raw    *string
	budget *float64
}

func newSize(raw *string, budget *float64) *sizeFlag {
	return &sizeFlag{raw: raw, budget: budget}
}

func (s *sizeFlag) String() string { return *s.raw }

func (s *sizeFlag) Set(v string) error {
	budget, err := ops.ParseSizeBudget(v)
	if err != nil {
		return err
	}
	*s.raw = strings.ToLower(strings.TrimSpace(v))
	*s.budget = budget
	return nil
}

func (s *sizeFlag) Type() string { return "size" }

var audioSampleRates = []int{8000, 11025, 12000, 16000, 22050, 24000, 32000, 44100, 48000, 64000, 88200, 96000}

type sampleRateFlag struct {
	target *int
}

func newSampleRate(target *int, def int) *sampleRateFlag {
	*target = def
	return &sampleRateFlag{target: target}
}

func (s *sampleRateFlag) String() string { return strconv.Itoa(*s.target) }

func (s *sampleRateFlag) Set(v string) error {
	n, err := strconv.Atoi(strings.TrimSpace(v))
	if err != nil || !slices.Contains(audioSampleRates, n) {
		allowed := make([]string, 0, len(audioSampleRates))
		for _, rate := range audioSampleRates {
			allowed = append(allowed, strconv.Itoa(rate))
		}
		return fmt.Errorf("must be one of: %s", strings.Join(allowed, ", "))
	}
	*s.target = n
	return nil
}

func (s *sampleRateFlag) Type() string { return "Hz" }

var bitratePattern = regexp.MustCompile(`^[1-9][0-9]*k?$`)

type bitrateFlag struct {
	target *string
}

func newBitrate(target *string) *bitrateFlag { return &bitrateFlag{target: target} }

func (b *bitrateFlag) String() string { return *b.target }

func (b *bitrateFlag) Set(v string) error {
	lower := strings.ToLower(strings.TrimSpace(v))
	if !bitratePattern.MatchString(lower) {
		return fmt.Errorf("must look like 320k or 320000")
	}
	*b.target = lower
	return nil
}

func (b *bitrateFlag) Type() string { return "rate" }

func parseTimestamp(v string) (float64, error) {
	malformed := fmt.Errorf("must be a timestamp like 00:01:30 or 90")
	parts := strings.Split(strings.TrimSpace(v), ":")
	if len(parts) > 3 {
		return 0, malformed
	}
	total := 0.0
	for i, part := range parts {
		n, err := strconv.ParseFloat(part, 64)
		if err != nil || n < 0 || math.IsInf(n, 0) {
			return 0, malformed
		}
		if i < len(parts)-1 && n != math.Trunc(n) {
			return 0, fmt.Errorf("only the seconds component may be fractional")
		}
		total = total*60 + n
	}
	return total, nil
}

type timestampFlag struct {
	target  *string
	seconds *float64
}

func newTimestamp(target *string) *timestampFlag { return &timestampFlag{target: target} }

func newTimestampSeconds(target *string, seconds *float64) *timestampFlag {
	return &timestampFlag{target: target, seconds: seconds}
}

func (t *timestampFlag) String() string { return *t.target }

func (t *timestampFlag) Set(v string) error {
	seconds, err := parseTimestamp(v)
	if err != nil {
		return err
	}
	*t.target = v
	if t.seconds != nil {
		*t.seconds = seconds
	}
	return nil
}

func (t *timestampFlag) Type() string { return "time" }

var mediaOptionPattern = regexp.MustCompile(`:(at|vol)=`)

type mediaInput struct {
	Path    string
	DelayMS int
	Volume  float64
}

func parseMediaInput(spec string) (mediaInput, error) {
	in := mediaInput{Path: spec, Volume: 1.0}
	matches := mediaOptionPattern.FindAllStringSubmatchIndex(spec, -1)
	if len(matches) > 0 {
		in.Path = spec[:matches[0][0]]
	}
	if in.Path == "" {
		return in, fmt.Errorf("no input path before the options")
	}
	for i, m := range matches {
		end := len(spec)
		if i+1 < len(matches) {
			end = matches[i+1][0]
		}
		value := spec[m[1]:end]
		switch spec[m[2]:m[3]] {
		case "at":
			seconds, err := parseTimestamp(value)
			if err != nil {
				return in, fmt.Errorf("at=%s: %w", value, err)
			}
			in.DelayMS = int(math.Round(seconds * 1000))
		case "vol":
			vol, err := strconv.ParseFloat(value, 64)
			if err != nil || vol <= 0 || vol > 10 {
				return in, fmt.Errorf("vol=%s must be greater than 0 and at most 10", value)
			}
			in.Volume = vol
		}
	}
	return in, nil
}

type mediaInputSlice struct {
	target *[]mediaInput
}

func newMediaInputs(target *[]mediaInput) *mediaInputSlice {
	return &mediaInputSlice{target: target}
}

func (m *mediaInputSlice) String() string {
	paths := make([]string, 0, len(*m.target))
	for _, in := range *m.target {
		paths = append(paths, in.Path)
	}
	return strings.Join(paths, ",")
}

func (m *mediaInputSlice) Set(v string) error {
	in, err := parseMediaInput(v)
	if err != nil {
		return err
	}
	*m.target = append(*m.target, in)
	return nil
}

func (m *mediaInputSlice) Type() string { return "file[:at=t][:vol=n]" }
