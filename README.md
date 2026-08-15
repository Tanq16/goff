<div align="center">
  <img src=".github/assets/logo.svg" alt="goff Logo" width="200">
  <h1>goff</h1>

  <a href="https://github.com/Tanq16/goff/actions/workflows/release.yaml"><img alt="Build Workflow" src="https://github.com/Tanq16/goff/actions/workflows/release.yaml/badge.svg"></a>&nbsp;<a href="https://github.com/Tanq16/goff/releases"><img alt="GitHub Release" src="https://img.shields.io/github/v/release/Tanq16/goff"></a><br><br>
  <a href="#capabilities">Capabilities</a> &bull; <a href="#installation">Installation</a> &bull; <a href="#usage">Usage</a> &bull; <a href="#presets">Presets</a> &bull; <a href="#tips-and-notes">Tips & Notes</a>
</div>

---

**goff** (Go FFmpeg) is a lightweight, keyboard-driven Terminal User Interface (TUI) and CLI harness for FFmpeg. It eliminates the cognitive friction of remembering complex FFmpeg flags by providing intelligent media probing, curated presets, interactive wizard workflows, and headless AI-scriptable automation.

## Capabilities

| Category | Features | Description |
|----------|----------|-------------|
| **Video Compression** | `web-optimize`, `discord-25mb`, `discord-10mb` | High-efficiency H.265 / AV1 / H.264 compression, max 1080p SDR downscale, auto Hable HDR tone-mapping |
| **Stream Operations** | `remux`, `hls-fmp4`, `hls-ts`, `extract`, `trim`, `gif` | Instant lossless container switching, VoD HLS fMP4 / MPEG-TS packaging, audio extraction, 2-pass palettegen GIF/WebP |
| **Transformations** | `scale`, `rotate-90`, `rotate-180`, `rotate-270`, `vertical-9-16`, `speed`, `mute` | Multi-resolution scaling, 90°/180°/270° rotation & horizontal/vertical flips, 9:16 vertical crop, pitch-corrected speed |
| **Audio Suite** | `podcast-master`, `convert`, `normalize` | Broadcast standard EBU R128 loudness normalization (-16 LUFS), bitrate transcoding, downmixing |
| **Multi-File & Batch** | `watermark`, `concat`, `mux`, `batch` | Watermark / logo overlays with corner/center alignment, multi-clip concatenation, external audio/subtitle muxing, parallel batch pool |
| **Metadata Inspection** | `inspect` | Detailed stream inspector displaying codecs, profiles, resolutions, FPS, HDR transfer characteristics |

## Installation

### Pre-built Binary

Download the compiled binary for your platform from [Releases](https://github.com/Tanq16/goff/releases):

```bash
# Linux / macOS (Apple Silicon & Intel)
ARCH=$(uname -m); [ "$ARCH" = "x86_64" ] && ARCH=amd64; [ "$ARCH" = "aarch64" ] && [ "$ARCH" = "arm64" ] || ARCH=arm64
curl -sL https://github.com/Tanq16/goff/releases/latest/download/goff-$(uname -s | tr '[:upper:]' '[:lower:]')-$ARCH -o goff
chmod +x goff
sudo mv goff /usr/local/bin/
```

### Build from Source

Requires Go 1.26+:

```bash
git clone https://github.com/Tanq16/goff.git
cd goff
make build
```

*Note: Ensure `ffmpeg` and `ffprobe` are installed on your system PATH.*

## Usage

### 1. Interactive TUI Mode (Default)

Simply run `goff` to browse the current directory or pass a target media file to enter the interactive wizard:

```bash
# Launch interactive file picker in current directory
goff

# Launch interactive action menu for a specific video/audio file
goff video.mp4

# Launch multi-file operations (concat, mux, batch)
goff clip1.mp4 clip2.mp4 clip3.mp4
```

### 2. Direct Preset Execution

Bypass the interactive menu for instant headless execution:

```bash
# Optimize video with standard web preset
goff input.mkv --preset web-optimize

# Compress video to strictly fit Discord 25MB upload limit
goff input.mp4 --preset discord-25mb

# Extract 320kbps MP3 audio from a video
goff video.mp4 --preset extract-mp3-320

# Convert video to high-quality animated GIF
goff recording.mp4 --preset animated-gif

# Crop widescreen video to 9:16 vertical for Shorts / Reels
goff clip.mp4 --preset vertical-9-16
```

### 3. Stream & Metadata Inspection

Inspect video, audio, and subtitle streams with formatted terminal tables:

```bash
goff inspect movie.mkv
```

### 4. Concurrent Batch Processing

Process multiple media files in parallel with configurable worker concurrency:

```bash
# Compress all MP4 files with 4 parallel workers
goff batch *.mp4 --preset web-optimize -j 4

# Extract MP3 audio from all video files
goff batch *.mkv --preset extract-mp3-320 -j 2
```

### 5. AI-Agent Scriptable Mode (`--for-ai`)

Run headlessly with deterministic plain-text output prefixes (`[OK]`, `[ERROR]`, `[PROGRESS]`, `[INFO]`) and piped stdin:

```bash
echo "video.mp4" | goff --for-ai
goff inspect video.mp4 --for-ai
```

## Presets

| Preset ID | Category | Description |
|-----------|----------|-------------|
| `web-optimize` | Video | High-efficiency H.265 / AV1 compression, 1080p max SDR, faststart web atom |
| `discord-25mb` | Video | Dynamic bitrate calculation strictly fitting video under 25MB |
| `discord-10mb` | Video | Dynamic bitrate calculation strictly fitting video under 10MB |
| `fast-remux` | Video | Lossless zero re-encode container change into `.mp4` (+faststart) |
| `animated-gif` | Video | 480p 15fps animated loop using 2-pass palettegen filter |
| `extract-mp3-320`| Video | 320 kbps constant bitrate MP3 audio extraction |
| `strip-audio` | Video | Removes all audio tracks (mute) |
| `vertical-9-16` | Video | Center-crop 16:9 widescreen video to 9:16 vertical |
| `podcast-master` | Audio | Broadcast standard EBU R128 loudness normalization (-16 LUFS) with 192k AAC/MP3 |
| `normalize-audio`| Audio | General EBU R128 broadcast normalization |
| `hls-fmp4`       | Video | VoD HLS streaming playlist with fMP4 segments & init.mp4 |
| `hls-ts`         | Video | Classic VoD HLS playlist with MPEG-TS segments |
| `rotate-90`      | Video | Rotates video 90 degrees clockwise |
| `rotate-180`     | Video | Rotates video 180 degrees (upside down) |
| `rotate-270`     | Video | Rotates video 90 degrees counter-clockwise |

## Tips and Notes

- **Safe Output Naming**: `goff` never overwrites source files by default; output files are named `<basename>.<operation>.<ext>` (and incremented to `.1.<ext>` if a collision is detected). Use `-y` / `--yes` to allow overwrite.
- **HDR Tone-Mapping**: 10-bit HDR10 and HLG videos are automatically tone-mapped to 8-bit SDR using the Hable curve during web optimization.
- **Audio Preservation**: Multi-channel audio (5.1/7.1) can be downmixed to stereo while preserving speech dialog clarity.
