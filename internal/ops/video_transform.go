package ops

import (
	"fmt"
	"strings"

	"github.com/Tanq16/goff/internal/probe"
)

type VideoTransformOpts struct {
	Scale         string
	CropVertical  bool
	Rotate        string
	Speed         float64
	StripAudio    bool
	VolumeBoost   float64
	DownmixStereo bool
	CustomSuffix  string
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
		var scaleFilter string
		switch strings.ToLower(opts.Scale) {
		case "4k", "2160p":
			scaleFilter = "scale='min(3840,iw)':'min(2160,ih)':force_original_aspect_ratio=decrease"
			suffix = "4k"
		case "1440p", "2k":
			scaleFilter = "scale='min(2560,iw)':'min(1440,ih)':force_original_aspect_ratio=decrease"
			suffix = "1440p"
		case "1080p", "fhd":
			scaleFilter = "scale='min(1920,iw)':'min(1080,ih)':force_original_aspect_ratio=decrease"
			suffix = "1080p"
		case "720p", "hd":
			scaleFilter = "scale='min(1280,iw)':'min(720,ih)':force_original_aspect_ratio=decrease"
			suffix = "720p"
		case "480p", "sd":
			scaleFilter = "scale='min(854,iw)':'min(480,ih)':force_original_aspect_ratio=decrease"
			suffix = "480p"
		default:
			scaleFilter = fmt.Sprintf("scale=%s", opts.Scale)
			suffix = "scaled"
		}
		vfFilters = append(vfFilters, scaleFilter+",scale=trunc(iw/2)*2:trunc(ih/2)*2")
	}

	hasComplexSpeed := false
	var filterComplex string
	if opts.Speed > 0 && opts.Speed != 1.0 {
		ptsFactor := 1.0 / opts.Speed
		suffix = fmt.Sprintf("%.2fx", opts.Speed)
		if !opts.StripAudio && p != nil && len(p.AudioStreams()) > 0 {
			hasComplexSpeed = true
			filterComplex = fmt.Sprintf("[0:v]setpts=%f*PTS[v];[0:a]atempo=%f[a]", ptsFactor, opts.Speed)
		} else {
			vfFilters = append(vfFilters, fmt.Sprintf("setpts=%f*PTS", ptsFactor))
		}
	}

	if opts.VolumeBoost > 0 && opts.VolumeBoost != 1.0 {
		afFilters = append(afFilters, fmt.Sprintf("volume=%f", opts.VolumeBoost))
		if suffix == "transformed" {
			suffix = "boosted"
		}
	}

	if opts.DownmixStereo {
		afFilters = append(afFilters, "pan=stereo|FL=0.5*FC+0.707*FL+0.707*BL+0.5*LFE|FR=0.5*FC+0.707*FR+0.707*BR+0.5*LFE")
		if suffix == "transformed" {
			suffix = "stereo"
		}
	}

	var args []string
	args = append(args, "-i", inputPath)

	if hasComplexSpeed {
		args = append(args, "-filter_complex", filterComplex, "-map", "[v]", "-map", "[a]")
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
	} else if !hasComplexSpeed {
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
