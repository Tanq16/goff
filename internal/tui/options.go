package tui

import (
	"path/filepath"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/Tanq16/goff/internal/ops"
	"github.com/Tanq16/goff/internal/presets"
	"github.com/Tanq16/goff/internal/probe"
)

type OptionChoice struct {
	ID          string
	Title       string
	Description string
	Build       func(input string, p *probe.ProbeResult) (*ops.OpResult, error)
}

type OptionsModel struct {
	actionID    string
	title       string
	choices     []OptionChoice
	cursor      int
	probeResult *probe.ProbeResult
	inputPath   string
}

func NewOptionsModel(actionID string, inputPath string, p *probe.ProbeResult) OptionsModel {
	var choices []OptionChoice
	title := "Configure Options"

	switch actionID {
	case "video_opt":
		title = "Video Optimization & Compression"
		choices = []OptionChoice{
			{
				ID: "hevc_balanced", Title: "H.265 (HEVC) — Balanced (CRF 30)", Description: "Recommended: high compression, 1080p SDR with HDR tone-map",
				Build: func(in string, pr *probe.ProbeResult) (*ops.OpResult, error) {
					return ops.BuildVideoOptimize(in, pr, ops.VideoOptimizeOpts{Codec: "hevc", CRF: 30})
				},
			},
			{
				ID: "av1_high", Title: "AV1 (SVT-AV1) — Next-Gen (CRF 32)", Description: "Superior compression for modern players",
				Build: func(in string, pr *probe.ProbeResult) (*ops.OpResult, error) {
					return ops.BuildVideoOptimize(in, pr, ops.VideoOptimizeOpts{Codec: "av1", CRF: 32})
				},
			},
			{
				ID: "h264_compat", Title: "H.264 — Maximum Compatibility (CRF 23)", Description: "Universal playback on older devices & browsers",
				Build: func(in string, pr *probe.ProbeResult) (*ops.OpResult, error) {
					return ops.BuildVideoOptimize(in, pr, ops.VideoOptimizeOpts{Codec: "h264", CRF: 23})
				},
			},
			{
				ID: "discord_25mb", Title: "Discord 25MB Target Size Fit", Description: "Calculates optimal bitrate to stay strictly under 25MB",
				Build: func(in string, pr *probe.ProbeResult) (*ops.OpResult, error) {
					return ops.BuildVideoOptimize(in, pr, ops.VideoOptimizeOpts{Codec: "hevc", TargetSizeMB: 24.5, CustomSuffix: "discord-25mb"})
				},
			},
		}

	case "video_remux":
		title = "Fast Remux (Lossless Container Switch)"
		choices = []OptionChoice{
			{
				ID: "remux_mp4", Title: "Remux to MP4 (+faststart)", Description: "Lossless stream copy to universal MP4 container",
				Build: func(in string, pr *probe.ProbeResult) (*ops.OpResult, error) {
					return ops.BuildVideoRemux(in, pr, ops.VideoRemuxOpts{TargetExt: "mp4"})
				},
			},
			{
				ID: "remux_mkv", Title: "Remux to MKV", Description: "Lossless stream copy to Matroska container",
				Build: func(in string, pr *probe.ProbeResult) (*ops.OpResult, error) {
					return ops.BuildVideoRemux(in, pr, ops.VideoRemuxOpts{TargetExt: "mkv"})
				},
			},
			{
				ID: "remux_mov", Title: "Remux to QuickTime MOV", Description: "Lossless stream copy to Apple MOV container",
				Build: func(in string, pr *probe.ProbeResult) (*ops.OpResult, error) {
					return ops.BuildVideoRemux(in, pr, ops.VideoRemuxOpts{TargetExt: "mov"})
				},
			},
		}

	case "video_extract":
		title = "Extract Audio from Video"
		choices = []OptionChoice{
			{
				ID: "extract_mp3_320", Title: "MP3 — 320 kbps (High Quality)", Description: "Standard high-bitrate MP3 audio file",
				Build: func(in string, pr *probe.ProbeResult) (*ops.OpResult, error) {
					return ops.BuildVideoExtract(in, pr, ops.VideoExtractOpts{Format: "mp3", Bitrate: "320k"})
				},
			},
			{
				ID: "extract_aac_256", Title: "AAC / M4A — 256 kbps (Apple Standard)", Description: "High efficiency AAC stream in M4A container",
				Build: func(in string, pr *probe.ProbeResult) (*ops.OpResult, error) {
					return ops.BuildVideoExtract(in, pr, ops.VideoExtractOpts{Format: "aac", Bitrate: "256k"})
				},
			},
			{
				ID: "extract_flac", Title: "FLAC — Lossless", Description: "Lossless uncompressed audio archive",
				Build: func(in string, pr *probe.ProbeResult) (*ops.OpResult, error) {
					return ops.BuildVideoExtract(in, pr, ops.VideoExtractOpts{Format: "flac"})
				},
			},
			{
				ID: "extract_opus", Title: "Opus — 160 kbps", Description: "Best fidelity-to-size ratio for speech and music",
				Build: func(in string, pr *probe.ProbeResult) (*ops.OpResult, error) {
					return ops.BuildVideoExtract(in, pr, ops.VideoExtractOpts{Format: "opus", Bitrate: "160k"})
				},
			},
		}

	case "video_gif":
		title = "Create Animated GIF / WebP"
		choices = []OptionChoice{
			{
				ID: "gif_480p", Title: "GIF — 480p 15fps (Standard)", Description: "2-pass palettegen filter for high color fidelity",
				Build: func(in string, pr *probe.ProbeResult) (*ops.OpResult, error) {
					return ops.BuildVideoGIF(in, pr, ops.VideoGIFOpts{Width: 480, FPS: 15, Format: "gif"})
				},
			},
			{
				ID: "gif_720p", Title: "GIF — 720p 24fps (High Res)", Description: "Crisp high-resolution animated GIF",
				Build: func(in string, pr *probe.ProbeResult) (*ops.OpResult, error) {
					return ops.BuildVideoGIF(in, pr, ops.VideoGIFOpts{Width: 720, FPS: 24, Format: "gif"})
				},
			},
			{
				ID: "webp_480p", Title: "Animated WebP — 480p (Lightweight)", Description: "Smaller size than GIF with alpha support",
				Build: func(in string, pr *probe.ProbeResult) (*ops.OpResult, error) {
					return ops.BuildVideoGIF(in, pr, ops.VideoGIFOpts{Width: 480, FPS: 15, Format: "webp"})
				},
			},
		}

	case "video_transform":
		title = "Video Transformations"
		choices = []OptionChoice{
			{
				ID: "crop_vertical", Title: "Crop 9:16 Vertical (Shorts / Reels / TikTok)", Description: "Center-crop 16:9 widescreen to 9:16 vertical video",
				Build: func(in string, pr *probe.ProbeResult) (*ops.OpResult, error) {
					return ops.BuildVideoTransform(in, pr, ops.VideoTransformOpts{CropVertical: true})
				},
			},
			{
				ID: "scale_1080p", Title: "Scale to 1080p FHD", Description: "Downscale to max 1920x1080 preserving aspect ratio",
				Build: func(in string, pr *probe.ProbeResult) (*ops.OpResult, error) {
					return ops.BuildVideoTransform(in, pr, ops.VideoTransformOpts{Scale: "1080p"})
				},
			},
			{
				ID: "scale_720p", Title: "Scale to 720p HD", Description: "Downscale to max 1280x720 preserving aspect ratio",
				Build: func(in string, pr *probe.ProbeResult) (*ops.OpResult, error) {
					return ops.BuildVideoTransform(in, pr, ops.VideoTransformOpts{Scale: "720p"})
				},
			},
			{
				ID: "speed_15x", Title: "Speed up 1.5x (Pitch Corrected)", Description: "1.5x playback speed with audio atempo pitch correction",
				Build: func(in string, pr *probe.ProbeResult) (*ops.OpResult, error) {
					return ops.BuildVideoTransform(in, pr, ops.VideoTransformOpts{Speed: 1.5})
				},
			},
			{
				ID: "speed_20x", Title: "Speed up 2.0x (Pitch Corrected)", Description: "2.0x playback speed with audio atempo pitch correction",
				Build: func(in string, pr *probe.ProbeResult) (*ops.OpResult, error) {
					return ops.BuildVideoTransform(in, pr, ops.VideoTransformOpts{Speed: 2.0})
				},
			},
			{
				ID: "strip_audio", Title: "Strip Audio (Mute Video)", Description: "Remove all audio streams from container",
				Build: func(in string, pr *probe.ProbeResult) (*ops.OpResult, error) {
					return ops.BuildVideoTransform(in, pr, ops.VideoTransformOpts{StripAudio: true})
				},
			},
		}

	case "audio_convert":
		title = "Audio Transcoding"
		choices = []OptionChoice{
			{
				ID: "conv_mp3", Title: "Convert to MP3 (320 kbps)", Description: "Universal MP3 constant high bitrate",
				Build: func(in string, pr *probe.ProbeResult) (*ops.OpResult, error) {
					return ops.BuildAudioConvert(in, pr, ops.AudioConvertOpts{Format: "mp3", Bitrate: "320k"})
				},
			},
			{
				ID: "conv_aac", Title: "Convert to AAC / M4A (256 kbps)", Description: "High-efficiency AAC stream",
				Build: func(in string, pr *probe.ProbeResult) (*ops.OpResult, error) {
					return ops.BuildAudioConvert(in, pr, ops.AudioConvertOpts{Format: "aac", Bitrate: "256k"})
				},
			},
			{
				ID: "conv_flac", Title: "Convert to FLAC (Lossless)", Description: "Lossless audio compression",
				Build: func(in string, pr *probe.ProbeResult) (*ops.OpResult, error) {
					return ops.BuildAudioConvert(in, pr, ops.AudioConvertOpts{Format: "flac"})
				},
			},
			{
				ID: "conv_opus", Title: "Convert to Opus (160 kbps)", Description: "Next-gen audio codec for high fidelity",
				Build: func(in string, pr *probe.ProbeResult) (*ops.OpResult, error) {
					return ops.BuildAudioConvert(in, pr, ops.AudioConvertOpts{Format: "opus", Bitrate: "160k"})
				},
			},
		}

	case "audio_loudnorm":
		title = "EBU R128 Loudness Normalization"
		choices = []OptionChoice{
			{
				ID: "loud_podcast", Title: "Podcast Standard (-16 LUFS)", Description: "Ideal target loudness for podcasts and speech",
				Build: func(in string, pr *probe.ProbeResult) (*ops.OpResult, error) {
					return ops.BuildAudioLoudnorm(in, pr, ops.AudioLoudnormOpts{IntegratedLoudness: -16.0, TruePeak: -1.5, LoudnessRange: 11.0})
				},
			},
			{
				ID: "loud_streaming", Title: "Streaming Standard (-14 LUFS)", Description: "Target loudness for Spotify / YouTube Music",
				Build: func(in string, pr *probe.ProbeResult) (*ops.OpResult, error) {
					return ops.BuildAudioLoudnorm(in, pr, ops.AudioLoudnormOpts{IntegratedLoudness: -14.0, TruePeak: -1.0, LoudnessRange: 10.0})
				},
			},
		}

	case "presets_menu":
		title = "Select Preset"
		available := presets.MatchPresetForMedia(p)
		for _, pr := range available {
			presetCopy := pr
			choices = append(choices, OptionChoice{
				ID:          presetCopy.ID,
				Title:       presetCopy.Name,
				Description: presetCopy.Description,
				Build: func(in string, prb *probe.ProbeResult) (*ops.OpResult, error) {
					return presetCopy.Build(in, prb)
				},
			})
		}
	}

	return OptionsModel{
		actionID:    actionID,
		title:       title,
		choices:     choices,
		cursor:      0,
		probeResult: p,
		inputPath:   inputPath,
	}
}

