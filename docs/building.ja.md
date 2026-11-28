# ソースからビルド

<div align="right">
  <b>ja</b> | <a href="building.md">en</a>
</div>

> [!NOTE]
> ビルドには **Go 1.27.4 以上** が必要です。

```bash
git clone https://github.com/yuruyurau/sketch
cd sketch
go build
```

## 依存ライブラリ

- `github.com/chromedp/chromedp` v0.24.0（録画時のヘッドレス Chrome 操作）
- `github.com/fsnotify/fsnotify` v1.10.1（ファイル変更監視・自動リロード）

## Nushell スクリプトを使う場合

スクリプトの実行には **Nushell** と **7-Zip** が必要です。

### 現在の環境用

```bash
nu build.nu
```

- `bin/sketch`（Windows は `bin/sketch.exe`）に出力します。

### 複数環境向けリリースビルド

```bash
nu release.nu
```

- 以下のターゲット向けにクロスコンパイルし、zip を `release/` に出力します：
  - `windows/amd64`
  - `darwin/amd64`, `darwin/arm64`
  - `linux/amd64`, `linux/arm64`
- バージョンは最新の Git タグから取得します（タグがない場合は `0.0.0`）。
- 各 zip に `LICENSE` と `config.json` を同梱します。
