let pollingInterval = null;
let isPollingActive = true;
let isUserScrolledUp = false;
let isUpdatingLogs = false;
let lastRenderedLines = [];

document.addEventListener("DOMContentLoaded", () => {
  // Bind console scroll to check if user has scrolled up to inspect previous lines
  const consoleScreen = document.getElementById("console-screen");
  const banner = document.getElementById("autoscroll-banner");
  
  consoleScreen.addEventListener("scroll", () => {
    // Ignore scroll events triggered programmatically during log render cycles
    if (isUpdatingLogs) return;
    
    // If they are near the bottom (within 40px), auto-scroll is allowed
    const threshold = 40;
    const isAtBottom = (consoleScreen.scrollHeight - consoleScreen.scrollTop - consoleScreen.clientHeight) <= threshold;
    isUserScrolledUp = !isAtBottom;
    
    // Toggle floating banner visibility
    if (isUserScrolledUp) {
      banner.style.display = "flex";
    } else {
      banner.style.display = "none";
    }
  });

  // Initial fetch and start polling
  fetchLogs();
  startPolling();
});

function startPolling() {
  if (pollingInterval) clearInterval(pollingInterval);
  pollingInterval = setInterval(fetchLogs, 2000);
  isPollingActive = true;
  
  const dot = document.getElementById("refresh-dot");
  const text = document.getElementById("refresh-btn-text");
  dot.classList.add("active");
  text.textContent = "Polling Active";
}

function stopPolling() {
  if (pollingInterval) clearInterval(pollingInterval);
  isPollingActive = false;
  
  const dot = document.getElementById("refresh-dot");
  const text = document.getElementById("refresh-btn-text");
  dot.classList.remove("active");
  text.textContent = "Polling Paused";
}

function toggleAutoRefresh() {
  if (isPollingActive) {
    stopPolling();
    showToast("Auto-refresh paused", "success", "⏸️");
  } else {
    startPolling();
    showToast("Auto-refresh resumed", "success", "▶️");
  }
}

async function fetchLogs() {
  try {
    const res = await fetch("/api/logs");
    if (!res.ok) throw new Error("HTTP error " + res.status);
    const data = await res.json();
    
    if (data.error) {
      showErrorToast("Failed to fetch logs: " + data.error);
      return;
    }
    
    renderLogs(data);
  } catch (err) {
    console.error("Error fetching logs:", err);
  }
}

function parseLogLine(line) {
  // Typical log line: 2026-05-25 11:14:25,123 - scraper - INFO - Log message content
  // Or: 2026-05-25 11:14:25 - scraper - INFO - Log message content
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
  
  // Fallback for lines that don't match format perfectly
  let cleanLine = escapeHtml(line);
  let classLevel = "info";
  if (line.includes("DEBUG")) classLevel = "debug";
  else if (line.includes("WARNING") || line.includes("WARN")) classLevel = "warning";
  else if (line.includes("ERROR")) classLevel = "error";
  else if (line.includes("CRITICAL")) classLevel = "critical";
  
  return `<div class="log-line log-${classLevel}">${cleanLine}</div>`;
}

function renderLogs(lines) {
  const emptyState = document.getElementById("empty-state");
  const container = document.getElementById("logs-container");
  const consoleScreen = document.getElementById("console-screen");
  
  if (!lines || lines.length === 0) {
    emptyState.style.display = "flex";
    container.style.display = "none";
    lastRenderedLines = [];
    container.innerHTML = "";
    return;
  }
  
  emptyState.style.display = "none";
  container.style.display = "block";
  
  // Set update flag to ignore scroll events triggered by this render cycle
  isUpdatingLogs = true;
  
  // Try to find if we can just append new lines to avoid destroying DOM and resetting scroll anchoring
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
      
      // Limit DOM size to 1000 lines for efficiency
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
    // Fallback: render all lines
    const parsedHtml = lines.map(parseLogLine).join('');
    container.innerHTML = parsedHtml;
    lastRenderedLines = [...lines];
    
    if (!isUserScrolledUp) {
      consoleScreen.scrollTop = consoleScreen.scrollHeight;
    }
  }
  
  // Release update flag asynchronously to allow browser layout calculations to settle
  setTimeout(() => {
    isUpdatingLogs = false;
  }, 50);
}

async function updateLogLevel(level) {
  lastRenderedLines = []; // Force full re-render on next fetch
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
      showToast(`Log level dynamically changed to ${level}`, "success", "⚙️");
      // Fetch logs immediately to display the server confirmation log
      setTimeout(fetchLogs, 400);
    } else {
      showErrorToast("Failed to update log level: " + (data.error || "Unknown error"));
    }
  } catch (err) {
    showErrorToast("Network error: " + err.message);
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
      showToast("Logs successfully cleared on server", "success", "🗑️");
      lastRenderedLines = [];
      document.getElementById("logs-container").innerHTML = "";
      document.getElementById("empty-state").style.display = "flex";
      document.getElementById("logs-container").style.display = "none";
    } else {
      showErrorToast("Failed to clear logs: " + (data.error || "Unknown error"));
    }
  } catch (err) {
    showErrorToast("Network error: " + err.message);
  }
}

function resumeAutoscrollClick() {
  isUserScrolledUp = false;
  document.getElementById("autoscroll-banner").style.display = "none";
  const consoleScreen = document.getElementById("console-screen");
  consoleScreen.scrollTop = consoleScreen.scrollHeight;
  showToast("Auto-scroll resumed", "success", "⬇️");
}

// Helpers
function escapeHtml(unsafe) {
  return unsafe
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&#039;");
}

function showSuccessToast(message) {
  showToast(message, "success", "✓");
}

function showErrorToast(message) {
  showToast(message, "error", "⚠️");
}

function showToast(message, type, icon) {
  const toast = document.getElementById("toast-box");
  const toastIcon = document.getElementById("toast-icon");
  const toastMsg = document.getElementById("toast-message");
  
  toast.className = `toast ${type}`;
  toastIcon.textContent = icon;
  toastMsg.textContent = message;
  
  toast.classList.add("show");
  setTimeout(() => {
    toast.classList.remove("show");
  }, 4000);
}
