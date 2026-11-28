# Advanced Usage

<div align="right">
  <a href="advanced.ja.md">ja</a> | <b>en</b>
</div>

## Using Multiple Files

The local server roots itself at the directory containing the sketch you opened.  
You can load files using relative paths from that sketch. Subdirectories are allowed, but `../` is blocked (the server does not serve outside that directory).

```
~/Documents/sketches/          ← sketchDir
├── 2026-11-28.js              ← root when this file is opened
├── photo.png                  ← "./photo.png" works
└── project-a/
    ├── main.js                ← root when this file is opened
    └── assets/
        └── image.png          ← "./assets/image.png" works
```

- From `2026-11-28.js`: `./photo.png` → OK
- From `2026-11-28.js`: `./project-a/assets/image.png` → OK
- From `project-a/main.js`: `./assets/image.png` → OK
- From `project-a/main.js`: `../photo.png` → blocked

### Global Mode

In p5.js 2.x, `preload()` has been removed. Use `async setup()` + `await`.

```js
// main.js (inside project-a/)
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

See the complete example in [`examples/multi-file/`](../examples/multi-file/).

> Live reload only watches the main file you opened. Saving auxiliary files does not trigger a reload.

### Instance Mode

`.mod.js` files are loaded as ES modules and run in instance mode.

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

// Required for recording
window.p5Instance = new p5(sketch);
```

- Relative imports (`./utils.js`, `./lib/grid.js`) work.
- Set `window.p5Instance` for recording.

See the complete example in [`examples/instance-mode/`](../examples/instance-mode/).
