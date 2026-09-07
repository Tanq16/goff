<div align="center">
  <img src=".github/assets/logo.svg" alt="goff Logo" width="200">
  <h1>goff</h1>

  <a href="https://github.com/Tanq16/goff/actions/workflows/release.yaml"><img alt="Build Workflow" src="https://github.com/Tanq16/goff/actions/workflows/release.yaml/badge.svg"></a>&nbsp;<a href="https://github.com/Tanq16/goff/releases"><img alt="GitHub Release" src="https://img.shields.io/github/v/release/Tanq16/goff"></a><br><br>
  <a href="#capabilities">Capabilities</a> &bull; <a href="#install">Install</a> &bull; <a href="#usage">Usage</a> &bull; <a href="#notes">Notes</a>
</div>

---

**goff** (Go FFmpeg) is a terminal media suite for FFmpeg: one verb per operation, sensible defaults, and no filter syntax to look up.

It exists so you stop looking up filter syntax for the same dozen jobs. It is not a video editor or a replacement for FFmpeg itself.

## Capabilities

| Category | Commands | Description |
|----------|----------|-------------|
| **Compression** | `compress` | H.265 / AV1 / H.264 re-encoding, height caps, target file size, lossless mode, automatic HDR tone-mapping, and playback normalization that keeps audio locked to video |
| **Containers** | `remux`, `hls` | Lossless container switching, VoD HLS packaging as fMP4 or MPEG-TS |
| **Audio** | `extract`, `convert`, `normalize`, `mix` | Audio and subtitle extraction from video, format transcoding, EBU R128 loudness normalization, layering tracks into one |
| **Transforms** | `transform` | Rotation and flips, resolution tiers, pitch-corrected speed, 9:16 vertical crop, audio removal, composed into one encode |
| **Segments** | `trim`, `gif`, `thumbnail` | Time-range cuts, 2-pass palettegen GIF and animated WebP, single-frame JPEG stills |
| **Multi-file** | `concat`, `mux`, `subs`, `watermark` | Clip joining, external audio muxing, subtitle embedding, logo overlays |
| **Inspection** | `inspect` | Stream table with codecs, resolutions, languages, per-stream and container bitrates, and HDR transfer characteristics, plus a packet-level browser playability check |

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

Requires Go 1.26+:

```bash
git clone https://github.com/Tanq16/goff.git
cd goff
make build
```

## Usage

Every command takes one or more input files and writes alongside them, so nothing is overwritten by default.

| Flag | Effect | Where |
|------|--------|-------|
| `--debug` | Structured logs, including the underlying FFmpeg error | every command |
| `-o`, `--output` | Explicit output path, overwritten if it exists, single input only | every command except `inspect` |
| `-j`, `--jobs` | Concurrent encodes when several inputs are given (default 2) | the per-file commands, not `concat`, `mux`, `subs`, `watermark`, or `mix` |

Pass several files to any per-file command and they process in parallel:

```bash
goff compress *.mkv -j 4
```

`goff --help` lists every command grouped by what it does, and `goff <command> --help` carries an example of the invocation you want. A run where any file failed exits non-zero.

### Compress

```bash
goff compress input.mkv                       # H.265, 1080p cap, 8-bit, CFR, 48kHz stereo, faststart
goff compress input.mkv --codec av1 --crf 28
goff compress input.mp4 --size 25MB           # bitrate solved to land under 25MB
goff compress input.mkv --height 720
goff compress master.mov --lossless           # no video quality loss, source resolution kept
goff compress input.mkv --keep-10bit --keep-hifi-audio   # for VLC rather than a browser
goff compress drifting.mp4 --copy-video       # repair audio sync, video untouched
goff compress input.mkv --audio-track eng     # keep one language instead of every track
goff compress input.mkv --sub-track none      # drop subtitles instead of carrying them
goff compress input.mkv --preset slow --audio-bitrate 192k --audio-rate 44100
```

### Containers and streaming

```bash
goff remux input.mkv --to mp4     # no re-encoding
goff remux capture.mkv --to mp4 --fix-timestamps
goff remux film.mkv --to mp4 --audio-track eng   # leave the other language tracks behind
goff hls input.mp4 --segment-type fmp4   # writes input.hls-fmp4/index.m3u8 plus segments
goff hls input.mp4 --segment-type ts --segment-duration 4
```

### Audio

