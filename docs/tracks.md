# Tracks and subtitles

A rip carries several audio and subtitle tracks, and not every container holds all of them. This covers which tracks goff keeps, which it converts, which it drops, and how it names the files it writes.

## Selecting tracks

`compress`, `remux`, and `extract` all take `--audio-track` and `--sub-track`, both defaulting to `all`:

| Value | Selects |
|---|---|
| `all` | every track of that kind |
| `none` | no track of that kind |
| a stream index, such as `2` | the one stream at that index, as `inspect` prints it |
| a language code, such as `eng` | every track tagged with that language |

A selector matching nothing is an error rather than a silent drop:

```
$ goff compress film.mkv --audio-track jpn
✗ cannot compress film.mkv: no audio stream is tagged jpn
```

## What each container carries

An MKV target carries every selected track through untouched, including image subtitles and font attachments.

An MP4 or MOV target converts text subtitles to `mov_text`, and copies them straight through when they already are. It holds no image subtitle format, so PGS and VobSub tracks are dropped, and no attachment format, so embedded fonts are dropped. A run that drops either kind says so on a warning line naming the count.

## Extracting to sidecars

`extract` writes one file per selected track, named `<name>.<language>.<ext>`. A track carrying no language tag falls back to `.audio.` or `.subs.`. Every file is named as it is written:

```
$ goff extract film.mkv --to vtt
✓ film.mkv → 2 files (0s, 4.6 MB → 138 B)
  ✓ film.eng.vtt (69 B)
  ✓ film.fra.vtt (69 B)
```

`--to srt` and `--to vtt` read subtitle tracks. Every other format reads audio.

A flag that does not apply to the chosen `--to` is rejected before any work starts. A subtitle target rejects `--bitrate`, `--rate`, `--channels`, and `--audio-track`; an audio target rejects `--sub-track`.

```
$ goff extract film.mkv --to vtt --bitrate 320k
✗ --bitrate does not apply to --to vtt
```

`-o` names one destination, so it works only when exactly one track matches the selector.

## Subtitles in a browser

Only Safari renders a subtitle track stored inside an MP4, so a browser player needs a sidecar. `goff extract film.mkv --to vtt` writes one `.vtt` per text subtitle track for a page to attach with a `<track>` element.

Going the other way, `subs` adds one subtitle file to a video as a track:

```bash
goff subs video.mp4 --subtitles subs.srt
goff subs video.mp4 --subtitles subs.eng.srt --lang eng --default
```

`--default` clears the default flag from the tracks already there and sets it on the one being added.
