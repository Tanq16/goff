package presets

import (
	"github.com/Tanq16/goff/internal/ops"
	"github.com/Tanq16/goff/internal/probe"
)

func init() {
	Register(Preset{
		ID:          "web-optimize",
		Name:        "Web Optimize (H.265 / SDR)",
		Description: "High compression 1080p SDR with auto tone-mapping & faststart",
		Category:    ops.CategoryVideo,
		Build: func(input string, p *probe.ProbeResult) (*ops.OpResult, error) {
			return ops.BuildVideoOptimize(input, p, ops.VideoOptimizeOpts{
				Codec:     "hevc",
				CRF:       30,
				MaxHeight: 1080,
			})
		},
	})

	Register(Preset{
		ID:          "discord-25mb",
		Name:        "Discord 25MB Fit",
		Description: "Calculates optimal bitrate to guarantee output is strictly under 25MB",
		Category:    ops.CategoryVideo,
		Build: func(input string, p *probe.ProbeResult) (*ops.OpResult, error) {
			return ops.BuildVideoOptimize(input, p, ops.VideoOptimizeOpts{
				Codec:        "hevc",
				TargetSizeMB: 24.5,
				CustomSuffix: "discord-25mb",
			})
		},
	})

	Register(Preset{
		ID:          "discord-10mb",
		Name:        "Discord 10MB Fit",
		Description: "Calculates optimal bitrate to guarantee output is strictly under 10MB",
		Category:    ops.CategoryVideo,
		Build: func(input string, p *probe.ProbeResult) (*ops.OpResult, error) {
			return ops.BuildVideoOptimize(input, p, ops.VideoOptimizeOpts{
				Codec:        "hevc",
				TargetSizeMB: 9.5,
				CustomSuffix: "discord-10mb",
			})
		},
	})

	Register(Preset{
		ID:          "fast-remux",
		Name:        "Fast Remux to MP4",
		Description: "Lossless container switch with zero re-encoding and +faststart",
		Category:    ops.CategoryVideo,
		Build: func(input string, p *probe.ProbeResult) (*ops.OpResult, error) {
			return ops.BuildVideoRemux(input, p, ops.VideoRemuxOpts{
				TargetExt: "mp4",
			})
		},
	})

	Register(Preset{
		ID:          "animated-gif",
		Name:        "Animated GIF (Palettegen)",
		Description: "High-quality 480p 15fps animated loop using 2-pass palette generation",
		Category:    ops.CategoryVideo,
		Build: func(input string, p *probe.ProbeResult) (*ops.OpResult, error) {
			return ops.BuildVideoGIF(input, p, ops.VideoGIFOpts{
				Width:  480,
				FPS:    15,
				Format: "gif",
			})
		},
	})

	Register(Preset{
		ID:          "extract-mp3-320",
		Name:        "Extract MP3 320k",
		Description: "Pulls the audio stream from a video and encodes to 320 kbps MP3",
		Category:    ops.CategoryVideo,
		Build: func(input string, p *probe.ProbeResult) (*ops.OpResult, error) {
			return ops.BuildVideoExtract(input, p, ops.VideoExtractOpts{
				Format:  "mp3",
				Bitrate: "320k",
			})
		},
	})

	Register(Preset{
		ID:          "strip-audio",
		Name:        "Strip Audio (Mute)",
		Description: "Removes all audio tracks from the video",
		Category:    ops.CategoryVideo,
		Build: func(input string, p *probe.ProbeResult) (*ops.OpResult, error) {
			return ops.BuildVideoTransform(input, p, ops.VideoTransformOpts{
				StripAudio: true,
			})
		},
	})

	Register(Preset{
		ID:          "vertical-9-16",
		Name:        "Crop 9:16 Vertical",
		Description: "Center-crops video to 9:16 vertical aspect ratio for Shorts / Reels",
		Category:    ops.CategoryVideo,
		Build: func(input string, p *probe.ProbeResult) (*ops.OpResult, error) {
			return ops.BuildVideoTransform(input, p, ops.VideoTransformOpts{
				CropVertical: true,
			})
		},
	})

	Register(Preset{
		ID:          "podcast-master",
		Name:        "Podcast Master (EBU R128)",
		Description: "Broadcast-standard loudness normalization (-16 LUFS) with 192k AAC/MP3",
		Category:    ops.CategoryAudio,
		Build: func(input string, p *probe.ProbeResult) (*ops.OpResult, error) {
			return ops.BuildAudioLoudnorm(input, p, ops.AudioLoudnormOpts{
				IntegratedLoudness: -16.0,
				TruePeak:           -1.5,
				LoudnessRange:      11.0,
				OutputExt:          "mp3",
			})
		},
	})

	Register(Preset{
		ID:          "normalize-audio",
		Name:        "Normalize Loudness",
		Description: "EBU R128 broadcast normalization for video or audio files",
		Category:    ops.CategoryAudio,
		Build: func(input string, p *probe.ProbeResult) (*ops.OpResult, error) {
			return ops.BuildAudioLoudnorm(input, p, ops.AudioLoudnormOpts{
				IntegratedLoudness: -16.0,
			})
		},
	})

	Register(Preset{
		ID:          "hls-fmp4",
		Name:        "HLS VoD (fMP4 Streamable)",
		Description: "VoD HLS streaming playlist with fMP4 segments & init.mp4",
		Category:    ops.CategoryVideo,
		Build: func(input string, p *probe.ProbeResult) (*ops.OpResult, error) {
			return ops.BuildVideoHLS(input, p, ops.VideoHLSOpts{
				Format: ops.HLSFormatFMP4,
			})
		},
	})

	Register(Preset{
		ID:          "hls-ts",
		Name:        "HLS VoD (MPEG-TS)",
		Description: "Classic VoD HLS playlist with MPEG-TS segments for legacy players",
		Category:    ops.CategoryVideo,
		Build: func(input string, p *probe.ProbeResult) (*ops.OpResult, error) {
			return ops.BuildVideoHLS(input, p, ops.VideoHLSOpts{
				Format: ops.HLSFormatMPEGTS,
			})
		},
	})

	Register(Preset{
		ID:          "rotate-90",
		Name:        "Rotate 90° Clockwise",
		Description: "Rotates video 90 degrees clockwise",
		Category:    ops.CategoryVideo,
		Build: func(input string, p *probe.ProbeResult) (*ops.OpResult, error) {
			return ops.BuildVideoTransform(input, p, ops.VideoTransformOpts{
				Rotate:       "90_cw",
				CustomSuffix: "rot90",
			})
		},
	})

	Register(Preset{
		ID:          "rotate-180",
		Name:        "Rotate 180°",
		Description: "Rotates video 180 degrees (upside down)",
		Category:    ops.CategoryVideo,
		Build: func(input string, p *probe.ProbeResult) (*ops.OpResult, error) {
			return ops.BuildVideoTransform(input, p, ops.VideoTransformOpts{
				Rotate:       "180",
				CustomSuffix: "rot180",
			})
		},
	})

	Register(Preset{
		ID:          "rotate-270",
		Name:        "Rotate 90° Counter-Clockwise",
		Description: "Rotates video 90 degrees counter-clockwise (270° CW)",
		Category:    ops.CategoryVideo,
		Build: func(input string, p *probe.ProbeResult) (*ops.OpResult, error) {
			return ops.BuildVideoTransform(input, p, ops.VideoTransformOpts{
				Rotate:       "90_ccw",
				CustomSuffix: "rot270",
			})
		},
	})
}