func (m OptionsModel) Init() tea.Cmd {
	return nil
}

func (m OptionsModel) Update(msg tea.Msg) (OptionsModel, tea.Cmd, *OptionChoice) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "enter":
			if len(m.choices) > 0 {
				choice := m.choices[m.cursor]
				return m, nil, &choice
			}
		case "ctrl+c", "q":
			return m, tea.Quit, nil
		}
	}
	return m, nil, nil
}

func (m OptionsModel) View() string {
	boxWidth := defaultBoxWidth
	var lines []string

	lines = append(lines, renderBoxTop("Configure: "+m.title, boxWidth))
	lines = append(lines, renderBoxEmpty(boxWidth))

	if m.inputPath != "" {
		lines = append(lines, padBoxLine("  "+labelStyle.Render("Target Media: ")+valueStyle.Render(filepath.Base(m.inputPath)), boxWidth))
		lines = append(lines, renderBoxEmpty(boxWidth))
		lines = append(lines, renderBoxDivider(boxWidth))
		lines = append(lines, renderBoxEmpty(boxWidth))
	}

	for i, c := range m.choices {
		if i == m.cursor {
			lines = append(lines, padBoxLine("  "+activeBulletStyle.Render("● ")+activeItemStyle.Render(c.Title), boxWidth))
			lines = append(lines, padBoxLine("    "+branchStyle.Render("└─ ")+descStyle.Render(c.Description), boxWidth))
		} else {
			lines = append(lines, padBoxLine("  "+normalBulletStyle.Render("○ ")+normalItemStyle.Render(c.Title), boxWidth))
		}
	}

	lines = append(lines, renderBoxEmpty(boxWidth))
	lines = append(lines, renderBoxDivider(boxWidth))
	lines = append(lines, padBoxLine("  "+footerStyle.Render("↑/↓ or j/k to navigate  •  Enter to start  •  Esc to go back  •  q to quit"), boxWidth))
	lines = append(lines, renderBoxBottom(boxWidth))

	return strings.Join(lines, "\n")
}
