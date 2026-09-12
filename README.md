<div align="center">
  <img src=".github/assets/logo.svg" alt="goff Logo" width="200">
  <h1>goff</h1>

  <a href="https://github.com/Tanq16/goff/actions/workflows/release.yaml"><img alt="Build Workflow" src="https://github.com/Tanq16/goff/actions/workflows/release.yaml/badge.svg"></a>&nbsp;<a href="https://github.com/Tanq16/goff/releases"><img alt="GitHub Release" src="https://img.shields.io/github/v/release/Tanq16/goff"></a><br><br>
  <a href="#capabilities">Capabilities</a> &bull; <a href="#install">Install</a> &bull; <a href="#usage">Usage</a> &bull; <a href="#notes">Notes</a>
</div>

---

**goff** (Go FFmpeg) is a terminal media suite for FFmpeg: one verb per operation, sensible defaults, and no filter syntax to look up.

It exists so the same dozen jobs stop costing a trip through the FFmpeg manual. It is not a video editor or a replacement for FFmpeg itself.

## Capabilities

| Category | Commands | Description |
|----------|----------|-------------|
| **Video** | `compress`, `remux`, `hls`, `transform`, `watermark` | H.265 / AV1 / H.264 re-encoding normalized for playback, lossless container switching, HLS VoD packaging as fMP4 or MPEG-TS, rotation and scaling composed into one pass, logo overlays |
| **Audio** | `extract`, `convert`, `normalize`, `mix` | Audio and subtitle extraction from video, format transcoding, EBU R128 loudness normalization, layering tracks into one |
| **Segments** | `trim`, `gif`, `thumbnail` | Time-range cuts, 2-pass palettegen GIF and animated WebP, single-frame JPEG stills |
| **Combining** | `concat`, `mux`, `subs` | Clip joining, external audio muxing, subtitle embedding |
| **Info** | `inspect` | Stream table with codecs, resolutions, languages, per-stream and container bitrates, and HDR transfer characteristics, plus a packet-level browser playability check |

## Install

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

Requires Go 1.27+:

```bash
git clone https://github.com/Tanq16/goff.git
cd goff
make build
```

## Usage

Every per-file command writes one output per input next to the source, named `<name>.<operation>.<ext>` and numbered to `<name>.<operation>.1.<ext>` on a collision, so an existing file is never replaced.

| Flag | Effect | Where |
|------|--------|-------|
| `--debug` | Structured logs instead of styled output, carrying the FFmpeg error that a failed encode otherwise withholds | every command |
| `-o`, `--output` | Explicit output path, overwritten if it exists, single input only | every command except `inspect` |
| `-j`, `--jobs` | Concurrent encodes when several inputs are given (default 2) | the per-file commands, not `concat`, `mux`, `subs`, `watermark`, or `mix` |

Pass several files to any per-file command and they process in parallel:

```bash
goff compress *.mkv -j 4
```

`goff --help` lists every command grouped by what it does, and `goff <command> --help` carries an example of the invocation you want. A run where any file failed exits non-zero.

### Compress

`compress` re-encodes for playback rather than for archival, forcing 8-bit color, a constant frame rate, stereo 48kHz AAC, and a faststart MP4. [docs/playback.md](docs/playback.md) covers each change, the flags that opt out of it, and how to verify the result.

```bash
goff compress input.mkv                       # H.265, 1080p cap, 8-bit, CFR, 48kHz stereo, faststart
goff compress input.mkv --codec av1 --crf 28
goff compress input.mp4 --size 25MB           # bitrate solved to land under 25MB
goff compress input.mkv --height 720          # 1280x720 cap, HDR sources included
goff compress master.mov --lossless           # no video quality loss, source resolution kept
goff compress input.mkv --keep-10bit --keep-hifi-audio   # for VLC rather than a browser
goff compress drifting.mp4 --copy-video       # repair audio sync, video untouched
goff compress surround.mkv --copy-audio       # 5.1 or hi-fi audio passed through untouched
goff compress film.mkv --fps 24               # cap the frame rate, cutting frames and encode time
goff compress input.mkv --preset slow --audio-bitrate 192k --audio-rate 44100
```

### Containers and streaming

```bash
goff remux input.mkv --to mp4     # no re-encoding
goff remux capture.mkv --to mp4 --fix-timestamps
goff hls input.mp4 --segment-type fmp4   # writes input.hls-fmp4/index.m3u8 plus segments
goff hls input.mp4 --segment-type ts --segment-duration 4
```

### Tracks

`--audio-track` and `--sub-track` take `all`, `none`, a stream index, or a language code such as `eng`, and `compress`, `remux`, and `extract` all accept both. [docs/tracks.md](docs/tracks.md) covers what each container carries and how `extract` names the files it writes.

```bash
goff compress input.mkv --audio-track eng --sub-track none
goff remux film.mkv --to mp4 --audio-track eng
goff extract film.mkv --to vtt              # one sidecar per text subtitle track
```

### Audio

```bash
goff extract video.mp4 --to mp3 --bitrate 320k
goff extract video.mp4 --to wav --rate 16000 --channels 1   # ready for a transcriber
goff convert song.wav --to opus --rate 48000
goff normalize podcast.wav --to mp3          # -16 LUFS, -1.5 dBTP
goff normalize lecture.mp4 --lufs -14

goff mix voice.wav music.mp3                 # both play from 0
goff mix voice.wav music.mp3:at=5:vol=0.3    # music enters 5s in, at 30% level
goff mix a.mp3 b.mp3 --fit shortest          # stop at the shortest input
```

