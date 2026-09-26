const state = {
  token: localStorage.getItem("panel_token") || "",
  currentPath: "/",
};

function wsUrl(path) {
  const proto = location.protocol === "https:" ? "wss:" : "ws:";
  return `${proto}//${location.host}${path}?token=${encodeURIComponent(state.token)}`;
}

function apiUrl(path) {
  const sep = path.includes("?") ? "&" : "?";
  return `${path}${sep}token=${encodeURIComponent(state.token)}`;
}

// ---------- Connexion ----------
async function tryLogin(token) {
  const res = await fetch(apiUrl("/api/files/list"), {
    headers: { Authorization: `Bearer ${token}` },
  });
  return res.ok;
}

document.getElementById("loginBtn").addEventListener("click", async () => {
  const token = document.getElementById("tokenInput").value.trim();
  if (!token) return;
  const ok = await tryLogin(token);
  if (!ok) {
    document.getElementById("loginError").textContent = "Token invalide.";
    return;
  }
  state.token = token;
  localStorage.setItem("panel_token", token);
  enterApp();
});

async function enterApp() {
  document.getElementById("login").classList.add("hidden");
  document.getElementById("app").classList.remove("hidden");
  initTerminal();
  loadFiles(state.currentPath);
  loadServices();
}

// Auto-login si un token est déjà stocké
if (state.token) {
  tryLogin(state.token).then((ok) => {
    if (ok) enterApp();
  });
}

// ---------- Onglets ----------
document.querySelectorAll(".tab").forEach((btn) => {
  btn.addEventListener("click", () => {
    document.querySelectorAll(".tab").forEach((b) => b.classList.remove("active"));
    document.querySelectorAll(".tab-content").forEach((c) => c.classList.remove("active"));
    btn.classList.add("active");
    document.getElementById(`tab-${btn.dataset.tab}`).classList.add("active");
    if (btn.dataset.tab === "terminal" && window.fitAddon) {
      setTimeout(() => window.fitAddon.fit(), 50);
    }
  });
});

// ---------- Terminal ----------
function initTerminal() {
  if (window.__termInit) return;
  window.__termInit = true;

  const term = new Terminal({
    theme: { background: "#0d1117" },
    fontSize: 14,
    cursorBlink: true,
  });
  const fitAddon = new FitAddon.FitAddon();
  term.loadAddon(fitAddon);
  term.open(document.getElementById("terminal"));
  fitAddon.fit();
  window.fitAddon = fitAddon;

  const socket = new WebSocket(wsUrl("/api/terminal"));
  socket.binaryType = "arraybuffer";

  socket.onmessage = (ev) => {
    const data = typeof ev.data === "string" ? ev.data : new Uint8Array(ev.data);
    term.write(typeof data === "string" ? data : new TextDecoder().decode(data));
  };

  term.onData((data) => socket.readyState === 1 && socket.send(data));

  const sendResize = () => {
    if (socket.readyState !== 1) return;
    socket.send(JSON.stringify({ type: "resize", cols: term.cols, rows: term.rows }));
  };
  socket.onopen = sendResize;
  window.addEventListener("resize", () => {
    fitAddon.fit();
    sendResize();
  });
}

// ---------- Fichiers ----------
async function loadFiles(path) {
  state.currentPath = path;
  document.getElementById("filesPath").textContent = path;
  const res = await fetch(apiUrl(`/api/files/list?path=${encodeURIComponent(path)}`));
  const items = await res.json();

  const list = document.getElementById("filesList");
  list.innerHTML = "";

  if (path !== "/") {
    const up = document.createElement("li");
    up.innerHTML = `<span>⬆️ ..</span>`;
    up.onclick = () => loadFiles(path.split("/").slice(0, -1).join("/") || "/");
    list.appendChild(up);
  }

  items.forEach((it) => {
    const li = document.createElement("li");
    const icon = it.isDir ? "📁" : "📄";
    li.innerHTML = `
      <span class="entry-name">${icon} ${it.name}</span>
      <span>
        ${it.isDir ? "" : `<button data-act="dl">⬇️</button>`}
        <button data-act="rm">🗑️</button>
      </span>`;
    li.querySelector(".entry-name").onclick = () => {
      if (it.isDir) loadFiles(it.path);
    };
    const dlBtn = li.querySelector('[data-act="dl"]');
    if (dlBtn) dlBtn.onclick = () => window.open(apiUrl(`/api/files/download?path=${encodeURIComponent(it.path)}`));
    li.querySelector('[data-act="rm"]').onclick = async (e) => {
      e.stopPropagation();
      if (!confirm(`Supprimer ${it.name} ?`)) return;
      await fetch(apiUrl(`/api/files/delete?path=${encodeURIComponent(it.path)}`), { method: "DELETE" });
      loadFiles(path);
    };
    list.appendChild(li);
  });
}

document.getElementById("uploadInput").addEventListener("change", async (e) => {
  const file = e.target.files[0];
  if (!file) return;
  const fd = new FormData();
  fd.append("path", state.currentPath);
  fd.append("file", file);
  await fetch(apiUrl("/api/files/upload"), { method: "POST", body: fd });
  loadFiles(state.currentPath);
  e.target.value = "";
});

// ---------- Services ----------
async function loadServices() {
  const res = await fetch(apiUrl("/api/services/list"));
  const units = await res.json();
  const list = document.getElementById("servicesList");
  list.innerHTML = "";

  units.forEach((u) => {
    const li = document.createElement("li");
    li.innerHTML = `
      <div>
        <div class="svc-name">${u.name}</div>
        <div class="svc-state ${u.activeState}">${u.activeState} / ${u.subState}</div>
      </div>
      <div class="svc-actions">
        <button data-act="start">▶️</button>
        <button data-act="stop">⏹️</button>
        <button data-act="restart">🔄</button>
      </div>`;
    ["start", "stop", "restart"].forEach((act) => {
      li.querySelector(`[data-act="${act}"]`).onclick = async () => {
        await fetch(apiUrl(`/api/services/action?name=${encodeURIComponent(u.name)}&action=${act}`), { method: "POST" });
        setTimeout(loadServices, 800);
      };
    });
    list.appendChild(li);
  });
}

// ---------- Logs ----------
let logSocket = null;
document.getElementById("logUnit").addEventListener("keydown", (e) => {
  if (e.key !== "Enter") return;
  const unit = e.target.value.trim();
  const out = document.getElementById("logsOutput");
  out.textContent = "";
  if (logSocket) logSocket.close();
  logSocket = new WebSocket(wsUrl(`/api/logs${unit ? `&unit=${encodeURIComponent(unit)}` : ""}`));
  logSocket.onmessage = (ev) => {
    out.textContent += ev.data + "\n";
    out.scrollTop = out.scrollHeight;
  };
});
