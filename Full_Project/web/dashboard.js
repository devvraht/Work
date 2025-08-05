console.log("⚡ Dashboard Loaded");

const MAX_SLOTS = 200;
const droneContainer = document.getElementById("droneContainer");
const droneMap = {};
const latestFrames = {};
const streamActive = {};

// Create slots
for (let i = 0; i < MAX_SLOTS; i++) {
  const slot = document.createElement("div");
  slot.className = "drone";

  const placeholder = document.createElement("img");
  placeholder.src = "image.png";
  placeholder.className = "placeholder";
  slot.appendChild(placeholder);

  const canvas = document.createElement("canvas");
  canvas.width = 160;
  canvas.height = 120;
  canvas.style.display = "none";
  slot.appendChild(canvas);

  const label = document.createElement("div");
  label.className = "label";
  label.textContent = "Drone :";
  slot.appendChild(label);

  droneContainer.appendChild(slot);
  canvas.placeholder = placeholder;

  canvas.addEventListener("click", () => {
    const droneID = Object.keys(droneMap).find(key => droneMap[key] === canvas);
    if (droneID && streamActive[droneID]) {
      window.location.href = `streamer.html?droneID=${encodeURIComponent(droneID)}`;
    }
  });
}

// WebSocket
let ws;
function connectWS() {
  ws = new WebSocket("ws://localhost:8080/view");

  ws.onmessage = (event) => {
    const [droneID, base64] = event.data.split("|");
    if (!droneID || !base64) return;
    latestFrames[droneID] = base64; // Drop old frames
  };

  ws.onclose = () => {
    console.warn("WebSocket closed. Reconnecting...");
    setTimeout(connectWS, 1000);
  };
}
connectWS();

// Render loop ~60fps
function renderLoop() {
  for (const droneID in latestFrames) {
    const base64 = latestFrames[droneID];
    const canvas = droneMap[droneID] || assignSlot(droneID);
    if (!canvas) continue;

    streamActive[droneID] = true;
    const ctx = canvas.getContext("2d");

    const img = new Image();
    img.src = "data:image/jpeg;base64," + base64;
    img.onload = () => {
      if (canvas.style.display === "none") {
        canvas.style.display = "block";
        if (canvas.placeholder) canvas.placeholder.style.display = "none";
      }
      ctx.clearRect(0, 0, canvas.width, canvas.height);
      ctx.drawImage(img, 0, 0, canvas.width, canvas.height);
    };
  }
  requestAnimationFrame(renderLoop);
}

function assignSlot(droneID) {
  const slots = document.querySelectorAll(".drone");
  for (let slot of slots) {
    const label = slot.querySelector(".label");
    if (label.textContent === "Drone :") {
      const canvas = slot.querySelector("canvas");
      label.textContent = `Drone : ${droneID.replace("drone_", "")}`;
      droneMap[droneID] = canvas;
      streamActive[droneID] = false;
      return canvas;
    }
  }
  return null;
}

renderLoop();
