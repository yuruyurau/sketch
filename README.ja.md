# sketch

<div align="right">
  <b>ja</b> | <a href="README.md">en</a>
</div>

日常的に書くために。

p5.js のスケッチを管理・編集するための CLI ツール。

![sketch](.github/sketch.webp)

## 必要環境

- [fzf](https://github.com/junegunn/fzf)
- [fd](https://github.com/sharkdp/fd)
- [ripgrep](https://github.com/BurntSushi/ripgrep)
- [bat](https://github.com/sharkdp/bat)
- [ffmpeg](https://github.com/FFmpeg/FFmpeg)
- Chrome / Chromium または [Chrome for Testing](https://github.com/GoogleChromeLabs/chrome-for-testing)（録画時）
- ブラウザ（プレビュー用）
- エディタ（後述の対応エディタのいずれか。デフォルトは nvim）

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

### 対応エディタ

- [nvim](https://github.com/neovim/neovim) / [vim](https://github.com/vim/vim) / `vi` / [neovide](https://github.com/neovide/neovide)
- [nano](https://nano-editor.org/) / [micro](https://github.com/zyedidia/micro) / [edit](https://github.com/microsoft/edit)
- [hx / helix](https://github.com/helix-editor/helix)
- [fresh](https://github.com/sinelaw/fresh)
- [zed / zeditor](https://zed.dev)
- [code / code-insiders](https://code.visualstudio.com/)

<details>
<summary>Zed の設定</summary>

`~/.config/sketch/config.json` に以下を設定:

```json
"editor": "zed"
```

（環境によってはバイナリ名が `zeditor` の場合があるので、そのときは `"zeditor"` にしてください）

#### Zed 内から sketch を起動する（推奨）

`tasks.json` と `keymap.json` の場所:

- macOS / Linux: `~/.config/zed/`
- Windows: `%APPDATA%\Zed\`
- プロジェクトローカル（全OS共通）: `.zed/`

**`tasks.json`:**

```json
[
  {
    "label": "Sketch",
    "command": "sketch",
    "reveal": "always",
    "reveal_target": "center",
    "hide": "on_success",
    "use_new_terminal": true
  }
]
```

**`keymap.json` (`space s` で起動する場合):**

```json
[
  {
    "context": "Workspace",
    "bindings": {
      "space s": [
        "task::Spawn",
        { "task_name": "Sketch", "reveal_target": "center" }
      ]
    }
  }
]
```

#### コマンドパレットから開く

1. `Ctrl/Cmd+Shift+P` を押す
2. `task: spawn` と入力して選択
3. 表示されたリストから `Sketch` を選ぶ

Vim モードの場合は `:` を押して `task Sketch` と入力しても起動できます。

> **注意**: すでに Zed が起動している状態だと、`zed --wait` が正しくブロックしない既知の挙動があります（タブを閉じてもプロセスが戻ってこない場合がある）。詳細は [zed-industries/zed#54203](https://github.com/zed-industries/zed/issues/54203) を参照。

</details>

<details>
<summary>VS Code の設定</summary>

`~/.config/sketch/config.json` に以下を設定:

```json
"editor": "code"
```

（Insiders 版を使っている場合は `"code-insiders"`）

#### VS Code 内から sketch を起動する（推奨）

`tasks.json` と `settings.json` の場所:

- macOS / Linux: `~/.config/Code/User/`
- Windows: `%APPDATA%\Code\User\`
- プロジェクトローカル（全OS共通）: `.vscode/`

**推奨する `settings.json` の追加:**

```json
{
  "terminal.integrated.defaultLocation": "editor"
}
```

ターミナルがエディタタブとして開くため、最大化しやすくなります。

**`tasks.json`:**

```json
{
  "version": "2.0.0",
  "tasks": [
    {
      "label": "Sketch",
      "type": "shell",
      "command": "sketch",
      "problemMatcher": [],
      "presentation": {
        "reveal": "always",
        "revealProblems": "onProblem",
        "panel": "new",
        "focus": true,
        "echo": false,
        "showReuseMessage": false,
        "clear": true,
        "close": true
      }
    }
  ]
}
```

**`keybindings.json` (`space s` で起動する場合):**

```json
[
  {
    "key": "space s",
    "command": "workbench.action.tasks.runTask",
    "args": "Sketch",
    "when": "editorTextFocus"
  }
]
```

#### コマンドパレットから開く

1. `Ctrl/Cmd+Shift+P` を押す
2. `Tasks: Run Task` と入力して選択
3. `Sketch` を選ぶ

> **fzf のキー操作について（VS Code ターミナル内）**:  
> VS Code が `Ctrl+J` / `Ctrl+K` / `Ctrl+Q` を横取りするため、fzf まで届きません。  
> 選択の移動は矢印キー（`↑` / `↓`）、キャンセルは `Esc` を使ってください。

</details>

## インストール

[Releases](https://github.com/yuruyurau/sketch/releases) から環境に合った zip をダウンロードし、展開してください。

各 zip にはバイナリ、`LICENSE`、`config.json` が含まれています。  
検証用の `checksums.txt` も付属しています。

### Windows

展開後、`sketch.exe`をパスの通ったディレクトリに配置してください。

<details>
<summary>警告が出る場合の対処</summary>

### Windows

> **「Windows によって PC が保護されました」と表示される場合**
> 署名のない実行ファイルのため、初回実行時に SmartScreen の警告が出ることがあります。
>
> - **画面から解除**: 警告ダイアログの **「詳細情報」** → **「実行」** をクリックします。
> - **コマンドで解除 (PowerShell)**:
>   ```powershell
>   Unblock-File -Path .\sketch.exe
>   ```

</details>

### macOS

展開後、パスの通ったディレクトリに配置してください。  
ブラウザからダウンロードした場合は隔離属性も解除します。

```bash
xattr -d com.apple.quarantine sketch
mv sketch ~/.local/bin/
```

### Linux

展開後、パスの通ったディレクトリに配置してください。

```bash
mv sketch ~/.local/bin/
```

## セットアップ

1. `~/.config/sketch/config.json` を作成する

   > **config.json が存在しない場合は起動しません。初回は必ず作成してください。**

2. 必要な項目を設定する（例は後述）

   > **Windows の場合**  
   > `~/.config/sketch/config.json` は `%USERPROFILE%\.config\sketch\config.json` を指します。  
   > `config.json` 内のパス指定では `~` または `$USERPROFILE` も使えます。

### p5.js の読み込み方法

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

**config.json の例：**

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

## config.json の項目

### 全体

| キー          | 説明                                                           | 例 / デフォルト                 |
| ------------- | -------------------------------------------------------------- | ------------------------------- |
| `editor`      | 使用するエディタ                                               | `nvim`（デフォルト）            |
| `p5Version`   | CDN で使う p5.js のバージョン<br>`"2"` で `p5@2`（2.x 系最新） | `"2"`                           |
| `p5Source`    | p5.js の読み込み元<br>`"cdn"` / `"local"` / `"auto"`           | `"auto"`                        |
| `p5LocalPath` | ローカルの p5.js のパス                                        | `"~/.config/sketch/p5.min.js"`  |
| `sketchDir`   | スケッチを保存するディレクトリ                                 | `"~/Documents/sketches"`        |
| `outDir`      | 録画ファイルの出力先                                           | `"~/Documents/sketches/record"` |

### server

| キー           | 説明                                       | デフォルト |
| -------------- | ------------------------------------------ | ---------- |
| `port`         | サーバーのポート番号                       | `5173`     |
| `openBrowser`  | 起動時にブラウザを開くか                   | `true`     |
| `browser.path` | 使用するブラウザのパス（空ならデフォルト） | `""`       |
| `browser.args` | ブラウザに渡す引数                         | `[]`       |

### record

| キー           | 説明                         | 備考                                               |
| -------------- | ---------------------------- | -------------------------------------------------- |
| `port`         | 録画用サーバーのポート       | 通常のサーバーと別ポート推奨                       |
| `fps`          | フレームレート               |                                                    |
| `total`        | 録画する総フレーム数         |                                                    |
| `start`        | 開始フレーム                 |                                                    |
| `codec`        | `vp9` または `h264`          | `vp9` は webm、それ以外は mp4                      |
| `crf`          | 品質（低いほど高画質）       |                                                    |
| `preset`       | h264 時のプリセット          |                                                    |
| `cpuUsed`      | vp9 時の速度設定             |                                                    |
| `g`            | キーフレーム間隔             |                                                    |
| `pixelDensity` | 録画時の pixelDensity        |                                                    |
| `timeout`      | 録画全体のタイムアウト（分） | `0` または負の値でタイムアウトなし。デフォルト: 60 |
| `maxFail`      | 連続フレーム失敗の許容回数   | デフォルト: 10                                     |
| `browser.path` | 録画に使う Chrome のパス     | **必須**（headless で起動）                        |
| `browser.args` | 録画用ブラウザの引数         |                                                    |

## アーキテクチャ

```
sketch
├── 対話モード（fzf）
│   ├── today / add / file / fuzzy / exact
│   │   └── ファイル選択 → サーバー起動 → エディタ起動
│   ├── record
│   │   └── ファイル選択 + オプション → 録画用サーバー起動 → chromedp + ffmpeg
│   └── config
│       └── config.json を開く
│
└── サーバーモード（--server）
    ├── p5.js ラッパー付きでスケッチを配信
    ├── SSE によるライブリロード
    └── fsnotify によるファイル監視
```

## 使い方

起動すると fzf でモードを選びます。

https://github.com/user-attachments/assets/37e20539-2a49-4cc4-a0ff-eb6a9a268570

| モード | 説明                                                                                                      |
| ------ | --------------------------------------------------------------------------------------------------------- |
| today  | 今日のスケッチを開く。存在しない場合は global / instance を選んで作成。両方ある場合はどちらを開くか選択。 |
| fuzzy  | 内容のあいまい検索（fzf）                                                                                 |
| exact  | 内容の完全一致検索（fzf --exact）                                                                         |
| file   | ファイル名で検索                                                                                          |
| add    | 新しいスケッチを作成。名前入力後に global / instance を選択（拡張子を明示した場合はスキップ）。           |
| record | 録画する                                                                                                  |
| config | 設定ファイル（config.json）を開く                                                                         |

> **検索について**
>
> - `file` では `fd` でファイル一覧を取得します。
> - `fuzzy` / `exact` では `ripgrep` で全行を取得します。
> - 実際の絞り込みはいずれも `fzf` が行います（`exact` は `--exact` モード）。
> - `.gitignore`（または `.ignore`）が存在する場合はそれを尊重して除外します（`.git` 管理外のディレクトリでも有効）。

### fzf の操作

<!--  todo fzf の動画を追加 -->

共通：

| キー                                 | 動作       |
| ------------------------------------ | ---------- |
| `Ctrl+J` / `Ctrl+K` または `↑` / `↓` | 選択を移動 |
| `Enter`                              | 決定       |
| `Ctrl+Q` または `Esc`                | キャンセル |

ファイル検索・内容検索時のみ：

| キー                  | 動作                    |
| --------------------- | ----------------------- |
| `Ctrl+T`              | プレビューの表示/非表示 |
| `Shift+J` / `Shift+K` | プレビューのスクロール  |

### 各モードの操作

#### today

| 項目               | 内容                                                                      |
| ------------------ | ------------------------------------------------------------------------- |
| 動作               | 今日の日付（`YYYY-MM-DD.js` または `YYYY-MM-DD.mod.js`）のスケッチを開く  |
| ファイルがない場合 | global モード（`.js`）か instance モード（`.mod.js`）を選んで作成して開く |
| 両方ある場合       | どちらを開くか選択する                                                    |
| 開くエディタ       | config の `editor`（デフォルト: nvim）                                    |
| サーバー           | 起動する（エディタ終了後に停止）。使用中なら確認してエディタのみ開く      |

<!-- todo 動画 -->

#### fuzzy

| 項目         | 内容                                                                 |
| ------------ | -------------------------------------------------------------------- |
| 動作         | スケッチの内容をあいまい検索して開く                                 |
| 検索ツール   | ripgrep（全行出力）+ fzf（あいまい検索）                             |
| カーソル位置 | マッチした行にカーソルを置いてエディタが開く                         |
| プレビュー   | `Ctrl+T` で表示/非表示、`Shift+J` / `Shift+K` でスクロール           |
| サーバー     | 起動する（エディタ終了後に停止）。使用中なら確認してエディタのみ開く |

<!-- todo 動画 -->

#### exact

| 項目         | 内容                                                                 |
| ------------ | -------------------------------------------------------------------- |
| 動作         | スケッチの内容を完全一致検索して開く                                 |
| 検索ツール   | ripgrep（全行出力）+ fzf（`--exact`）                                |
| カーソル位置 | マッチした行にカーソルを置いてエディタが開く                         |
| プレビュー   | `Ctrl+T` で表示/非表示、`Shift+J` / `Shift+K` でスクロール           |
| サーバー     | 起動する（エディタ終了後に停止）。使用中なら確認してエディタのみ開く |

<!-- todo 動画 -->

#### file

| 項目       | 内容                                                                 |
| ---------- | -------------------------------------------------------------------- |
| 動作       | ファイル名で検索して開く                                             |
| 検索ツール | fd + fzf                                                             |
| プレビュー | `Ctrl+T` で表示/非表示、`Shift+J` / `Shift+K` でスクロール           |
| サーバー   | 起動する（エディタ終了後に停止）。使用中なら確認してエディタのみ開く |

<!-- todo 動画 -->

#### add

| 項目                     | 内容                                                                            |
| ------------------------ | ------------------------------------------------------------------------------- |
| 動作                     | 新しいスケッチを作成して開く                                                    |
| 拡張子                   | `.js` または `.mod.js` を付けた場合はそのモードを使う。付けない場合は選択する。 |
| ファイルがすでにある場合 | そのまま開く                                                                    |
| サーバー                 | 起動する（エディタ終了後に停止）。使用中なら確認してエディタのみ開く            |

<!-- todo 動画 -->

#### record

| 項目     | 内容                                                                                                       |
| -------- | ---------------------------------------------------------------------------------------------------------- |
| 動作     | スケッチを選んで録画する                                                                                   |
| 開始前   | fps / total / start / codec / crf / pixelDensity / timeout / maxFail / p5Version / p5Source などを変更可能 |
| 出力先   | config の `outDir`                                                                                         |
| 出力形式 | vp9 → webm、それ以外 → mp4                                                                                 |
| サーバー | `record.port` で一時起動（終了後に停止）。使用中なら起動しない                                             |

> **`total` と `start` について**
>
> - `total` は整数の計算式を入力できます（例: `60*30`, `1e4`, `1_000`, `0x10`）。  
>   結果は 1 以上の整数である必要があります。整数にならない値（`1.5` など）はエラーになります。  
>   除算は整数除算です（`1/2` → `0`）。
> - `start` は録画開始前にそのフレーム数だけスケッチを進めます。

> **`timeout` と `maxFail` について**
>
> - `timeout`: 録画全体のタイムアウト（分）。`0` または負の値でタイムアウトなし（デフォルト: 60）。
> - `maxFail`: 連続してフレーム取得に失敗したときの許容回数（デフォルト: 10）。

<!-- todo 動画 -->

#### config

| 項目     | 内容                                            |
| -------- | ----------------------------------------------- |
| 動作     | `~/.config/sketch/config.json` をエディタで開く |
| サーバー | 起動しない                                      |

## 複数ファイルの利用

### Global Mode

スケッチと同じディレクトリに置いたファイル（画像・補助スクリプト・JSON など）は、相対パスで読み込めます。

完全な動作例はこちらを参照してください：

- [`examples/multi-file/`](examples/multi-file/)

> p5.js 2.x では `preload()` が削除されています。`async setup()` + `await` を使ってください。  
> ライブリロードは開いたメインファイルだけを監視します。`utils.js` などを保存してもリロードは走りません。リロードしたいときは `main.js` を保存してください。

### Instance Mode

`.mod.js` ファイルは ES Modules として読み込まれ、**instance mode** で使います。

| 拡張子     | モード        | 備考                                |
| ---------- | ------------- | ----------------------------------- |
| `*.js`     | global mode   | 従来のグローバルモード              |
| `*.mod.js` | instance mode | `import` が使えるインスタンスモード |

```js
// 2026-11-25.mod.js
import { drawGrid } from "./utils.js";

const sketch = (p) => {
  p.setup = () => p.createCanvas(400, 400);

  p.draw = () => {
    p.background(0);
    drawGrid(p);
  };
};

// 録画に必要
window.p5Instance = new p5(sketch);
```

- 相対 import（`./utils.js` や `./lib/grid.js`）は可能です。
- `../` は使えません。
- ユーティリティ関数は第1引数で `p` を受け取るように書いてください。
- 録画時は `window.p5Instance` を設定してください。

完全な例は [`examples/instance-mode/`](examples/instance-mode/) を参照してください。

## ビルドする場合

> [!NOTE]
> ビルドには **Go 1.27.4 以上** が必要です。

```bash
git clone https://github.com/yuruyurau/sketch
cd sketch
go build
```

### 依存ライブラリ

- `github.com/chromedp/chromedp` v0.24.0（録画時のヘッドレス Chrome 操作）
- `github.com/fsnotify/fsnotify` v1.10.1（ファイル変更監視・自動リロード）

### Nushell スクリプトを使う場合

スクリプトの実行には **Nushell** と **7-Zip** が必要です。

#### 現在の環境用

```bash
nu build.nu
```

- `bin/sketch`（Windows は `bin/sketch.exe`）に出力します。

#### 複数環境向けリリースビルド

```bash
nu release.nu
```

- 以下のターゲット向けにクロスコンパイルし、zip を `release/` に出力します：
  - `windows/amd64`
  - `darwin/amd64`, `darwin/arm64`
  - `linux/amd64`, `linux/arm64`
- バージョンは最新の Git タグから取得します（タグがない場合は `0.0.0`）。
- 各 zip に `LICENSE` と `config.json` を同梱します。
