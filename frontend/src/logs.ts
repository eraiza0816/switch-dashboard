import { showToast } from './dashboard-utils';

let pollingInterval = null;
let isPollingActive = true;
let isUserScrolledUp = false;
let isUpdatingLogs = false;
let lastRenderedLines = [];

document.addEventListener("DOMContentLoaded", () => {
  const consoleScreen = document.getElementById("console-screen")!;
  const banner = document.getElementById("autoscroll-banner")!;
  
  consoleScreen.addEventListener("scroll", () => {
    if (isUpdatingLogs) return;
    
    const threshold = 40;
    const isAtBottom = (consoleScreen.scrollHeight - consoleScreen.scrollTop - consoleScreen.clientHeight) <= threshold;
    isUserScrolledUp = !isAtBottom;
    
    if (isUserScrolledUp) {
      banner.style.display = "flex";
    } else {
      banner.style.display = "none";
    }
  });

  fetchLogs();
  startPolling();
});

function startPolling() {
  if (pollingInterval) clearInterval(pollingInterval);
  pollingInterval = setInterval(fetchLogs, 2000);
  isPollingActive = true;
  
  const dot = document.getElementById("refresh-dot")!;
  const text = document.getElementById("refresh-btn-text")!;
  dot.classList.add("active");
  text.textContent = "Polling Active";
}

function stopPolling() {
  if (pollingInterval) clearInterval(pollingInterval);
  isPollingActive = false;
  
  const dot = document.getElementById("refresh-dot")!;
  const text = document.getElementById("refresh-btn-text")!;
  dot.classList.remove("active");
  text.textContent = "Polling Paused";
}

function toggleAutoRefresh() {
  if (isPollingActive) {
    stopPolling();
    showToast("Auto-refresh paused", "success");
  } else {
    startPolling();
    showToast("Auto-refresh resumed", "success");
  }
}

async function fetchLogs() {
  try {
    const res = await fetch("/api/logs");
    if (!res.ok) throw new Error("HTTP error " + res.status);
    const data = await res.json();
    
    if (data.error) {
      showToast("Failed to fetch logs: " + data.error, "error");
      return;
    }
    
    renderLogs(data);
  } catch (err) {
    console.error("Error fetching logs:", err);
  }
}

function parseLogLine(line: string): string {
  const regex = /^(\d{4}-\d{2}-\d{2}\s\d{2}:\d{2}:\d{2}(?:,\d+)?)\s-\s([^\s]+)\s-\s([A-Z]+)\s-\s(.*)$/;
  const match = line.match(regex);
  
  if (match) {
    const [_, time, module, level, message] = match;
    const cleanLevel = level.toLowerCase();
    
    return `
      <div class="log-line log-${cleanLevel}">
        <span class="log-time">${time}</span>
        <span class="level-tag">${level}</span>
        <span class="log-module">[${module}]</span>
        <span class="log-content">${escapeHtml(message)}</span>
      </div>
    `;
  }
  
  let cleanLine = escapeHtml(line);
  let classLevel = "info";
  if (line.includes("DEBUG")) classLevel = "debug";
  else if (line.includes("WARNING") || line.includes("WARN")) classLevel = "warning";
  else if (line.includes("ERROR")) classLevel = "error";
  else if (line.includes("CRITICAL")) classLevel = "critical";
  
  return `<div class="log-line log-${classLevel}">${cleanLine}</div>`;
}

