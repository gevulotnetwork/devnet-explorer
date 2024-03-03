(function initRAF() {
  const vendors = ["webkit", "moz"];
  for (const vendor of vendors) {
    if (window.requestAnimationFrame) break;
    window.requestAnimationFrame = window[`${vendor}RequestAnimationFrame`];
    window.cancelAnimationFrame =
      window[`${vendor}CancelAnimationFrame`] ||
      window[`${vendor}CancelRequestAnimationFrame`];
  }

  if (!window.requestAnimationFrame) {
    let lastTime = 0;
    window.requestAnimationFrame = function (callback) {
      const currTime = new Date().getTime();
      const timeToCall = Math.max(0, 16 - (currTime - lastTime));
      const id = setTimeout(() => callback(currTime + timeToCall), timeToCall);
      lastTime = currTime + timeToCall;
      return id;
    };
  }

  if (!window.cancelAnimationFrame) {
    window.cancelAnimationFrame = (id) => clearTimeout(id);
  }
})();

class Matrix {
  constructor(canvas) {
    this.canvas = canvas;
    this.updateDimensions();
    this.ctx = canvas.getContext("2d");

    this.ctx.font = "30px Courier New";
    this.xSpacing = 10;
    this.ySpacing = 10;
    this.speed = 0.2;
    this.devicePixelRatio = window.devicePixelRatio || 1;

    this.yPositions = Array(Math.ceil(this.width / this.xSpacing))
      .fill(0)
      .map(() => Math.random() * (this.height / this.ySpacing));

    this.directions = Array(Math.ceil(this.width / this.xSpacing))
      .fill(0)
      .map(() => (Math.random() < 0.5 ? 1 : 1)); // 1 for down, -1 for up

    this.ySpeeds = this.yPositions.map(
      () => (Math.random() + 0.2) * this.speed
    );
    this.yTimes = this.yPositions.map(() => 0);
    this.lastChars = this.yPositions.map(() => " ");
    window.addEventListener("mousemove", (e) => this.onMouseMove(e));
  }

  onMouseMove(event) {
    const rect = this.canvas.getBoundingClientRect();
    const x = event.clientX - rect.left;
    const y = event.clientY - rect.top;

    const col = Math.floor(x / this.xSpacing);
    this.yPositions[col] = y / this.ySpacing;
    this.yTimes[col] = 0; // Reset the time for this column
  }

  updateDimensions() {
    const dpr = window.devicePixelRatio || 1;
    this.devicePixelRatio = dpr; // Store the dpr
    this.width = window.innerWidth;
    this.height = window.innerHeight;
    this.canvas.width = this.width * dpr;
    this.canvas.height = this.height * dpr;
    this.canvas.style.width = `${this.width}px`;
    this.canvas.style.height = `${this.height}px`;
  }

  draw() {
    requestAnimationFrame(this.draw.bind(this));

    // Drawing logic here
    this.ctx.fillStyle = "rgba(0, 0, 0, 0.05)";
    this.ctx.fillRect(0, 0, this.width, this.height);
    this.ctx.fillStyle = "#A678ED";

    const charArr = [
      "⫖",
      "⫈",
      "⪸",
      "⪬",
      "⫛",
      "⫏",
      "⫐",
      "⩱",
      "⩸",
      "⩦",
      "⩨",
      "⩢",
      "⩽",
      "⨺",
      "⨻",
      "⩥",
      "⩩",
      "⫒",
      "⫕",
      "⪫",
      "⪭",
      "⫑",
      "⫓",
      "⪷",
      "⪵",
    ];

    this.yPositions.forEach((y, i) => {
      if (this.yTimes[i] > 1) {
        const char = charArr[Math.floor(Math.random() * charArr.length)];
        this.lastChars[i] = char;

        this.ctx.fillText(
          char,
          i * this.xSpacing + 1,
          y * this.ySpacing + this.ySpacing
        );

        this.yPositions[i] =
          y + this.directions[i] < 0
            ? this.height / this.ySpacing
            : y + this.directions[i] >= this.height / this.ySpacing
            ? 0
            : y + this.directions[i];

        this.yTimes[i] = 0;
      }
      this.yTimes[i] += this.ySpeeds[i];
    });
  }

  start() {
    this.draw();
  }

  resize() {
    this.updateDimensions();
    this.ctx.setTransform(
      this.devicePixelRatio,
      0,
      0,
      this.devicePixelRatio,
      0,
      0
    );

    // You might also want to update the directions array here, if needed

    const columns = Math.ceil(this.width / this.xSpacing);
    while (this.yPositions.length < columns) {
      this.yPositions.push(Math.random() * (this.height / this.ySpacing));
      this.ySpeeds.push((Math.random() + 0.2) * this.speed);
      this.yTimes.push(0);
      this.lastChars.push(" ");
      this.directions.push(Math.random() < 0.5 ? 1 : 1);
    }

    if (this.yPositions.length > columns) {
      this.yPositions = this.yPositions.slice(0, columns);
      this.ySpeeds = this.ySpeeds.slice(0, columns);
      this.yTimes = this.yTimes.slice(0, columns);
      this.lastChars = this.lastChars.slice(0, columns);
      this.directions = this.directions.slice(0, columns);
    }
  }
}

const matrix = new Matrix(document.getElementById("matrixCanvas"));
window.addEventListener("resize", () => matrix.resize());
matrix.resize(); // Initialize dimensions
matrix.start();
