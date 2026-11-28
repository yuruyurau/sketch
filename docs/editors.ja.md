# エディタ設定

<div align="right">
  <b>ja</b> | <a href="editors.md">en</a>
</div>

## 対応エディタ

| エディタ                         |
| -------------------------------- |
| nvim / vim / vi / neovide        |
| nano / micro / edit              |
| hx / helix                       |
| fresh                            |
| [zed / zeditor](#zed)            |
| [code / code-insiders](#vs-code) |

`~/.config/sketch/config.json` で指定します：

```json
"editor": "nvim"
```

## エディタの自動選択（Zed / VS Code）

Zed または VS Code の統合ターミナルから `sketch` を起動した場合、自動的にそのエディタが使われます。

- Zed のターミナル内 → `zed`
- VS Code のターミナル内 → `code`

通常のターミナルでは `config.json` の `"editor"` の値が使われます（デフォルトは `nvim`）。

この判定はターミナルが自動でセットする環境変数（`ZED_TERM` / `TERM_PROGRAM`）を利用しているため、追加の設定は不要です。

**tasks 経由での起動を推奨します。**

## Zed

`config.json` に以下を設定：

```json
"editor": "zed"
```

（環境によってはバイナリ名が `zeditor` の場合があるので、そのときは `"zeditor"` にしてください）

### Zed 内から sketch を起動する（推奨）

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
    "cwd": "${ZED_WORKTREE_ROOT:~/Documents}",
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

### コマンドパレットから開く

1. `Ctrl/Cmd+Shift+P` を押す
2. `task: spawn` と入力して選択
3. 表示されたリストから `Sketch` を選ぶ

Vim モードの場合は `:` を押して `task Sketch` と入力しても起動できます。

> **注意**: すでに Zed が起動している状態だと、`zed --wait` が正しくブロックしない既知の挙動があります（タブを閉じてもプロセスが戻ってこない場合がある）。詳細は [zed-industries/zed#54203](https://github.com/zed-industries/zed/issues/54203) を参照。

## VS Code

`config.json` に以下を設定：

```json
"editor": "code"
```

（Insiders 版を使っている場合は `"code-insiders"`）

### VS Code 内から sketch を起動する（推奨）

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
      "options": {
        "cwd": "${workspaceFolder:~/Documents}"
      },
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

### コマンドパレットから開く

1. `Ctrl/Cmd+Shift+P` を押す
2. `Tasks: Run Task` と入力して選択
3. `Sketch` を選ぶ

> **fzf のキー操作について（VS Code ターミナル内）**:  
> VS Code が `Ctrl+J` / `Ctrl+K` / `Ctrl+Q` を横取りするため、fzf まで届きません。  
> 選択の移動は矢印キー（`↑` / `↓`）、キャンセルは `Esc` を使ってください。
