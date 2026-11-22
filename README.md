# sketch

<div align="right">
  <a href="README.ja.md">ja</a> | <b>en</b>
</div>

For writing every day.

A CLI tool for managing and editing p5.js sketches.

![sketch](.github/sketch.webp)

## Requirements

- [fzf](https://github.com/junegunn/fzf)
- [fd](https://github.com/sharkdp/fd)
- [ripgrep](https://github.com/BurntSushi/ripgrep)
- [bat](https://github.com/sharkdp/bat)
- [ffmpeg](https://github.com/FFmpeg/FFmpeg)
- Chrome / Chromium or [Chrome for Testing](https://github.com/GoogleChromeLabs/chrome-for-testing) (for recording)
- A browser (for preview)
- An editor (any of the supported editors listed below; default is nvim)

### Supported Editors

- [nvim](https://github.com/neovim/neovim) / [vim](https://github.com/vim/vim) / `vi` / [neovide](https://github.com/neovide/neovide)
- [nano](https://nano-editor.org/) / [micro](https://github.com/zyedidia/micro) / [edit](https://github.com/microsoft/edit)
- [hx / helix](https://github.com/helix-editor/helix)
- [fresh](https://github.com/sinelaw/fresh)

<details>
<summary>Homebrew / Scoop / Winget</summary>

- **macOS (Homebrew)**

  ```bash
  brew install fzf fd ripgrep bat neovim ffmpeg
  ```

- **Windows (Scoop)**
  ```powershell
  scoop install fzf fd ripgrep bat neovim ffmpeg
  ```
- **Windows (Winget)**
  ```powershell
  winget install --id junegunn.fzf --id sharkdp.fd --id BurntSushi.ripgrep.MSVC --id sharkdp.bat --id Neovim.Neovim --id Gyan.FFmpeg
  ```

</details>

## Installation

Download the appropriate zip from [releases](https://github.com/yuruyurau/sketch/releases) and extract it.

### Windows

After extracting, place `sketch.exe` in a directory that is on your PATH.

<details>
<summary>If a warning appears</summary>

### Windows

> **If you see “Windows protected your PC”**
> Because the binary is unsigned, SmartScreen may show a warning on first run.
>
> - **From the dialog**: Click **More info** → **Run anyway**.
> - **From PowerShell**:
>   ```powershell
>   Unblock-File -Path .\sketch.exe
>   ```

</details>

### macOS

After extracting, grant execute permission and place it in a directory on your PATH.  
If you downloaded it with a browser, also remove the quarantine attribute.

```bash
chmod +x ./sketch
xattr -d com.apple.quarantine ./sketch
mv ./sketch /usr/local/bin/
```

### Linux

After extracting, grant execute permission and place it in a directory on your PATH.

```bash
chmod +x sketch
mv sketch /usr/local/bin/
```

## Setup

1. Create `~/.config/sketch/config.json`

   > **The tool will not start if config.json is missing. You must create it on first use.**

2. Configure the required fields (see the example below)

   > **On Windows**  
   > `~/.config/sketch/config.json` corresponds to `%USERPROFILE%\.config\sketch\config.json`.  
   > Inside config.json you can use either `~` or `$USERPROFILE` for the home directory.

### Loading p5.js

Control the source with `p5Source`.

| Value     | Behavior                                               |
| --------- | ------------------------------------------------------ |
| `"cdn"`   | Always use the CDN (version controlled by `p5Version`) |
| `"local"` | Always use the local file (`p5LocalPath`)              |
| `"auto"`  | Prefer CDN, fall back to local if unreachable          |
| omitted   | Same as `"auto"`                                       |

To use a local file:

1. Download from [GitHub Releases](https://github.com/processing/p5.js/releases) or [jsDelivr](https://cdn.jsdelivr.net/npm/p5@2/lib/p5.min.js)
2. Place it at `~/.config/sketch/p5.min.js` (or similar)
3. Set `p5LocalPath`

```json
"p5Source": "auto",
"p5LocalPath": "~/.config/sketch/p5.min.js"
```

## config.json Example

```json
{
  "editor": "nvim",
  "p5Version": "2",
  "p5Source": "auto",
  "p5LocalPath": "~/.config/sketch/p5.min.js",
  "server": {
    "port": 5173,
    "openBrowser": true,
    "browser": {
      "path": "",
      "args": []
    }
  },
  "record": {
    "port": 5174,
    "fps": 60,
    "total": 300,
    "start": 0,
    "crf": 14,
    "codec": "vp9",
    "preset": "slow",
    "cpuUsed": "1",
    "g": 15,
    "pixelDensity": 1.0,
    "browser": {
      "path": "",
      "args": []
    }
  },
  "sketchDir": "~/Documents/sketches",
  "outDir": "~/Documents/sketches/record"
}
```

## config.json Reference

### General

| Key           | Description                                                 | Example / Default               |
| ------------- | ----------------------------------------------------------- | ------------------------------- |
| `editor`      | Editor to use                                               | `nvim` (default)                |
| `p5Version`   | p5.js version on the CDN<br>`"2"` loads `p5@2` (latest 2.x) | `"2"`                           |
| `p5Source`    | Source of p5.js<br>`"cdn"` / `"local"` / `"auto"`           | `"auto"`                        |
| `p5LocalPath` | Path to the local p5.js                                     | `"~/.config/sketch/p5.min.js"`  |
| `sketchDir`   | Directory for sketches                                      | `"~/Documents/sketches"`        |
| `outDir`      | Output directory for recordings                             | `"~/Documents/sketches/record"` |

### server

| Key            | Description                                 | Default |
| -------------- | ------------------------------------------- | ------- |
| `port`         | Server port                                 | `5173`  |
| `openBrowser`  | Open a browser on start                     | `true`  |
| `browser.path` | Browser executable (empty = system default) | `""`    |
| `browser.args` | Extra arguments for the browser             | `[]`    |

### record

| Key            | Description                               | Notes                                          |
| -------------- | ----------------------------------------- | ---------------------------------------------- |
| `port`         | Port for the recording server             | Prefer a different port from the normal server |
| `fps`          | Frame rate                                |                                                |
| `total`        | Total frames to record                    |                                                |
| `start`        | Starting frame                            |                                                |
| `codec`        | `vp9` or `h264`                           | `vp9` → webm, otherwise mp4                    |
| `crf`          | Quality (lower = better)                  |                                                |
| `preset`       | Preset for h264                           |                                                |
| `cpuUsed`      | Speed setting for vp9                     |                                                |
| `g`            | Keyframe interval                         |                                                |
| `pixelDensity` | pixelDensity during recording             |                                                |
| `browser.path` | Chrome binary for recording               | **Required** (launched headless)               |
| `browser.args` | Extra arguments for the recording browser |                                                |

## Architecture

```
sketch
├── Interactive mode (fzf)
│   ├── today / add / file / fuzzy / exact
│   │   └── Select file → start server → open editor
│   ├── record
│   │   └── Select file + options → start recording server → chromedp + ffmpeg
│   └── config
│       └── Open config.json
│
└── Server mode (--server)
    ├── Serve sketch with p5.js wrapper
    ├── Live reload via SSE
    └── File watching with fsnotify
```

## Usage

On launch you select a mode with fzf.

| Mode   | Description                                         |
| ------ | --------------------------------------------------- |
| today  | Open today’s sketch (creates a template if missing) |
| fuzzy  | Fuzzy search in sketch contents (fzf)               |
| exact  | Exact search in sketch contents (fzf --exact)       |
| file   | Search by filename                                  |
| add    | Create a new sketch                                 |
| record | Record a sketch                                     |
| config | Open the config file (config.json)                  |

> **About search**
>
> - `file` uses `fd` to list files.
> - `fuzzy` / `exact` use `ripgrep` to output all lines.
> - Actual filtering is done by `fzf` in both cases (`exact` uses `--exact` mode).
> - If `.gitignore` (or `.ignore`) exists, it is respected for exclusion (effective even outside a Git repository).

### fzf Keybindings

<!--  todo add fzf video -->

Common:

| Key                              | Action         |
| -------------------------------- | -------------- |
| `Ctrl+J` / `Ctrl+K` or `↑` / `↓` | Move selection |
| `Enter`                          | Confirm        |
| `Ctrl+Q` or `Esc`                | Cancel         |

Only during file / content search:

| Key                   | Action         |
| --------------------- | -------------- |
| `Ctrl+T`              | Toggle preview |
| `Shift+J` / `Shift+K` | Scroll preview |

### Mode Details

#### today

| Item       | Description                                                                                             |
| ---------- | ------------------------------------------------------------------------------------------------------- |
| Action     | Opens today’s sketch (`YYYY-MM-DD.js`)                                                                  |
| If missing | Creates a template then opens it                                                                        |
| Editor     | The editor set in config (default: nvim)                                                                |
| Server     | Starts a server (stops when the editor exits). If already in use, asks whether to open the editor only. |

<!-- todo video -->

#### fuzzy

| Item        | Description                                                                                             |
| ----------- | ------------------------------------------------------------------------------------------------------- |
| Action      | Fuzzy-search sketch contents and open the match                                                         |
| Search tool | ripgrep (full output) + fzf (fuzzy search)                                                              |
| Cursor      | Opens at the matching line                                                                              |
| Preview     | `Ctrl+T` to toggle, `Shift+J` / `Shift+K` to scroll                                                     |
| Server      | Starts a server (stops when the editor exits). If already in use, asks whether to open the editor only. |

<!-- todo video -->

#### exact

| Item        | Description                                                                                             |
| ----------- | ------------------------------------------------------------------------------------------------------- |
| Action      | Exact-search sketch contents and open the match                                                         |
| Search tool | ripgrep (full output) + fzf (`--exact`)                                                                 |
| Cursor      | Opens at the matching line                                                                              |
| Preview     | `Ctrl+T` to toggle, `Shift+J` / `Shift+K` to scroll                                                     |
| Server      | Starts a server (stops when the editor exits). If already in use, asks whether to open the editor only. |

<!-- todo video -->

#### file

| Item        | Description                                                                                             |
| ----------- | ------------------------------------------------------------------------------------------------------- |
| Action      | Search by filename and open                                                                             |
| Search tool | fd + fzf                                                                                                |
| Preview     | `Ctrl+T` to toggle, `Shift+J` / `Shift+K` to scroll                                                     |
| Server      | Starts a server (stops when the editor exits). If already in use, asks whether to open the editor only. |

<!-- todo video -->

#### add

| Item      | Description                                                                                             |
| --------- | ------------------------------------------------------------------------------------------------------- |
| Action    | Create (or open) a sketch and open it                                                                   |
| Extension | `.js` is added automatically if omitted                                                                 |
| If exists | Opens the existing file                                                                                 |
| Server    | Starts a server (stops when the editor exits). If already in use, asks whether to open the editor only. |

<!-- todo video -->

#### record

| Item         | Description                                                                                       |
| ------------ | ------------------------------------------------------------------------------------------------- |
| Action       | Select a sketch and record it                                                                     |
| Before start | You can change fps / total / start / codec / crf / pixelDensity / p5Version / p5Source etc.       |
| Output       | `outDir` from config                                                                              |
| Format       | vp9 → webm, otherwise mp4                                                                         |
| Server       | Temporarily starts a server on `record.port` (stops afterward). Does not start if already in use. |

> **About `total` and `start`**
>
> - `total` accepts integer arithmetic expressions (e.g. `60*30`, `1e4`, `1_000`, `0x10`).  
>   The result must be an integer ≥ 1. Non-integer values (such as `1.5`) are rejected.  
>   Division is integer division (`1/2` → `0`).
> - `start` advances the sketch by that many frames before recording begins.

<!-- todo video -->

#### config

| Item   | Description                                        |
| ------ | -------------------------------------------------- |
| Action | Opens `~/.config/sketch/config.json` in the editor |
| Server | Not started                                        |

## Building

> [!NOTE]
> **Go 1.27.4 or later** is required to build.

```bash
git clone https://github.com/yuruyurau/sketch
cd sketch
go build
```

### Dependencies

- `github.com/chromedp/chromedp` v0.24.0 (headless Chrome control for recording)
- `github.com/fsnotify/fsnotify` v1.10.1 (file watching / live reload)

### Using the Nushell scripts

**Nushell** and **7-Zip** are required to run the scripts.

#### Current environment only

```bash
nu build.nu
```

- Outputs to `bin/sketch` (or `bin/sketch.exe` on Windows).

#### Multi-platform release build

```bash
nu release.nu
```

- Cross-compiles and creates zips in `release/` for the following targets:
  - `windows/amd64`
  - `darwin/amd64`, `darwin/arm64`
  - `linux/amd64`, `linux/arm64`
- The version is taken from the latest Git tag (falls back to `0.0.0` if no tag exists).
- Each zip includes `LICENSE` and `config.json`.