```bash
goff extract video.mp4 --to mp3 --bitrate 320k
goff extract video.mp4 --to wav --rate 16000 --channels 1   # ready for a transcriber
goff extract film.mkv --to vtt                              # one sidecar per text subtitle track
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

### Multi-file

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

### Scripting and agents

Styled output is a property of the destination rather than a flag: piping any command anywhere strips the colors and keeps every progress line instead of redrawing one. `--debug` swaps the styled tier for structured logs carrying the underlying FFmpeg error, and `inspect --json` emits a stable struct instead of a table to scrape:

```bash
goff compress video.mp4 --debug
goff inspect video.mp4 --json | jq '.streams[] | select(.type == "audio")'
goff inspect video.mp4 --json --check | jq '.conformance.browserSafe'
```

Piped under `--debug`, a single-input run reports progress as one JSON object a second carrying `percent`, `currentSeconds`, and `totalSeconds`, with a final object at completion. A run given several inputs at once reports per-file results instead, with no progress objects. A parent process reads those rather than scraping the bar, and a tool wanting goff's encode contract runs the binary rather than reproducing its FFmpeg arguments.

## Notes

- **Safe output naming**: outputs are written as `<name>.<operation>.<ext>` next to the input, incrementing to `.1.<ext>` on a collision, so an existing file is never replaced. `-o` is the exception, since it names the destination outright.
- **Composed transforms**: giving `transform` two or more flags produces one encode named `<name>.transform.<ext>`, while a single flag keeps its descriptive name, such as `clip.rot90.mp4`.
- **Input offsets and levels**: `mix` and `mux --audio` accept `<file>:at=<time>` to delay a track and `:vol=<factor>` to change its level, both optional and in either order. A track with neither starts at 0 at its own level.
- **Audio modes**: `mux --audio-mode` decides what happens to the audio a video already has. `mix` layers it with the new tracks into one, `replace` drops it, and `separate` keeps every track selectable.
- **HLS layout**: each packaged video gets its own directory holding `index.m3u8` and the segments, so two packaged videos never share a segment name. fMP4 packaging adds an `init.mp4` next to them.
- **Size targets**: `--size` holds back headroom below the number you give, so a 25MB budget targets 24.5MB and the muxed result stays under the limit.
- **HDR tone-mapping**: HDR10 and HLG sources are tone-mapped to 8-bit SDR with the Hable curve during `compress`, which takes priority over `--height` for those inputs. `--keep-10bit` skips it and keeps the source grade.
- **Lossless compression**: `compress --lossless` keeps the source resolution, bit depth, and channel layout, and skips tone-mapping, since all of those are lossy. Audio is still re-encoded to AAC like every other `compress` run. It cannot be combined with `--crf`, `--size`, or `--height`.
- **Playback normalization**: every `compress` run takes the first video track, keeps every audio track with the default one first, forces 8-bit color and a constant frame rate, downmixes audio to stereo at 48kHz, tags H.265 as `hvc1` so Safari and the Apple media stack accept it, and writes a faststart MP4. `--keep-10bit` and `--keep-hifi-audio` opt out of the first two, and `--audio-rate` sets the third.
- **Audio sync**: `compress` fills holes in the source audio timeline so the output holds exactly as much audio as its timeline claims. Without that, forcing a constant frame rate lengthens the video while leaving the audio short, and the two drift apart by however much the source was missing. VLC hides the problem by honoring the timestamps; a browser playing the file directly does not.
- **Repairing a file**: `compress --copy-video` re-encodes only the audio and copies the video stream through, which fixes drift and the `hvc1` tag in seconds rather than a full encode. It cannot be combined with any video-encoding flag.
- **Track selection**: `--audio-track` and `--sub-track` take `all`, `none`, a stream index as `inspect` prints it, or a language code such as `eng`, and default to `all` on `compress`, `remux`, and `extract`. A selector matching nothing is an error rather than a silent drop.
- **Subtitles and containers**: text subtitle tracks are converted to `mov_text` for an MP4 and copied untouched into an MKV. Image tracks such as PGS and VobSub survive only an MKV target, since MP4 holds no image subtitle format, and a run that drops them says so.
- **Subtitles in a browser**: only Safari renders a subtitle track stored inside an MP4, so a browser player needs a sidecar. `goff extract film.mkv --to vtt` writes one `.vtt` per text track for a page to attach with a `<track>` element.
- **Extraction naming**: `extract` writes one file per selected track, named `<name>.<language>.<ext>`, falling back to `.audio.` or `.subs.` when a track carries no language tag. `--to srt` and `--to vtt` read subtitle tracks, and every other format reads audio.
- **Audio track order**: `compress` writes the default audio track first and flags only that one, because Chrome plays the first enabled track and Firefox plays the first track in file order. `inspect --check` reports a file where those two disagree.
- **Timestamp shifting**: `remux --fix-timestamps` moves a negative start time to zero, which matters for captures whose audio and video begin at different points.
- **Container bitrate**: `inspect` prints the container bitrate on its summary line, deriving it from size over duration when FFprobe reports none. A derived figure can differ slightly from a reported one, and a file FFprobe gives no duration for shows `-`.
- **Conformance check**: `inspect --check` reads every packet to report how much real content each stream holds against the timeline it declares, then lists what would stop a browser playing the file. It costs a fraction of a second on a feature-length file, and is off by default because the plain reading needs only the header.
- **Out-of-range numbers**: a numeric flag given a value past its range is rejected before any work starts. `--help` prints the accepted range as the flag's type, such as `--crf 1..63`. H.265 and H.264 both cap at 51, so a `--crf` above that lands at 51 for either codec, while AV1 uses the full range.
- **Thumbnails**: `thumbnail` seeks before decoding and writes one JPEG at quality 2, defaulting to the midpoint of the file and 640px wide with the height following the aspect ratio. An `--at` past the end of the file is rejected before FFmpeg runs.
- **Failure detail**: a failed encode reports the FFmpeg error only under `--debug`, which keeps a wall of filter-graph text out of normal runs.
