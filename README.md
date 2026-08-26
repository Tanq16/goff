<div align="center">
  <img src=".github/assets/logo.svg" alt="goff Logo" width="200">
  <h1>goff</h1>

  <a href="https://github.com/Tanq16/goff/actions/workflows/release.yaml"><img alt="Build Workflow" src="https://github.com/Tanq16/goff/actions/workflows/release.yaml/badge.svg"></a>&nbsp;<a href="https://github.com/Tanq16/goff/releases"><img alt="GitHub Release" src="https://img.shields.io/github/v/release/Tanq16/goff"></a><br><br>
  <a href="#capabilities">Capabilities</a> &bull; <a href="#installation">Installation</a> &bull; <a href="#usage">Usage</a> &bull; <a href="#tips-and-notes">Tips & Notes</a>
</div>

---

**goff** (Go FFmpeg) is a terminal media suite for FFmpeg: one verb per operation, sensible defaults, and no filter syntax to look up.

It exists so you stop looking up filter syntax for the same dozen jobs. It is not a video editor or a replacement for FFmpeg itself.

## Capabilities

| Category | Commands | Description |
|----------|----------|-------------|
| **Compression** | `compress` | H.265 / AV1 / H.264 re-encoding, height caps, target file size, lossless mode, automatic HDR tone-mapping, playback-compatibility normalization |
| **Containers** | `remux`, `hls` | Lossless container switching, VoD HLS packaging as fMP4 or MPEG-TS |
| **Audio** | `extract`, `convert`, `normalize`, `mix` | Audio extraction from video, format transcoding, EBU R128 loudness normalization, layering tracks into one |
| **Transforms** | `rotate`, `scale`, `speed`, `crop`, `mute` | Rotation and flips, resolution tiers, pitch-corrected speed, 9:16 vertical crop, audio removal |
| **Segments** | `trim`, `gif` | Time-range cuts, 2-pass palettegen GIF and animated WebP |
| **Multi-file** | `concat`, `mux`, `subs`, `watermark` | Clip joining, external audio muxing, subtitle embedding, logo overlays |
| **Inspection** | `inspect` | Stream table with codecs, resolutions, per-stream and container bitrates, and HDR transfer characteristics |

## Installation

`ffmpeg` and `ffprobe` must be on your `PATH`, whichever path below you take.

### Pre-built Binary

