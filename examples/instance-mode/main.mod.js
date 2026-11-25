import { drawGrid, config } from "./utils.js";

const sketch = (p) => {
  p.setup = () => {
    p.createCanvas(400, 400);
  };

  p.draw = () => {
    p.background(0);
    drawGrid(p, config.gridSize);
    p.fill(255);
    p.noStroke();
    p.circle(p.mouseX, p.mouseY, config.circleSize);
  };
};

window.p5Instance = new p5(sketch);
