# Mode Details

<div align="right">
  <a href="modes.ja.md">ja</a> | <b>en</b>
</div>

On launch, select a mode with fzf.

| Mode                         | Description                                      |
| ---------------------------- | ------------------------------------------------ |
| [today](#today)              | Open today’s sketch. Create if it doesn’t exist. |
| [fuzzy](#fuzzy--exact--file) | Fuzzy search in sketch contents                  |
| [exact](#fuzzy--exact--file) | Exact search in sketch contents                  |
| [file](#fuzzy--exact--file)  | Search by filename                               |
| [add](#add)                  | Create a new sketch                              |
| [record](#record)            | Record a sketch                                  |
| [build](#build)              | Bundle into static HTML (requires Bun)           |
| [sheet](#sheet)              | Create a contact sheet (requires ImageMagick)    |
| [config](#config)            | Open config.json                                 |

## Common Behavior

Most modes that open a sketch share the following:

- **Editor**: The one set in `config.json` (default: `nvim`)
- **Server**: Starts a local server (stops when the editor exits).  
  If the port is already in use, you are asked whether to open the editor only.
- **Preview** (search modes): `Ctrl+T` to toggle, `Shift+J` / `Shift+K` to scroll

## today

Opens today’s sketch (`YYYY-MM-DD.js` or `YYYY-MM-DD.mod.js`).

- If missing: Asks whether to create it in global mode (`.js`) or instance mode (`.mod.js`)
- If both exist: Asks which one to open

## fuzzy / exact / file

| Mode  | Search tool             | Notes                    |
| ----- | ----------------------- | ------------------------ |
| fuzzy | ripgrep + fzf           | Fuzzy search on contents |
| exact | ripgrep + fzf `--exact` | Exact search on contents |
| file  | fd + fzf                | Search by filename       |

- Cursor opens at the matching line (content search modes)

## add

Creates (or opens) a sketch.

- If the name ends with `.js` or `.mod.js`, that mode is used
- Otherwise you are asked to choose global or instance mode
- If the file already exists, it is opened

## record

Selects a sketch and records it.

- Before start you can change fps / total / start / codec / crf / pixelDensity / timeout / maxFail / p5Version / p5Source etc.
- Output goes to `outDir` from config
- Format: `vp9` → webm, otherwise mp4
- Temporarily starts a server on `record.port` (does not start if already in use)

> **About `total` and `start`**
>
> - `total` accepts integer arithmetic expressions (e.g. `60*30`, `1e4`, `1_000`, `0x10`).  
>   The result must be an integer ≥ 1. Division is integer division.
> - `start` advances the sketch by that many frames before recording begins.

> **About `timeout` and `maxFail`**
>
> - `timeout`: overall timeout in minutes (`0` or negative = disabled, default: 60)
> - `maxFail`: max consecutive frame failures before abort (default: 10)

## build

Bundles a sketch into static HTML.

- Requires [Bun](https://bun.sh) in PATH
- Output: `index.html` + `bundle.js` (or single `index.html` with `--single`)
- Follows `p5Source` / `p5LocalPath` in config
- Normal `.js` → IIFE, `.mod.js` → ESM (relative imports resolved automatically)

**CLI examples**

```bash
sketch build path/to/sketch.js
sketch build --single path/to/sketch.js
sketch build --outdir ./public path/to/sketch.js
```

| Option     | Description                     | Default |
| ---------- | ------------------------------- | ------- |
| `--outdir` | Output directory                | `dist`  |
| `--single` | Single self-contained HTML file | false   |

In interactive mode the default output directory is `sketchDir/dist`.

## sheet

Captures selected frames and generates a contact sheet.

- Requires `magick` (ImageMagick) in PATH
- Uses the same Chrome settings as recording
- Target: `all` or `one`
- Output: `sketchDir/sheet/sheet-YYYYMMDD_HHMMSS.png`
- Options: frames / pixelDensity / geometry / columns / label / background / fill

**frames syntax**

- `30` → frame 30
- `1 5` / `1,5` → multiple frames
- `0..10` → 0 ≤ n < 10
- `0..10..2` → with step
- `0..100/5` → divide the range into 5 equal parts

Other notes:

- Empty `columns` → automatic near-square grid
- `geometry` uses ImageMagick’s `-geometry` format
- Sketches listed in `.ignore` / `.gitignore` are excluded
- Partial failures still produce a sheet (summary shown at the end)

## config

Opens `~/.config/sketch/config.json` in the editor.
