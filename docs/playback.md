# Playback normalization

A file that plays in VLC often will not play in a browser. `compress` re-encodes for the browser case, and this covers every default it applies, the flags that override them, and how to confirm the result before shipping the file.

## What every compress run changes

| Change | Flag that changes it |
|---|---|
| The default video track is kept, or the first when none is flagged | none |
| Color is forced to 8-bit `yuv420p` | `--keep-10bit` keeps the source depth |
| The frame rate is made constant | `--fps` caps it |
| Audio is re-encoded to AAC at 128k | `--audio-bitrate`, or `--copy-audio` to skip the re-encode |
| Audio is resampled to 48kHz | `--audio-rate`, or `--copy-audio` |
| Audio is downmixed to stereo | `--keep-hifi-audio` keeps the source layout |
| H.265 output is tagged `hvc1`, which Safari and the Apple media stack require | none |
| The MP4 index is moved to the front so playback can start before the file finishes downloading | none |

The video codec defaults to H.265 at CRF 30. `--codec av1` uses SVT-AV1 at CRF 32 and `--codec h264` uses x264 at CRF 23. All three take the same ten `--preset` words and default to `medium`.

## Resolution

`--height` caps the output height and derives the width cap from it at 16:9, so `--height 720` gives a 1280x720 cap and `--height 2160` gives 3840x2160. The default cap is 1080. Output is never upscaled, so a 720p source under `--height 2160` stays 720p.

## Frame rate

`--fps` caps the output frame rate. Encode cost tracks frame count rather than duration, so capping a 30fps source to 24 removes a fifth of the frames and close to a fifth of the encode time.

The cap never raises a rate. `--fps 60` on a 30fps source keeps 30fps and reports that it did, because duplicating frames costs size and buys nothing.

## HDR tone-mapping

HDR10 and HLG sources are tone-mapped to 8-bit SDR with the Hable curve, under the same `--height` cap as an SDR source. `--keep-10bit` skips the tone-map and keeps the source grade, which is what a VLC-bound file wants and a browser-bound file does not.

## Audio sync

`compress` fills holes in the source audio timeline, so the output holds exactly as much audio as its timeline claims. Without that, forcing a constant frame rate lengthens the video while leaving the audio short, and the two drift apart by however much the source was missing. VLC hides the problem by honoring the timestamps, and a browser playing the file directly does not.

## Repairing a file in seconds

`compress --copy-video` copies the video stream through and re-encodes only the audio, which fixes drift and a missing `hvc1` tag without a full encode. It rejects every video-encoding flag: `--codec`, `--crf`, `--fps`, `--height`, `--size`, `--lossless`, `--preset`, and `--keep-10bit`.

```bash
goff compress drifting.mp4 --copy-video
```

The output is named `<name>.repaired.<ext>`.

## Keeping the source audio

`--copy-audio` passes the audio streams through untouched instead of re-encoding them to AAC. It is the right move when the source audio is worth more than a 128k stereo downmix, such as a 5.1 track or a high-bitrate original.

```bash
goff compress surround.mkv --copy-audio
```

It rejects `--audio-bitrate`, `--audio-rate`, and `--keep-hifi-audio`, which all shape an encode that no longer happens. Paired with `--copy-video` it re-encodes nothing at all, so that combination is rejected in favor of `goff remux`.

MP4 holds more than AAC. `-c:a copy` accepts AAC, MP3, AC-3, E-AC-3, Opus, FLAC, ALAC, and PCM, and ffmpeg refuses to write the header for anything the container cannot carry. Nothing is filtered ahead of time.

Copied audio skips the sync filter that `compress` otherwise applies, so gaps in the source audio timeline survive into the output. Anything other than AAC also leaves the file outside the browser-safe set, which `inspect --check` reports either way.

## Presets and CRF

`--preset` takes one of ten words on every codec: `ultrafast`, `superfast`, `veryfast`, `faster`, `fast`, `medium`, `slow`, `slower`, `veryslow`, `placebo`. x265 and x264 take those words directly. SVT-AV1 takes an integer instead, so goff translates, mapping `medium` to 6, `slow` to 4, `veryslow` to 2, and `ultrafast` to 13.

CRF is not a constant quality across presets, so a slower preset does not shrink a file on its own. Measured on one 60-second 1080p segment, `medium` at CRF 32 gave 1627 kbps at VMAF 91.7. `slower` at the same CRF spent more bits rather than fewer, and reached VMAF 94.7. `slower` at CRF 35 returned to VMAF 91.8 for 1072 kbps.

A slower preset spends its extra effort on quality rather than on size. The saving appears only when CRF rises alongside it, so move the two together rather than either alone.

## When re-encoding costs more than it saves

A source already at or below the bitrate the target codec needs for the same quality has no headroom left. Re-encoding it spends hours to produce a larger file at lower quality. One file in a real library reached 132% of its source size at CRF 32 for exactly that reason.

`--copy-video` is the answer whenever the video is already efficient and only the container or the audio is wrong.

## Lossless

`compress --lossless` keeps the source resolution, bit depth, and channel layout, and skips tone-mapping, since all of those are lossy. It re-encodes the video into the target codec with compression that discards nothing, which is not the same as `--copy-video` leaving the original bitstream alone. A lossy source re-encoded this way usually grows several times larger, so the flag is for masters and intermediates rather than for finished files.

Audio is still re-encoded to AAC unless `--copy-audio` is given alongside. It cannot be combined with `--crf`, `--size`, `--fps`, or `--height`.

## Size targets

`--size` solves a video bitrate from the budget and the duration, then holds back headroom of 2% or 0.5MB, whichever is larger, so a 25MB budget targets 24.5MB and the muxed result stays under the limit. The output is named after the budget, so `--size 25MB` writes `<name>.25mb.mp4`. `--preset` still applies, since it trades encode time for quality at whatever bitrate the budget produced.

## Audio track order

`compress` writes the default audio track first and flags only that one. Chrome plays the first enabled track while Firefox plays the first track in file order, so a file where those two disagree plays different audio in each browser.

## Checking the result

`inspect --check` reads every packet, then reports how much real content each stream holds against the timeline it declares:

```
$ goff inspect movie.mp4 --check
→ Content: video 6.000s | audio 6.037s | drift -0.037s
→ Start: video 0.000s | audio -0.021s | offset +0.021s
→ Timeline gaps: video 0 (0.000s) | audio 0 (0.000s)
✓ Browser-safe
```

Content is how much media each stream actually carries, drift is the difference between the two, and offset is how far apart they start. A file is reported browser-safe only when all of the following hold:

- The container is MP4 or MOV.
- The video codec is H.264, H.265, or AV1, and H.265 is tagged `hvc1`.
- The pixel format is 8-bit or 10-bit 4:2:0.
- The primary audio track is AAC, at most stereo, at 48kHz.
- The first audio track carrying the default flag is the first audio track.
- Neither timeline has a gap.
- Drift and start offset are both within 0.1s.

The same reading is available as a data contract under `conformance`:

```
$ goff inspect movie.mp4 --json --check | jq -c '.conformance | {driftSeconds, startOffsetSeconds, browserSafe}'
{"driftSeconds":-0.037333333333332774,"startOffsetSeconds":0.021333333333333333,"browserSafe":true}
```

The check costs a fraction of a second on a feature-length file. It is off by default because the plain reading needs only the header.
