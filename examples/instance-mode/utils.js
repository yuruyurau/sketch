export const config = {
  gridSize: 40,
  circleSize: 50,
};

export function drawGrid(p, size = 40) {
  p.push();
  p.stroke(255, 50);
  p.strokeWeight(1);

  for (let x = 0; x <= p.width; x += size) {
    p.line(x, 0, x, p.height);
  }
  for (let y = 0; y <= p.height; y += size) {
    p.line(0, y, p.width, y);
  }

  p.pop();
}
