function drawGrid(size = 40) {
  push();
  stroke(60, 196);
  strokeWeight(1);
  for (let x = 0; x < width; x += size) {
    line(x, 0, x, height);
  }
  for (let y = 0; y < height; y += size) {
    line(0, y, width, y);
  }
  pop();
}

const config = {
  size: 40,
};
