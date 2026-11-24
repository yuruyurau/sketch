// Open this file with the sketch tool.
// Make sure sketch.webp is in the same folder.
// Note: Live reload only watches this file. Saving utils.js does not reload — save this file to refresh.

let img;

async function setup() {
  createCanvas(400, 400);

  img = await loadImage("sketch.webp");
  await loadScript("utils.js");
}

function draw() {
  background(0);
  if (img) image(img, 0, 0, width, height);

  drawGrid(40);

  fill(255);
  noStroke();
  circle(mouseX, mouseY, config.size);
}

function loadScript(src) {
  return new Promise((resolve, reject) => {
    const s = document.createElement("script");
    s.src = src + "?t=" + Date.now();
    s.onload = resolve;
    s.onerror = reject;
    document.head.appendChild(s);
  });
}
