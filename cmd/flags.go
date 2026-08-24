package cmd

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
)

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

type scaleFlag struct {
	target *string
}

func newScale(target *string) *scaleFlag {
	return &scaleFlag{target: target}
}

func (s *scaleFlag) String() string { return *s.target }

func (s *scaleFlag) Set(v string) error {
	if !validScaleTarget(v) {
		return fmt.Errorf("must be a tier (2160p, 1440p, 1080p, 720p, 480p) or WxH like 1280x720")
	}
	*s.target = strings.ToLower(v)
	return nil
}

func (s *scaleFlag) Type() string { return "tier|WxH" }

type boundedFloat struct {
	target *float64
	min    float64
	max    float64
	bounds string
}

func newBoundedFloat(target *float64, def, min, max float64, bounds string) *boundedFloat {
	*target = def
	return &boundedFloat{target: target, min: min, max: max, bounds: bounds}
}

func (b *boundedFloat) String() string {
	return strconv.FormatFloat(*b.target, 'g', -1, 64)
}

func (b *boundedFloat) Set(v string) error {
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return fmt.Errorf("must be a number")
	}
	if f <= b.min || f > b.max {
		return fmt.Errorf("must be %s", b.bounds)
	}
	*b.target = f
	return nil
}

func (b *boundedFloat) Type() string { return "float" }