`mix` layers tracks so they play at the same time. To join clips end to end instead, use `concat`.

### Transforms

One verb takes a flag per dimension, and any combination of them composes into a **single** encode instead of stacking generations of quality loss. At least one is required:

```bash
goff transform clip.mp4 --rotate 90     # 90, 180, 270, hflip, vflip
goff transform clip.mp4 --scale 720p    # 4k, 2160p, 1440p, 2k, 1080p, fhd, 720p, hd, 480p, sd, or 1280x720
goff transform clip.mp4 --speed 1.5     # audio pitch corrected
goff transform clip.mp4 --crop          # 9:16 for Shorts / Reels
goff transform clip.mp4 --mute

goff transform clip.mp4 --rotate 90 --scale 720p --mute   # one encode, three changes
```

### Segments

```bash
goff trim video.mp4 --start 00:01:30 --end 00:02:00
goff trim video.mp4 --start 90 --duration 30 --accurate
goff gif screencast.mp4 --width 640 --fps 20 --start 5 --duration 8
goff gif screencast.mp4 --to webp
goff thumbnail movie.mp4                              # JPEG still from the midpoint, 640px wide
goff thumbnail movie.mp4 --at 00:01:30 --width 1280
```

### Combining

```bash
goff concat part1.mp4 part2.mp4 part3.mp4
goff concat a.mp4 b.mp4 --reencode              # for mismatched codecs

goff mux talk.mp4 --audio music.mp3:vol=0.3     # music mixed under the original audio
goff mux talk.mp4 --audio dub.m4a --audio-mode replace
goff mux film.mkv --audio en.m4a --audio fr.m4a --audio-mode separate
goff mux clip.mp4 --audio bed.mp3 --fit longest # run to the end of the music

goff subs video.mp4 --subtitles subs.srt        # embed as a track
goff subs video.mp4 --subtitles subs.eng.srt --lang eng --default

goff watermark video.mp4 --logo logo.png --position bottom-right --width 12 --opacity 0.6
```

### Inspection

```bash
goff inspect movie.mkv
goff inspect movie.mkv --json     # the same reading as a data contract
goff inspect movie.mp4 --check    # timeline gaps, A/V drift, browser playability
```

[docs/browser-safe.md](docs/browser-safe.md) is the runbook for taking an arbitrary set of files to a browser-safe state, covering what `--check` reports and the cheapest fix for each issue it names.

### Scripting and agents

Styled output is a property of the destination rather than a flag: piping any command anywhere strips the colors and keeps every progress line instead of redrawing one. On a terminal a run given several inputs shows a single meter counting files, so a large batch reports its position instead of scrolling one line per file. That becomes one meter per file when `--jobs 1` and at most 15 inputs make the run sequential and short enough to watch. `--debug` swaps the styled tier for structured logs, and `inspect --json` emits a stable struct instead of a table to scrape:

```bash
goff compress video.mp4 --debug
goff inspect video.mp4 --json | jq '.streams[] | select(.type == "audio")'
goff inspect video.mp4 --json --check | jq '.conformance.browserSafe'
```

Under `--debug`, an encode reports progress as one JSON object a second carrying `current`, `total`, `unit`, `percent`, `rate`, and `eta`, with `message` naming the file and `percent` absent when the source duration is unknown. A run given several inputs interleaves those per file and drops the summary table, which the normal tier keeps even when piped. A parent process reads those objects rather than scraping the bar, and a tool wanting goff's encode contract runs the binary rather than reproducing its FFmpeg arguments.

## Notes

- **Composed transforms**: a `transform` run given one flag names its output after that change, such as `clip.rot90.mp4`, and a run given two or more names it `<name>.transform.<ext>`.
- **Input offsets and levels**: `mix` and `mux --audio` accept `<file>:at=<time>` to delay a track and `:vol=<factor>` to change its level, both optional and in either order. A track with neither starts at 0 at its own level.
- **Audio modes**: `mux --audio-mode` decides what happens to the audio a video already has. `mix` layers it with the new tracks into one, `replace` drops it, and `separate` keeps every track selectable, which leaves nothing for `at=` or `vol=` to apply to and so rejects them.
- **HLS layout**: each packaged video gets its own directory holding `index.m3u8` and the segments, so two packaged videos never share a segment name. fMP4 packaging adds an `init.mp4` next to them.
- **Timestamp shifting**: `remux --fix-timestamps` moves a negative start time to zero, which matters for captures whose audio and video begin at different points.
- **Out-of-range numbers**: a numeric flag given a value past its range is rejected before any work starts, and `--help` prints the accepted range as the flag's type, such as `--crf 1..63`. H.265 and H.264 both cap at 51, so a `--crf` above that lands at 51 for either codec, while AV1 uses the full range.
- **Thumbnails**: `thumbnail` seeks before decoding and writes one JPEG at quality 2, defaulting to the midpoint of the file and 640px wide with the height following the aspect ratio. An `--at` at or past the last frame is rejected before FFmpeg runs.
- **Container bitrate**: `inspect` prints the container bitrate on its summary line, deriving it from size over duration when FFprobe reports none. A derived figure can differ slightly from a reported one, and a file FFprobe gives no duration for shows `-`.
- **Interrupting a run**: Ctrl+C or `SIGTERM` stops FFmpeg and deletes the partial output, then reports the run as cancelled and exits non-zero.