function renderLogs(lines: string[]) {
  const emptyState = document.getElementById("empty-state")!;
  const container = document.getElementById("logs-container")!;
  const consoleScreen = document.getElementById("console-screen")!;
  
  if (!lines || lines.length === 0) {
    emptyState.style.display = "flex";
    container.style.display = "none";
    lastRenderedLines = [];
    container.innerHTML = "";
    return;
  }
  
  emptyState.style.display = "none";
  container.style.display = "block";
  
  isUpdatingLogs = true;
  
  let appendStartIndex = -1;
  if (lastRenderedLines.length > 0) {
    const matchCount = Math.min(3, lastRenderedLines.length);
    const lastFew = lastRenderedLines.slice(-matchCount);
    
    for (let i = lines.length - matchCount; i >= 0; i--) {
      let match = true;
      for (let j = 0; j < matchCount; j++) {
        if (lines[i + j] !== lastFew[j]) {
          match = false;
          break;
        }
      }
      if (match) {
        appendStartIndex = i + matchCount;
        break;
      }
    }
  }
  
  if (appendStartIndex !== -1) {
    const newLinesToAppend = lines.slice(appendStartIndex);
    if (newLinesToAppend.length > 0) {
      const savedScrollTop = consoleScreen.scrollTop;
      const wasAtBottom = !isUserScrolledUp;
      
      const fragment = document.createDocumentFragment();
      newLinesToAppend.forEach(line => {
        const div = document.createElement("div");
        div.innerHTML = parseLogLine(line).trim();
        fragment.appendChild(div.firstElementChild || div);
      });
      container.appendChild(fragment);
      
      const maxDomLines = 1000;
      while (container.children.length > maxDomLines) {
        container.removeChild(container.firstChild);
      }
      
      lastRenderedLines = lastRenderedLines.concat(newLinesToAppend);
      if (lastRenderedLines.length > maxDomLines) {
        lastRenderedLines = lastRenderedLines.slice(-maxDomLines);
      }
      
      if (wasAtBottom) {
        consoleScreen.scrollTop = consoleScreen.scrollHeight;
      } else {
        consoleScreen.scrollTop = savedScrollTop;
      }
    }
  } else {
    const parsedHtml = lines.map(parseLogLine).join('');
    container.innerHTML = parsedHtml;
    lastRenderedLines = [...lines];
    
    if (!isUserScrolledUp) {
      consoleScreen.scrollTop = consoleScreen.scrollHeight;
    }
  }
  
  setTimeout(() => {
    isUpdatingLogs = false;
  }, 50);
}

async function updateLogLevel(level: string) {
  lastRenderedLines = [];
  try {
    const res = await fetch("/api/logs/level", {
      method: "POST",
      headers: {
        "Content-Type": "application/json"
      },
      body: JSON.stringify({ level: level })
    });
    const data = await res.json();
    
    if (res.ok && data.status === "ok") {
      showToast(`Log level dynamically changed to ${level}`, "success");
      setTimeout(fetchLogs, 400);
    } else {
      showToast("Failed to update log level: " + (data.error || "Unknown error"), "error");
    }
  } catch (err) {
    showToast("Network error: " + err.message, "error");
  }
}

async function clearLogs() {
  if (!confirm("Are you sure you want to permanently clear the log file on the server? This will truncate the file immediately.")) {
    return;
  }
  
  try {
    const res = await fetch("/api/logs/clear", { method: "POST" });
    const data = await res.json();
    
    if (res.ok && data.status === "ok") {
      showToast("Logs successfully cleared on server", "success");
      lastRenderedLines = [];
      document.getElementById("logs-container")!.innerHTML = "";
      document.getElementById("empty-state")!.style.display = "flex";
      document.getElementById("logs-container")!.style.display = "none";
    } else {
      showToast("Failed to clear logs: " + (data.error || "Unknown error"), "error");
    }
  } catch (err) {
    showToast("Network error: " + err.message, "error");
  }
}

function resumeAutoscrollClick() {
  isUserScrolledUp = false;
  document.getElementById("autoscroll-banner")!.style.display = "none";
  const consoleScreen = document.getElementById("console-screen")!;
  consoleScreen.scrollTop = consoleScreen.scrollHeight;
  showToast("Auto-scroll resumed", "success");
}

function escapeHtml(unsafe: string): string {
  return unsafe
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&#039;");
}



(window as any).startPolling = startPolling;
(window as any).stopPolling = stopPolling;
(window as any).toggleAutoRefresh = toggleAutoRefresh;
(window as any).fetchLogs = fetchLogs;
(window as any).updateLogLevel = updateLogLevel;
(window as any).clearLogs = clearLogs;
(window as any).resumeAutoscrollClick = resumeAutoscrollClick;
export {};