Download the compiled binary for your platform from [Releases](https://github.com/Tanq16/goff/releases):

```bash
# Linux / macOS (Apple Silicon & Intel)
ARCH=$(uname -m); case "$ARCH" in x86_64) ARCH=amd64 ;; aarch64|arm64) ARCH=arm64 ;; esac
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

## Usage

Every command takes one or more input files and writes alongside them, so nothing is overwritten by default. These flags work on all of them:

| Flag | Effect |
|------|--------|
| `-o`, `--output` | Explicit output path, single input only |
| `-y`, `--yes` | Allow overwriting existing files |
| `-j`, `--jobs` | Concurrent encodes when several inputs are given (default 2) |
| `--for-ai` | Plain-text prefixed output for scripts and agents |
| `--debug` | Structured logs, including the underlying FFmpeg error |

Pass several files to any per-file command and they process in parallel:

```bash
goff compress *.mkv -j 4
```

`goff --help` lists every command grouped by what it does, and `goff <command> --help` carries an example of the invocation you want. A run where any file failed exits non-zero.

### Compress

```bash
goff compress input.mkv                       # H.265, 1080p cap, faststart
goff compress input.mkv --codec av1 --crf 28
goff compress input.mp4 --size 25MB           # bitrate solved to land under 25MB
goff compress input.mkv --height 720
goff compress master.mov --lossless           # no video quality loss, source resolution kept
goff compress input.mkv --compat              # first tracks, 8-bit, CFR, 48kHz stereo
goff compress input.mkv --subs all            # carry every subtitle track into the MP4
goff compress input.mkv --preset slow --audio-bitrate 192k
```

### Containers and streaming

```bash
goff remux input.mkv --to mp4     # no re-encoding
goff remux capture.mkv --to mp4 --fix-timestamps
goff hls input.mp4 --to fmp4      # writes input.hls-fmp4/index.m3u8 plus segments
goff hls input.mp4 --to ts --segment 4
```

### Audio

```bash
goff extract video.mp4 --to mp3 --bitrate 320k
goff convert song.wav --to opus --rate 48000
goff normalize podcast.wav --to mp3          # -16 LUFS, -1.5 dBTP
goff normalize lecture.mp4 --lufs -14

goff mix voice.wav music.mp3                 # both play from 0
goff mix voice.wav music.mp3:at=5:vol=0.3    # music enters 5s in, at 30% level
goff mix a.mp3 b.mp3 --fit shortest          # stop at the shortest input
```

`mix` layers tracks so they play at the same time. To join clips end to end instead, use `concat`.

### Transforms

Each transform is its own verb, and `--by` carries the value that verb is named for. Sibling flags compose into a **single** encode instead of stacking generations of quality loss:

```bash
goff rotate clip.mp4 --by 90          # 90, 180, 270, hflip, vflip
goff scale clip.mp4 --by 720p         # 720p, 1080p, 4k, or 1280x720
goff speed clip.mp4 --by 1.5          # audio pitch corrected
goff crop clip.mp4                    # 9:16 for Shorts / Reels
goff mute clip.mp4

goff rotate clip.mp4 --by 90 --scale 720p --mute   # one encode, three changes
```

### Segments

```bash
goff trim video.mp4 --start 00:01:30 --end 00:02:00
goff trim video.mp4 --start 90 --duration 30 --accurate
goff gif screencast.mp4 --width 640 --fps 20 --start 5 --duration 8
goff gif screencast.mp4 --to webp
```

### Multi-file

```bash
goff concat part1.mp4 part2.mp4 part3.mp4
goff concat a.mp4 b.mp4 --reencode              # for mismatched codecs

goff mux talk.mp4 --audio music.mp3:vol=0.3     # music mixed under the original audio
goff mux talk.mp4 --audio dub.m4a --audio-mode replace
goff mux film.mkv --audio en.m4a --audio fr.m4a --audio-mode separate
goff mux clip.mp4 --audio bed.mp3 --fit longest # run to the end of the music

goff subs video.mp4 --file subs.srt             # embed as a track
goff subs video.mp4 --file subs.srt --burn      # render into the picture

goff watermark video.mp4 --logo logo.png --at bottom-right --width 12 --opacity 0.6
```

### Inspection

```bash
goff inspect movie.mkv
```

### Scripting and agents

`--for-ai` swaps styled output for parseable prefixes (`[OK]`, `[ERROR]`, `[PROGRESS]`, `[INFO]`), keeps every progress line instead of redrawing one, and renders tables as Markdown:

```bash
goff compress video.mp4 --for-ai
goff inspect video.mp4 --for-ai
```

## Tips and Notes

- **Safe output naming**: outputs are written as `<name>.<operation>.<ext>` next to the input, incrementing to `.1.<ext>` on a collision. Nothing is overwritten without `-y`.
- **Composed transforms**: combining transform flags produces one encode named after the verb you typed, while a single transform keeps its descriptive name, such as `clip.rot90.mp4`.
- **Input offsets and levels**: `mix` and `mux --audio` accept `<file>:at=<time>` to delay a track and `:vol=<factor>` to change its level, both optional and in either order. A track with neither starts at 0 at its own level.
- **Audio modes**: `mux --audio-mode` decides what happens to the audio a video already has. `mix` layers it with the new tracks into one, `replace` drops it, and `separate` keeps every track selectable.
- **HLS layout**: each packaged video gets its own directory holding `index.m3u8` and the segments, so two packaged videos never share a segment name. fMP4 packaging adds an `init.mp4` next to them.
- **Size targets**: `--size` holds back headroom below the number you give, so a 25MB budget targets 24.5MB and the muxed result stays under the limit.
- **HDR tone-mapping**: HDR10 and HLG sources are tone-mapped to 8-bit SDR with the Hable curve during `compress`, which takes priority over `--height` for those inputs.
- **Lossless compression**: `compress --lossless` keeps the source resolution and skips tone-mapping, since both are lossy, and it re-encodes audio to AAC like every other `compress` run. It cannot be combined with `--crf`, `--size`, or `--height`.
- **Playback compatibility**: `compress --compat` normalizes a run for players that reject anything unusual. It takes the first video and audio track instead of letting FFmpeg choose, forces 8-bit color and a constant frame rate, and downmixes audio to 48kHz stereo. It cannot be combined with `--lossless`.
- **Subtitles**: `compress --subs all` maps every subtitle track and converts it to `mov_text`, which covers text subtitles and fails on image ones such as PGS. `none` drops them, and the default `auto` leaves the choice to FFmpeg.
- **Timestamp shifting**: `remux --fix-timestamps` moves a negative start time to zero, which matters for captures whose audio and video begin at different points.
- **Container bitrate**: `inspect` prints the container bitrate on its summary line, deriving it from size over duration when FFprobe reports none. A derived figure can differ slightly from a reported one, and a file FFprobe gives no duration for shows `-`.
- **Out-of-range numbers**: a numeric flag given a value past its range is pulled to the nearest end of that range rather than rejected. `--help` prints the accepted range as the flag's type, such as `--crf 1..63`. H.265 caps at 51, so a higher `--crf` lands there when `--codec hevc` is in play.
- **Failure detail**: a failed encode reports the FFmpeg error only under `--debug`, which keeps a wall of filter-graph text out of normal runs.
