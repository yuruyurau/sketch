# Recording & Contact Sheet

<div align="right">
  <a href="recording.ja.md">ja</a> | <b>en</b>
</div>

## record

Select a sketch and record it as a video.

Before recording starts you can override the following values (press Enter to accept the default):

| Option       | Description                    | Notes / Default                                            |
| ------------ | ------------------------------ | ---------------------------------------------------------- |
| fps          | Frame rate                     |                                                            |
| total        | Total frames to record         | Arithmetic expressions allowed (e.g. `60*30`). Integer ≥ 1 |
| start        | Starting frame                 | Advances the sketch by this many frames before recording   |
| codec        | `vp9` or `h264`                | `vp9` → webm, otherwise mp4                                |
| crf          | Quality (lower = better)       |                                                            |
| pixelDensity | pixelDensity during recording  |                                                            |
| timeout      | Overall timeout in minutes     | `0` or negative = disabled. Default: 60                    |
| maxFail      | Max consecutive frame failures | Default: 10                                                |
| p5Version    | p5.js version                  |                                                            |
| p5Source     | Source of p5.js                |                                                            |

Output directory is `outDir` from config.  
A temporary server is started on `record.port` (recording will not start if the port is already in use).

Recording uses the Chrome/Chromium binary specified in `record.browser.path` (required, launched headless).

### total and start

- `total` accepts integer arithmetic expressions (e.g. `60*30`, `1e4`, `1_000`, `0x10`). The result must be an integer ≥ 1. Division is integer division.
- `start` advances the sketch by that many frames before recording begins.

### timeout and maxFail

- `timeout`: overall timeout in minutes. `0` or negative disables it (default: 60).
- `maxFail`: maximum number of consecutive frame capture failures before aborting (default: 10).

## sheet

Captures selected frames from one or more sketches and generates a contact sheet image.

- Requires `magick` (ImageMagick) in PATH.
- Uses the same Chrome settings as recording.
- Target: `all` or `one`.
- Output: `sketchDir/sheet/sheet-YYYYMMDD_HHMMSS.png`.

### Options

| Option       | Description                      |
| ------------ | -------------------------------- |
| frames       | Frames to capture                |
| pixelDensity | pixelDensity for capture         |
| geometry     | ImageMagick `-geometry` format   |
| columns      | Number of columns (empty = auto) |
| label        | Whether to show labels           |
| background   | Background color                 |
| fill         | Fill color                       |

### frames syntax

- `30` → frame 30
- `1 5` / `1,5` → multiple frames
- `0..10` → 0 ≤ n < 10
- `0..10..2` → with step
- `0..100/5` → divide the range into 5 equal parts

### Notes

- Empty `columns` → automatic near-square grid.
- Sketches listed in `.ignore` or `.gitignore` are excluded.
- Partial failures still produce a sheet; a summary of successful captures is shown at the end.
