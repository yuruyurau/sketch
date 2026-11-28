# 高度な使い方

<div align="right">
  <b>ja</b> | <a href="advanced.md">en</a>
</div>

## 複数ファイルの利用

ローカルサーバーは**開いたスケッチファイルが置かれているディレクトリ**をルートにします。  
そのスケッチからの相対パスでファイルを読み込めます。サブディレクトリは利用可能ですが、`../` はブロックされます（そのディレクトリの外は配信されません）。

```
~/Documents/sketches/          ← sketchDir
├── 2026-11-28.js              ← このファイルを開いた場合のルート
├── photo.png                  ← "./photo.png" で読める
└── project-a/
    ├── main.js                ← このファイルを開いた場合のルート
    └── assets/
        └── image.png          ← "./assets/image.png" で読める
```

- `2026-11-28.js` から `./photo.png` → OK
- `2026-11-28.js` から `./project-a/assets/image.png` → OK
- `project-a/main.js` から `./assets/image.png` → OK
- `project-a/main.js` から `../photo.png` → ブロック

### Global Mode

p5.js 2.x では `preload()` が削除されているため、`async setup()` + `await` を使います。

```js
// main.js（project-a/ 内）
let img;

function loadScript(src) {
  return new Promise((resolve, reject) => {
    const s = document.createElement("script");
    s.src = src + "?t=" + Date.now();
    s.onload = resolve;
    s.onerror = reject;
    document.head.appendChild(s);
  });
}

async function setup() {
  createCanvas(400, 400);

  img = await loadImage("./assets/photo.png");
  await loadScript("./utils.js");
}

function draw() {
  background(0);
  if (img) image(img, 0, 0, width, height);
  drawGrid(40);
}
```

完全な例は [`examples/multi-file/`](../examples/multi-file/) を参照してください。

> ライブリロードは開いたメインファイルだけを監視します。補助ファイルを保存してもリロードは走りません。

### Instance Mode

`.mod.js` ファイルは ES Modules として読み込まれ、instance mode で実行されます。

```js
// main.mod.js
import { drawGrid, config } from "./utils.js";

const sketch = (p) => {
  p.setup = () => p.createCanvas(400, 400);

  p.draw = () => {
    p.background(0);
    drawGrid(p, config.gridSize);
  };
};

// 録画に必要
window.p5Instance = new p5(sketch);
```

- 相対 import（`./utils.js` や `./lib/grid.js`）は可能です。
- 録画時は `window.p5Instance` を設定してください。

完全な例は [`examples/instance-mode/`](../examples/instance-mode/) を参照してください。
