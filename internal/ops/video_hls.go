package ops

import (
	"path/filepath"
	"strconv"

	"github.com/Tanq16/goff/internal/probe"
)

type HLSFormat string

const (
	HLSFormatFMP4   HLSFormat = "fmp4"
	HLSFormatMPEGTS HLSFormat = "mpegts"
)

type VideoHLSOpts struct {
	Format          HLSFormat
	SegmentDuration int
	AudioBitrate    string
	CRF             int
	GOPSize         int
	CustomSuffix    string
	OutputDir       string
}

func BuildVideoHLS(inputPath string, p *probe.ProbeResult, opts VideoHLSOpts) (*OpResult, error) {
	fmtType := opts.Format
	if fmtType == "" {
		fmtType = HLSFormatFMP4
	}

	segDur := opts.SegmentDuration
	if segDur <= 0 {
		segDur = 6
	}

	audioBitrate := opts.AudioBitrate
	if audioBitrate == "" {
		audioBitrate = "192k"
	}

	crf := opts.CRF
	if crf <= 0 {
		crf = 21
	}

	gop := opts.GOPSize
	if gop <= 0 {
		gop = 48
	}

	suffix := "hls"
	if opts.CustomSuffix != "" {
		suffix = opts.CustomSuffix
	} else if fmtType == HLSFormatFMP4 {
		suffix = "hls-fmp4"
	} else {
		suffix = "hls-ts"
	}

	args := []string{
		"-i", inputPath,
		"-c:v", "libx264",
		"-crf", strconv.Itoa(crf),
		"-preset", "medium",
		"-g", strconv.Itoa(gop),
		"-keyint_min", strconv.Itoa(gop),
		"-sc_threshold", "0",
		"-c:a", "aac",
		"-b:a", audioBitrate,
		"-ar", "48000",
		"-f", "hls",
		"-hls_time", strconv.Itoa(segDur),
		"-hls_playlist_type", "vod",
		"-hls_flags", "independent_segments",
	}

	segmentExt := "ts"
	if fmtType == HLSFormatFMP4 {
		segmentExt = "m4s"
		args = append(args,
			"-hls_segment_type", "fmp4",
			"-hls_fmp4_init_filename", "init.mp4",
		)
	} else {
		args = append(args,
			"-hls_segment_type", "mpegts",
		)
	}

	res := &OpResult{
		Args:      args,
		Suffix:    suffix,
		TargetExt: "m3u8",
	}

	if opts.OutputDir != "" {
		res.Args = append(res.Args, "-hls_segment_filename", filepath.Join(opts.OutputDir, "segment%04d."+segmentExt))
		res.OutputPath = filepath.Join(opts.OutputDir, "index.m3u8")
	}

	return res, nil
}
