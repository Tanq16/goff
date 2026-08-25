package ops

import (
	"fmt"
	"strconv"
	"strings"
)

type AudioSource struct {
	Path    string
	DelayMS int
	Volume  float64
}

type MultiMixOpts struct {
	Sources []AudioSource
	Fit     string
	Format  string
	Bitrate string
}

func amixDuration(fit string) string {
	if fit == "shortest" {
		return "shortest"
	}
	return "longest"
}

func mixFilter(labels []string, sources []AudioSource, fit string, outLabel string) string {
	var chains []string
	mixInputs := make([]string, 0, len(labels))

	for i, label := range labels {
		var filters []string
		if sources[i].DelayMS > 0 {
			filters = append(filters, fmt.Sprintf("adelay=%d:all=1", sources[i].DelayMS))
		}
		if sources[i].Volume > 0 && sources[i].Volume != 1.0 {
			filters = append(filters, "volume="+strconv.FormatFloat(sources[i].Volume, 'g', -1, 64))
		}
		if len(filters) == 0 {
			mixInputs = append(mixInputs, label)
			continue
		}
		staged := fmt.Sprintf("[mix%d]", i)
		chains = append(chains, label+strings.Join(filters, ",")+staged)
		mixInputs = append(mixInputs, staged)
	}

	chains = append(chains, fmt.Sprintf("%samix=inputs=%d:duration=%s:normalize=0%s",
		strings.Join(mixInputs, ""), len(mixInputs), amixDuration(fit), outLabel))

	return strings.Join(chains, ";")
}

func BuildMultiMix(opts MultiMixOpts) (*OpResult, error) {
	if len(opts.Sources) < 2 {
		return nil, fmt.Errorf("mixing requires at least 2 audio inputs")
	}

	var args []string
	labels := make([]string, 0, len(opts.Sources))
	for i, s := range opts.Sources {
		args = append(args, "-i", s.Path)
		labels = append(labels, fmt.Sprintf("[%d:a]", i))
	}

	encodeArgs, ext := audioEncodeArgs(opts.Format, opts.Bitrate)
	args = append(args, "-filter_complex", mixFilter(labels, opts.Sources, opts.Fit, "[aout]"), "-map", "[aout]")
	args = append(args, encodeArgs...)

	return &OpResult{
		Args:      args,
		Suffix:    "mixed",
		TargetExt: ext,
	}, nil
}
