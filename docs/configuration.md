# Configuration

<div align="right">
  <a href="configuration.ja.md">ja</a> | <b>en</b>
</div>

Create `~/.config/sketch/config.json`.

> The tool will not start if `config.json` is missing. You must create it on first use.

On Windows, `~/.config/sketch/config.json` corresponds to `%USERPROFILE%\.config\sketch\config.json`.  
Inside the file you can use either `~` or `$USERPROFILE` for the home directory.

| Section                                          | Contents                                                         |
| ------------------------------------------------ | ---------------------------------------------------------------- |
| [Basic Settings](#basic-settings)                | editor / p5Version / p5Source / p5LocalPath / sketchDir / outDir |
| [Server Settings](#server-settings-server)       | port / openBrowser / browser                                     |
| [Recording Settings](#recording-settings-record) | port / fps / total / codec etc.                                  |

## Loading p5.js

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

## Example

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
    "timeout": 60,
    "maxFail": 10,
    "browser": {
      "path": "",
      "args": []
    }
  },
  "sketchDir": "~/Documents/sketches",
  "outDir": "~/Documents/sketches/record"
}
```

## Reference

### Basic Settings

| Key           | Description                                                 | Example / Default               |
| ------------- | ----------------------------------------------------------- | ------------------------------- |
| `editor`      | Editor to use                                               | `nvim` (default)                |
| `p5Version`   | p5.js version on the CDN<br>`"2"` loads `p5@2` (latest 2.x) | `"2"`                           |
| `p5Source`    | Source of p5.js<br>`"cdn"` / `"local"` / `"auto"`           | `"auto"`                        |
| `p5LocalPath` | Path to the local p5.js                                     | `"~/.config/sketch/p5.min.js"`  |
| `sketchDir`   | Directory for sketches                                      | `"~/Documents/sketches"`        |
| `outDir`      | Output directory for recordings                             | `"~/Documents/sketches/record"` |

### Server Settings (`server`)

| Key            | Description                                 | Default |
| -------------- | ------------------------------------------- | ------- |
| `port`         | Server port                                 | `5173`  |
| `openBrowser`  | Open a browser on start                     | `true`  |
| `browser.path` | Browser executable (empty = system default) | `""`    |
| `browser.args` | Extra arguments for the browser             | `[]`    |

> **Note: Preventing background throttling (Chrome / Chromium)**
>
> When the preview browser loses focus (e.g. while editing in another window), Chrome/Chromium may throttle timers and `requestAnimationFrame`.  
> To keep animations running at full speed, add the following to `server.browser.args`:
>
> ```json
> "args": [
>   "--disable-renderer-backgrounding",
>   "--disable-background-timer-throttling",
>   "--disable-backgrounding-occluded-windows"
> ]
> ```

### Recording Settings (`record`)

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
| `timeout`      | Overall timeout in minutes                | `0` or negative = no timeout. Default: 60      |
| `maxFail`      | Max consecutive frame failures            | Default: 10                                    |
| `browser.path` | Chrome binary for recording               | **Required** (launched headless)               |
| `browser.args` | Extra arguments for the recording browser |                                                |
