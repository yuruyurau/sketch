# Editor Setup

<div align="right">
  <a href="editors.ja.md">ja</a> | <b>en</b>
</div>

## Supported Editors

| Editor                           |
| -------------------------------- |
| nvim / vim / vi / neovide        |
| nano / micro / edit              |
| hx / helix                       |
| fresh                            |
| [zed / zeditor](#zed)            |
| [code / code-insiders](#vs-code) |

Set the editor in `~/.config/sketch/config.json`:

```json
"editor": "nvim"
```

## Automatic editor selection (Zed / VS Code)

When `sketch` is launched from the integrated terminal of Zed or VS Code, it automatically uses the corresponding editor.

- Inside Zed terminal → `zed`
- Inside VS Code terminal → `code`

In a normal terminal, the value of `"editor"` in `config.json` is used (default: `nvim`).

This detection uses environment variables that the terminals set automatically (`ZED_TERM` / `TERM_PROGRAM`), so no extra configuration is required.

**Launching via tasks is recommended.**

## Zed

Set in `config.json`:

```json
"editor": "zed"
```

(Use `"zeditor"` if that is the binary name on your system.)

### Launching sketch from inside Zed (recommended)

Locations for `tasks.json` and `keymap.json`:

- macOS / Linux: `~/.config/zed/`
- Windows: `%APPDATA%\Zed\`
- Project-local (all platforms): `.zed/`

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

**`keymap.json` (to launch with `space s`):**

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

### Opening via Command Palette

1. Press `Ctrl/Cmd+Shift+P`
2. Type `task: spawn` and select it
3. Choose `Sketch` from the list

In Vim mode, you can also press `:` and type `task Sketch`.

> **Note**: `zed --wait` currently does not always block correctly when Zed is already running (single-instance behavior). Closing the file/tab may not return control to `sketch` until the whole Zed process exits. See [zed-industries/zed#54203](https://github.com/zed-industries/zed/issues/54203).

## VS Code

Set in `config.json`:

```json
"editor": "code"
```

(Use `"code-insiders"` if you use the Insiders build.)

### Launching sketch from inside VS Code (recommended)

Locations for `tasks.json` and `settings.json`:

- macOS / Linux: `~/.config/Code/User/`
- Windows: `%APPDATA%\Code\User\`
- Project-local (all platforms): `.vscode/`

**Recommended `settings.json` addition:**

```json
{
  "terminal.integrated.defaultLocation": "editor"
}
```

This makes the terminal open as an editor tab, which is easier to maximize.

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

**`keybindings.json` (to launch with `space s`):**

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

### Opening via Command Palette

1. Press `Ctrl/Cmd+Shift+P`
2. Type `Tasks: Run Task`
3. Select `Sketch`

> **Note on fzf keybindings inside VS Code terminal**:  
> VS Code intercepts `Ctrl+J` / `Ctrl+K` / `Ctrl+Q`, so they do not reach fzf.  
> Use the arrow keys (`↑` / `↓`) to move the selection and `Esc` to cancel instead.
