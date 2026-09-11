# Browser-safe media with goff

Instructions for an agent handed an arbitrary set of media files and asked to make them play in a browser. Read this, survey the files, classify each one, then act. Everything needed to decide is here.

## The goal

A browser is not a media player. It decodes a narrow set of containers and codecs, renders one audio track with no way to switch, and ignores subtitles inside the container everywhere except Safari. A file that plays perfectly in VLC can be silent, black, or refuse to load in Chrome. The job is to get every file to the point where `goff inspect --check` reports `Browser-safe`, spending the least possible work to get there.

Playability in a desktop player is a separate goal with different answers. When an archive copy is wanted, that is `--keep-10bit --keep-hifi-audio`, it will not be browser-safe, and that is correct.

## Invariants

These are settled preferences. Do not re-derive them per file and do not offer alternatives.

- **MP4 is the container.** Always, for anything a browser touches.
- **HEVC is the re-encode target.** It is the reason the library fits on disk. Anything outside H.264, HEVC, and AV1 gets re-encoded to HEVC. All three are accepted terminal states, so H.264 and AV1 already present are never re-encoded for their codec alone.
- **10-bit stays.** `yuv420p10le` is browser-safe. Never spend a re-encode converting 10-bit to 8-bit. When re-encoding for another reason, pass `--keep-10bit` if the source is 10-bit.
- **Subtitles never live inside the video.** No `mov_text`, no burned-in text, no embedded track of any kind, in any output. Every subtitle becomes a sidecar `.vtt` file. This means `--sub-track none` on every `remux` and every `compress`, without exception, because both commands keep and transcode subtitles by default.
- **Text subtitles only.** Bitmap subtitles (`hdmv_pgs_subtitle`, `dvd_subtitle`, and similar) are ignored. They cannot become VTT.
- **All audio tracks are kept.** Nothing is discarded for being a second language. Both `remux` and `compress` keep every track by default.
- **English is the default audio track.** Set it explicitly with `remux --default-audio eng`. Firefox plays the first track in file order while Safari plays the default-flagged one, so the track must be moved to first position, not merely re-flagged.
- **Audio ends up AAC-LC, stereo, 48000 Hz.** That is the only combination every engine plays.
- **Lossless first.** A remux that rewrites container metadata beats a re-encode that touches audio, which beats a full video re-encode. Escalate only when the cheaper step cannot fix the actual failure.

## Step 1: survey

```
goff inspect <file> --check --json
```

Run it on every file. It reads every packet, so budget a few seconds per file on local disk and considerably longer on a network or USB mount. Parallelism follows the disk rather than the CPU: on spinning or USB media, four concurrent jobs typically saturates it, and raising that adds seek contention rather than throughput. Measure utilisation before going higher.

The JSON carries `conformance.browserSafe`, `conformance.issues`, and the per-stream summary. Bucket on the issue list, not on `browserSafe` alone, because the issue names the cheapest fix.

## Step 2: what `browserSafe` requires

Every one of these must hold:

| check | requirement |
|---|---|
| container | MP4 or MOV |
| video codec | H.264, H.265, or AV1 |
| HEVC tag | `hvc1`, since Safari refuses `hev1` |
| pixel format | `yuv420p` or `yuv420p10le` |
| video timeline | zero gaps |
| audio codec | AAC |
| audio channels | at most stereo |
| audio sample rate | exactly 48000 Hz |
| audio timeline | zero gaps |
| audio track order | the first default-flagged track is also first in file order |
| A/V drift | within 0.1 s |
| A/V start offset | within 0.1 s |

The last two do not block playback. They are in the verdict because `--check` is a verification gate, not an enforcement gate, and a file failing them is worth a human look.

## Step 3: decide

```
issues on the file
    │
    ├─ none ─────────────────────────────────────► done, no video work
    │
    ├─ container, hvc1 tag, or audio track order only
    │       └─► remux                       lossless, sub-second per file
    │
    ├─ audio codec, channel count, or sample rate wrong
    │       └─► compress --copy-video       video untouched, tens of seconds
    │           this covers AV1 with Opus audio, which is common
    │
    ├─ audio or video timeline gaps
    │       └─► compress --copy-video       rebuilds the audio timeline
    │           gaps persist? full re-encode
    │
    ├─ video codec outside H.264, HEVC, and AV1, or pixel format not 4:2:0
    │       └─► compress                    full HEVC re-encode, hours
    │
    ├─ video bitrate far above what the content needs
    │       └─► compress --crf 24           near-lossless shrink, opt-in
    │
    └─ drift or start offset only
            └─► sample and hand to a human. Do not re-encode.
```

Subtitles are independent of all of the above and always run, from the **source** file:

```
goff extract <source> --to vtt --sub-track <index>
```

**Order: extract before remux or compress.** Both write outputs with no subtitle track, so extracting afterwards fails with `no text subtitle stream to extract`. Extracting first also means the source can be deleted the moment the new file verifies.

### When a file needs two passes

`compress` promotes whichever track the source already flagged as default; it cannot name a different one. So a file that needs an audio re-encode **and** has the wrong language flagged as default takes both steps:

```
goff compress <file> --copy-video --sub-track none -o <work>
goff remux <work> --to mp4 --default-audio eng --sub-track none -o <final>
```

Run it the other way round only if the remux is the expensive part, which it never is.

## What to decide alone, and what to escalate

**Act without asking** when the issue is mechanical and the fix is deterministic:

- container, `hvc1` tag, audio track order, audio codec, channel count, sample rate
- video codec or pixel format outside the browser-safe set
- timeline gaps

**Sample and get human confirmation** before acting:

- **drift or start offset outside tolerance.** These are content observations rather than defects. Drift means one stream's timeline is longer than the other's, typically a dub that ends before the video does. Offset means a track starts late. Neither breaks playback. Extract two or three representative files, hand over the exact timestamps to seek to, and let a person watch before anything is re-encoded.
- **a re-encode purely to save space.** Shrinking is a budget decision, not a conformance one. Report the current bitrate, the estimated output size, and the wall-clock cost, and wait.
- **anything where the cheap fix has already failed once.** A remux that still fails `--check`, or a `--copy-video` that leaves gaps behind, means the assumption was wrong. Say so rather than escalating straight to a multi-hour re-encode.

Never overwrite a source on a judgment call. See the verification gate below.

## Commands

### Inspect

```
goff inspect <file>                  # streams, codecs, bitrates, no packet read
goff inspect <file> --check          # adds the conformance verdict, reads every packet
goff inspect <file> --check --json   # same, machine-readable
```

### Remux, the lossless path

```
goff remux <file> --to mp4 --default-audio eng --sub-track none
```

Fixes container, `hvc1` tag, and audio track order. Fixes nothing else. No stream is re-encoded.

- `--default-audio` takes a language code or a stream index and moves that track to first position while stamping the dispositions. Without it, source order survives and the track-order issue with it.
- `--sub-track none` is mandatory. The default is `--sub-track all`, and since MP4 accepts only `mov_text`, remux will silently transcode subtitles into an embedded track.
- Attachments are always dropped with a warning. Those are subtitle fonts, and no MP4 can hold them.
- MKV chapters come through as a `bin_data` stream that `inspect` lists as `data`. Benign.
- `--fix-timestamps` emits `-avoid_negative_ts make_zero`, which only shifts *negative* timestamps to zero. It does nothing to a track that starts late.

### Compress, when a stream has to change

```
goff compress <file> --copy-video --sub-track none --audio-bitrate 192k
goff compress <file> --sub-track none --crf 24 --keep-10bit
goff compress <file> --sub-track none --keep-10bit --keep-hifi-audio   # archive, not browser-safe
```

- `--copy-video` rewrites the container and the audio and leaves every video packet untouched. Tens of seconds per file against hours for a full re-encode. This is the repair tool, and it is the whole answer for an AV1 file whose only fault is its audio.
- `--copy-audio` does the reverse, passing the audio through and re-encoding only the video. It cannot produce a browser-safe file from non-AAC audio, so it belongs to archive copies rather than to this workflow.
- `--crf 24` is the near-lossless setting for a deliberate shrink. Lower is better quality; the codec default applies when the flag is absent.
- `--height` defaults to 1080, so a 4K source is downscaled unless told otherwise. Set it explicitly when the source resolution must survive.
- `--keep-10bit` keeps the source bit depth and skips HDR tone-mapping. Pass it whenever the source is 10-bit.
- `--keep-hifi-audio` keeps the source channel layout instead of downmixing to stereo. Browser-incompatible by definition; archive copies only.
- `--audio-track` selects which tracks to *keep*, not which becomes default. Leave it at `all`.
- `--lossless` and `--size` exist for cases where quality or a byte budget is fixed in advance.

### Extract subtitles

```
goff extract <file> --to vtt --sub-track <index>
goff extract <file> --to vtt                    # every text track, one file each
```

The output is cleaned automatically: inline `<b>`/`<i>` markup stripped, ASS `{...}` override blocks stripped, exact duplicate and empty cues dropped, cue identifiers removed, timings normalised to `HH:MM:SS.mmm`.

**Pick the track by index, not by language, when several tracks share a language.** The default-flagged subtitle track is frequently a signs-and-songs or forced-narrative track rather than the dialogue. Read the titles first:

```
ffprobe -v error -select_streams s -show_entries stream=index:stream_disposition=default,forced:stream_tags=language,title -of csv=p=0 <file>
```

A dialogue track has several thousand cues and full sentences; a signs track has fewer and reads as captions on objects. Compare cue counts when the titles are unhelpful.

### Where the sidecar goes

Next to the video, never inside it. Name it `subtitles.vtt` in a one-video-per-directory layout where the player expects a fixed name, or `<video basename>.<lang>.vtt` where several files share a directory. A player loads it through a `<track>` element, which is the only subtitle path that works in every engine.

## Reading `inspect --check`

```
Content: video 1448.989s | audio 1449.045s | drift -0.056s
Start:   video 0.000s | audio 1.323s | offset -1.323s
Timeline gaps: video 0 (0.000s) | audio 0 (0.000s)
```

- **Content** is each timeline's extent. **drift** is video minus audio. Positive means the audio ends first; negative means the audio outlasts the video.
- **Start** is where each stream's first packet sits. **offset** is video minus audio. Negative means the audio starts later than the video.
- **Timeline gaps** count intervals between consecutive packets exceeding 1.5x the observed modal period. This is the one that matters: gaps are genuine discontinuities and they do cause progressive desync.

Drift and offset are computed from the container's own timestamps, so they are exact in any container and comparable across a remux.

### What drift and offset mean in practice

Verified by playing deliberately damaged files in a real browser:

| case | what the browser does |
|---|---|
| audio ends 64 s before video | plays, silent tail, everything else correct |
| audio ends 25 s before video | same |
| audio outlasts video by 3.5 s | holds the last frame, no error |
| audio starts 1.3 s late | holds audio until that point, then perfect sync for the whole file |

The offset result is the decisive one: a late-starting track keeps sync throughout, so the browser honours the container's per-track start time rather than forcing audio to zero. Treat both as observations to confirm with a person, never as a reason to re-encode.

Untested: a track starting *before* the video, the negative direction. Escalate that case rather than guessing.

## Verified browser behaviour

Established against vendor source. Do not re-derive these.

| | Chrome and Edge | Firefox | Safari |
|---|---|---|---|
| subtitle track inside the MP4 | never rendered | never rendered | rendered |
| sidecar `.vtt` via `<track>` | yes | yes | yes |
| several audio tracks | plays one, no switching | plays one, no switching | switchable |
| which audio track wins | first with the MP4 `tkhd` enabled bit | first in file order | per media selection group |

Citations: Chromium `media/filters/ffmpeg_demuxer.cc:1369-1371`, `:84-92`, `:1416-1434`; Gecko `dom/media/MediaFormatReader.cpp:1275` and `modules/libpref/init/StaticPrefList.yaml:11989`; WebKit `MediaPlayerPrivateAVFoundationObjC.mm`.

Two consequences drive the invariants above. Sidecar VTT is the only subtitle path that works everywhere, which is why nothing is ever embedded. And Firefox picking file order while Safari picks the default flag is why the English track must be moved to first position rather than only re-flagged.

**Safari cannot play Opus in an MP4 container.** No device, no version. This is independent of the video codec, which is why an AV1 file carrying Opus still needs `compress --copy-video` even though its video is already acceptable.

**AV1 needs a hardware decoder in Safari,** meaning an M3 or later Mac, an iPhone 15 Pro or later, or the M4 iPad Pro. Chrome 70+, Firefox 67+, and Edge 121+ decode it in software on every platform. AV1 is accepted flatly here regardless, since the devices that cannot play it are the ones being aged out.

**10-bit is not gated by any engine.** Chromium's `IsDecoderHevcProfileSupported` in `media/base/supported_types.cc` delegates entirely to the platform decoder with no bit-depth branch, and Firefox's `dom/media/platforms/PDMFactory.cpp` hands HEVC to the platform module with no profile restriction. The real limit is whether the viewing device decodes Main 10 in hardware.

**`-tag:v hvc1` under `-c copy` rewrites only the sample-entry fourcc.** Output is byte-identical in size with the same parameter-set NALs. It cannot be applied when a second non-HEVC video stream is present, such as cover art, because ffmpeg then refuses to write the header entirely.

## Verification gate

Before any output replaces its source, all of these must hold:

- `goff inspect <output> --check` reports no blocking issue
- HEVC output carries the `hvc1` tag
- zero video and audio timeline gaps
- duration within 1 s of the source
- output size above half the source size

Only then move it into place, and restore the source's mtime and mode afterwards. A remux that fails this escalates to `compress --copy-video`; a `--copy-video` that fails it escalates to a full re-encode.

When several files are processed in parallel, give each one a distinct work filename. A fixed work name means two workers in the same directory overwrite each other.
