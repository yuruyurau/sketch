# Building from Source

<div align="right">
  <a href="building.ja.md">ja</a> | <b>en</b>
</div>

> [!NOTE]
> **Go 1.27.4 or later** is required to build.

```bash
git clone https://github.com/yuruyurau/sketch
cd sketch
go build
```

## Dependencies

- `github.com/chromedp/chromedp` v0.24.0 (headless Chrome control for recording)
- `github.com/fsnotify/fsnotify` v1.10.1 (file watching / live reload)

## Using the Nushell scripts

**Nushell** and **7-Zip** are required to run the scripts.

### Current environment only

```bash
nu build.nu
```

- Outputs to `bin/sketch` (or `bin/sketch.exe` on Windows).

### Multi-platform release build

```bash
nu release.nu
```

- Cross-compiles and creates zips in `release/` for the following targets:
  - `windows/amd64`
  - `darwin/amd64`, `darwin/arm64`
  - `linux/amd64`, `linux/arm64`
- The version is taken from the latest Git tag (falls back to `0.0.0` if no tag exists).
- Each zip includes `LICENSE` and `config.json`.
