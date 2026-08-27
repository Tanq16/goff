package ops

import (
	"fmt"
	"strings"

	"github.com/Tanq16/goff/internal/probe"
)

type VideoTransformOpts struct {
	Scale        string
	CropVertical bool
	Rotate       string
	Speed        float64
	StripAudio   bool
	CustomSuffix string
}

func BuildVideoTransform(inputPath string, p *probe.ProbeResult, opts VideoTransformOpts) (*OpResult, error) {
	var vfFilters []string
	var afFilters []string
	suffix := "transformed"

	if opts.Rotate != "" {
		switch strings.ToLower(opts.Rotate) {
		case "90_cw", "90", "cw", "transpose=1":
			vfFilters = append(vfFilters, "transpose=1")
			suffix = "rot90"
		case "90_ccw", "270", "ccw", "transpose=2":
			vfFilters = append(vfFilters, "transpose=2")
			suffix = "rot270"
		case "180":
			vfFilters = append(vfFilters, "transpose=1,transpose=1")
			suffix = "rot180"
		case "hflip", "flip_h", "mirror":
			vfFilters = append(vfFilters, "hflip")
			suffix = "hflip"
		case "vflip", "flip_v":
			vfFilters = append(vfFilters, "vflip")
			suffix = "vflip"
		default:
			vfFilters = append(vfFilters, opts.Rotate)
			suffix = "rotated"
		}
	}

	if opts.CropVertical {
		vfFilters = append(vfFilters, "crop=ih*(9/16):ih,scale=trunc(iw/2)*2:trunc(ih/2)*2")
		suffix = "vertical"
	}

	if opts.Scale != "" {
		scaleFilter := fmt.Sprintf("scale=%s", opts.Scale)
		suffix = "scaled"
		if tier, ok := LookupScaleTier(opts.Scale); ok {
			scaleFilter = fmt.Sprintf("scale='min(%d,iw)':'min(%d,ih)':force_original_aspect_ratio=decrease", tier.Width, tier.Height)
			suffix = tier.Suffix
		}
		vfFilters = append(vfFilters, scaleFilter+",scale=trunc(iw/2)*2:trunc(ih/2)*2")
	}

	retimeAudio := false
	if opts.Speed > 0 && opts.Speed != 1.0 {
		suffix = fmt.Sprintf("%.2fx", opts.Speed)
		vfFilters = append(vfFilters, fmt.Sprintf("setpts=%f*PTS", 1.0/opts.Speed))
		if !opts.StripAudio && p != nil && len(p.AudioStreams()) > 0 {
			retimeAudio = true
			afFilters = append(afFilters, fmt.Sprintf("atempo=%f", opts.Speed))
		}
	}

	var args []string
	args = append(args, "-i", inputPath)

	if retimeAudio {
		args = append(args, "-filter_complex", fmt.Sprintf("[0:v]%s[v];[0:a]%s[a]",
			strings.Join(vfFilters, ","), strings.Join(afFilters, ",")), "-map", "[v]", "-map", "[a]")
	} else {
		if len(vfFilters) > 0 {
			args = append(args, "-vf", strings.Join(vfFilters, ","))
		}
		if len(afFilters) > 0 {
			args = append(args, "-af", strings.Join(afFilters, ","))
		}
	}

	args = append(args, "-c:v", "libx264", "-crf", "20", "-preset", "medium")

	if opts.StripAudio {
		args = append(args, "-an")
		suffix = "muted"
	} else {
		args = append(args, "-c:a", "aac", "-b:a", "192k")
	}

	args = append(args, "-movflags", "+faststart")

	if opts.Rotate != "" {
		args = append(args, "-metadata:s:v:0", "rotate=0")
	}

	if opts.CustomSuffix != "" {
		suffix = opts.CustomSuffix
	}

	return &OpResult{
		Args:      args,
		Suffix:    suffix,
		TargetExt: "mp4",
	}, nil
}
