# モード詳細

<div align="right">
  <b>ja</b> | <a href="modes.md">en</a>
</div>

起動後、fzf でモードを選択します。

| Mode                         | 説明                                         |
| ---------------------------- | -------------------------------------------- |
| [today](#today)              | 今日のスケッチを開く（なければ作成）         |
| [fuzzy](#fuzzy--exact--file) | スケッチ内容をあいまい検索                   |
| [exact](#fuzzy--exact--file) | スケッチ内容を完全一致検索                   |
| [file](#fuzzy--exact--file)  | ファイル名で検索                             |
| [add](#add)                  | 新しいスケッチを作成                         |
| [record](#record)            | スケッチを録画                               |
| [build](#build)              | 静的 HTML にバンドル（Bun が必要）           |
| [sheet](#sheet)              | コンタクトシートを作成（ImageMagick が必要） |
| [config](#config)            | config.json を開く                           |

## 共通の動作

スケッチを開くモードの多くは以下を共通で行います：

- **エディタ**: `config.json` で設定したもの（デフォルト: `nvim`）
- **サーバー**: ローカルサーバーを起動（エディタ終了後に停止）。  
  ポートが使用中の場合は「エディタのみ開くか」確認する
- **プレビュー**（検索モード）: `Ctrl+T` で表示/非表示、`Shift+J` / `Shift+K` でスクロール

## today

今日のスケッチ（`YYYY-MM-DD.js` または `YYYY-MM-DD.mod.js`）を開く。

- ファイルがない場合: global モード（`.js`）か instance モード（`.mod.js`）を選んで作成
- 両方ある場合: どちらを開くか選択

## fuzzy / exact / file

| モード | 検索ツール              | 備考               |
| ------ | ----------------------- | ------------------ |
| fuzzy  | ripgrep + fzf           | 内容をあいまい検索 |
| exact  | ripgrep + fzf `--exact` | 内容を完全一致検索 |
| file   | fd + fzf                | ファイル名で検索   |

- 内容検索ではマッチした行にカーソルを置いて開く

## add

新しいスケッチを作成（または既存を開く）。

- 名前が `.js` または `.mod.js` で終わる場合はそのモードを使う
- そうでなければ global / instance を選択
- すでに存在する場合はそのまま開く

## record

スケッチを選んで録画する。

- 開始前に fps / total / start / codec / crf / pixelDensity / timeout / maxFail / p5Version / p5Source などを変更可能
- 出力先は config の `outDir`
- 形式: `vp9` → webm、それ以外 → mp4
- `record.port` で一時的にサーバーを起動（使用中なら起動しない）

> **`total` と `start` について**
>
> - `total` は整数の計算式を入力可能（例: `60*30`, `1e4`, `1_000`, `0x10`）。  
>   結果は 1 以上の整数である必要がある。除算は整数除算。
> - `start` は録画開始前にそのフレーム数だけスケッチを進める

> **`timeout` と `maxFail` について**
>
> - `timeout`: 全体のタイムアウト（分）。`0` または負で無効（デフォルト: 60）
> - `maxFail`: 連続フレーム取得失敗の許容回数（デフォルト: 10）

## build

スケッチを静的 HTML にバンドルする。

- PATH に [Bun](https://bun.sh) が必要
- 出力: `index.html` + `bundle.js`（`--single` の場合は単一の `index.html`）
- config の `p5Source` / `p5LocalPath` に従う
- 通常の `.js` → IIFE、`.mod.js` → ESM（相対 import は自動解決）

**CLI 例**

```bash
sketch build path/to/sketch.js
sketch build --single path/to/sketch.js
sketch build --outdir ./public path/to/sketch.js
```

| オプション | 説明                           | デフォルト |
| ---------- | ------------------------------ | ---------- |
| `--outdir` | 出力ディレクトリ               | `dist`     |
| `--single` | 単一の自己完結 HTML として出力 | false      |

対話モードでのデフォルト出力先は `sketchDir/dist`。

## sheet

指定フレームをキャプチャしてコンタクトシートを生成する。

- PATH に `magick`（ImageMagick）が必要
- 録画と同じ Chrome 設定を使用
- 対象: `all` または `one`
- 出力先: `sketchDir/sheet/sheet-YYYYMMDD_HHMMSS.png`
- オプション: frames / pixelDensity / geometry / columns / label / background / fill

**frames の記法**

- `30` → フレーム 30
- `1 5` / `1,5` → 複数指定
- `0..10` → 0 以上 10 未満
- `0..10..2` → ステップ指定
- `0..100/5` → 範囲を 5 等分

その他:

- `columns` が空の場合 → 自動で正方形に近いレイアウト
- `geometry` は ImageMagick の `-geometry` 形式
- `.ignore` / `.gitignore` に書かれたスケッチは除外
- 一部失敗しても可能な限りシートを作成（最後に成功枚数を表示）

## config

`~/.config/sketch/config.json` をエディタで開く。
