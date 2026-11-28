# 設定

<div align="right">
  <b>ja</b> | <a href="configuration.md">en</a>
</div>

`~/.config/sketch/config.json` を作成してください。

> **config.json が存在しない場合は起動しません。初回は必ず作成してください。**

Windows の場合、`~/.config/sketch/config.json` は `%USERPROFILE%\.config\sketch\config.json` を指します。  
`config.json` 内のパス指定では `~` または `$USERPROFILE` も使えます。

| 項目                                 | 内容                                                             |
| ------------------------------------ | ---------------------------------------------------------------- |
| [基本設定](#基本設定)                | editor / p5Version / p5Source / p5LocalPath / sketchDir / outDir |
| [サーバー設定](#サーバー設定-server) | port / openBrowser / browser                                     |
| [録画設定](#録画設定-record)         | port / fps / total / codec など                                  |

## p5.js の読み込み方法

`p5Source` で読み込み元を指定できます。

| 値        | 動作                                                       |
| --------- | ---------------------------------------------------------- |
| `"cdn"`   | 常に CDN を使う（`p5Version` でバージョン指定）            |
| `"local"` | 常にローカルファイルを使う（`p5LocalPath`）                |
| `"auto"`  | CDN を優先し、接続できない場合のみローカルにフォールバック |
| 未指定    | `"auto"` と同じ                                            |

ローカルファイルを使う場合：

1. [GitHub Releases](https://github.com/processing/p5.js/releases) または [jsDelivr](https://cdn.jsdelivr.net/npm/p5@2/lib/p5.min.js) からダウンロード
2. `~/.config/sketch/p5.min.js` などに配置
3. `p5LocalPath` にパスを指定

```json
"p5Source": "auto",
"p5LocalPath": "~/.config/sketch/p5.min.js"
```

## 例

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

## リファレンス

### 基本設定

| キー          | 説明                                                           | 例 / デフォルト                 |
| ------------- | -------------------------------------------------------------- | ------------------------------- |
| `editor`      | 使用するエディタ                                               | `nvim`（デフォルト）            |
| `p5Version`   | CDN で使う p5.js のバージョン<br>`"2"` で `p5@2`（2.x 系最新） | `"2"`                           |
| `p5Source`    | p5.js の読み込み元<br>`"cdn"` / `"local"` / `"auto"`           | `"auto"`                        |
| `p5LocalPath` | ローカルの p5.js のパス                                        | `"~/.config/sketch/p5.min.js"`  |
| `sketchDir`   | スケッチを保存するディレクトリ                                 | `"~/Documents/sketches"`        |
| `outDir`      | 録画ファイルの出力先                                           | `"~/Documents/sketches/record"` |

### サーバー設定 (`server`)

| キー           | 説明                                       | デフォルト |
| -------------- | ------------------------------------------ | ---------- |
| `port`         | サーバーのポート番号                       | `5173`     |
| `openBrowser`  | 起動時にブラウザを開くか                   | `true`     |
| `browser.path` | 使用するブラウザのパス（空ならデフォルト） | `""`       |
| `browser.args` | ブラウザに渡す引数                         | `[]`       |

> **補足: バックグラウンド時のスロットリング抑制（Chrome / Chromium）**
>
> プレビュー用ブラウザがフォーカスを失うと（エディタを操作している間など）、Chrome/Chromium はタイマーや `requestAnimationFrame` を大幅にスロットリングします。  
> アニメーションをフルスピードで維持したい場合は、`server.browser.args` に以下を追加してください。
>
> ```json
> "args": [
>   "--disable-renderer-backgrounding",
>   "--disable-background-timer-throttling",
>   "--disable-backgrounding-occluded-windows"
> ]
> ```

### 録画設定 (`record`)

| キー           | 説明                       | 備考                               |
| -------------- | -------------------------- | ---------------------------------- |
| `port`         | 録画用サーバーのポート     | 通常のサーバーと別ポート推奨       |
| `fps`          | フレームレート             |                                    |
| `total`        | 録画する総フレーム数       |                                    |
| `start`        | 開始フレーム               |                                    |
| `codec`        | `vp9` または `h264`        | `vp9` は webm、それ以外は mp4      |
| `crf`          | 品質（小さいほど高画質）   |                                    |
| `preset`       | h264 用プリセット          |                                    |
| `cpuUsed`      | vp9 の速度設定             |                                    |
| `g`            | キーフレーム間隔           |                                    |
| `pixelDensity` | 録画時の pixelDensity      |                                    |
| `timeout`      | 全体のタイムアウト（分）   | `0` または負で無効。デフォルト: 60 |
| `maxFail`      | 連続フレーム取得失敗の上限 | デフォルト: 10                     |
| `browser.path` | 録画用 Chrome のパス       | **必須**（ヘッドレスで起動）       |
| `browser.args` | 録画用ブラウザに渡す引数   |                                    |
