package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/Tanq16/goff/internal/probe"
)

type ActionItem struct {
	ID          string
	Title       string
	Description string
}

type MenuModel struct {
	items       []ActionItem
	cursor      int
	probeResult *probe.ProbeResult
	fileCount   int
}

func NewMenuModel(p *probe.ProbeResult, fileCount int) MenuModel {
	var items []ActionItem

	if fileCount > 1 {
		items = []ActionItem{
			{ID: "multi_concat", Title: "Concatenate / Merge Clips", Description: "Stitch multiple video or audio clips together"},
			{ID: "multi_mux_audio", Title: "Mux External Audio into Video", Description: "Replace or add audio track to video"},
			{ID: "multi_mux_subs", Title: "Embed Subtitles into Video", Description: "Embed soft or burn hard subtitles"},
			{ID: "multi_watermark", Title: "Watermark / Logo Overlay", Description: "Overlay image/logo onto video with position & scale"},
			{ID: "multi_batch", Title: "Batch Process", Description: "Apply one operation across all selected files"},
		}
	} else if p != nil && p.IsVideo() {
		items = []ActionItem{
			{ID: "video_opt", Title: "Optimize & Compress", Description: "High-efficiency H.265/AV1 compression with SDR tone-mapping"},
			{ID: "video_remux", Title: "Fast Remux", Description: "Lossless container switch (MP4/MKV/MOV) with +faststart"},
			{ID: "video_hls", Title: "HLS VoD Streaming Package", Description: "Generate fMP4 / MPEG-TS HTTP Live Streaming playlist & segments"},
			{ID: "video_extract", Title: "Extract Audio", Description: "Extract audio stream to MP3, AAC, FLAC, or Opus"},
			{ID: "video_trim", Title: "Trim / Cut Segment", Description: "Extract specific time range by timestamps"},
			{ID: "video_gif", Title: "Animated GIF / WebP", Description: "High-quality 2-pass palettegen animated loop"},
			{ID: "video_transform", Title: "Transform (Scale / Rotate / 9:16 / Speed)", Description: "Scale resolution, rotate/flip, crop vertical for Shorts, change speed"},
			{ID: "inspect", Title: "Inspect Streams & Metadata", Description: "View codecs, resolutions, bitrates, HDR parameters"},
		}
	} else {
		items = []ActionItem{
			{ID: "audio_convert", Title: "Transcode & Convert", Description: "Convert between MP3, AAC, Opus, FLAC, WAV, OGG"},
			{ID: "audio_loudnorm", Title: "EBU R128 Normalization", Description: "Standard loudness normalization for podcasts & music"},
			{ID: "audio_trim", Title: "Trim Audio", Description: "Extract specific time range from audio file"},
			{ID: "inspect", Title: "Inspect Audio Metadata", Description: "View audio stream sample rates, channels, bitrate"},
		}
	}

	return MenuModel{
		items:       items,
		cursor:      0,
		probeResult: p,
		fileCount:   fileCount,
	}
}

func (m MenuModel) Init() tea.Cmd {
	return nil
}

func (m MenuModel) Update(msg tea.Msg) (MenuModel, tea.Cmd, *ActionItem) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case "enter":
			if len(m.items) > 0 {
				item := m.items[m.cursor]
				return m, nil, &item
			}
		case "ctrl+c", "q":
			return m, tea.Quit, nil
		}
	}
	return m, nil, nil
}

func (m MenuModel) View() string {
	boxWidth := defaultBoxWidth
	var lines []string

	title := "Media Suite"
	if m.probeResult != nil {
		title = fmt.Sprintf("Media Suite: %s", m.probeResult.BaseFileName())
	} else {
		title = fmt.Sprintf("Media Suite (%d files)", m.fileCount)
	}

	lines = append(lines, renderBoxTop(title, boxWidth))
	lines = append(lines, renderBoxEmpty(boxWidth))

	if m.probeResult != nil {
		p := m.probeResult
		mediaType := "Audio"
		if p.IsVideo() {
			mediaType = fmt.Sprintf("Video (%s, %s, %s)", p.Resolution(), p.HDRType(), p.HumanDuration())
		}
		lines = append(lines, padBoxLine("  "+activeBulletStyle.Render("● ")+activeItemStyle.Render(p.BaseFileName())+"  "+footerStyle.Render(mediaType), boxWidth))
		lines = append(lines, padBoxLine("    "+branchStyle.Render("└─ ")+footerStyle.Render(fmt.Sprintf("Size: %s  •  Format: %s", p.HumanSize(), p.Format.FormatName)), boxWidth))
	} else {
		lines = append(lines, padBoxLine("  "+activeBulletStyle.Render("● ")+activeItemStyle.Render(fmt.Sprintf("Multi-File Mode (%d files selected)", m.fileCount)), boxWidth))
	}

	lines = append(lines, renderBoxEmpty(boxWidth))
	lines = append(lines, renderBoxDivider(boxWidth))
	lines = append(lines, renderBoxEmpty(boxWidth))

	for i, item := range m.items {
		if i == m.cursor {
			lines = append(lines, padBoxLine("  "+activeBulletStyle.Render("● ")+activeItemStyle.Render(item.Title), boxWidth))
			lines = append(lines, padBoxLine("    "+branchStyle.Render("└─ ")+descStyle.Render(item.Description), boxWidth))
		} else {
			lines = append(lines, padBoxLine("  "+normalBulletStyle.Render("○ ")+normalItemStyle.Render(item.Title), boxWidth))
		}
	}

	lines = append(lines, renderBoxEmpty(boxWidth))
	lines = append(lines, renderBoxDivider(boxWidth))
	lines = append(lines, padBoxLine("  "+footerStyle.Render("↑/↓ or j/k to navigate  •  Enter to select  •  Esc to go back  •  q to quit"), boxWidth))
	lines = append(lines, renderBoxBottom(boxWidth))

	return strings.Join(lines, "\n")
}
