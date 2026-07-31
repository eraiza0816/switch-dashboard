import { speedUnit, setSpeedUnit } from './dashboard-graph';

let setPaused: (paused: boolean) => void = () => {};
let refreshDashboard: () => void = () => {};

export function setPollingController(controller: {
  setPaused: (paused: boolean) => void;
  refresh: () => void;
}) {
  setPaused = controller.setPaused;
  refreshDashboard = controller.refresh;
}

export function focusMacSearch() {
  setPaused(true);
}

export function blurMacSearch() {
  setPaused(false);
}

export async function switchConsole(ip) {
  const cmd = prompt('Enter CLI command (e.g. show, port 5 1g):');
  if (!cmd) return;
  try {
    const res = await fetch(`/api/switches/${ip}/cmd`, {
      method: 'POST',
      headers: {'Content-Type':'application/json'},
      body: JSON.stringify({cmd})
    });
    const data = await res.json();
    alert('Response: ' + (data.response || 'OK'));
  } catch(e) {
    alert('Error: ' + e.message);
  }
}

export function backupConfig(ip, element) {
  if (element.dataset.loading === 'true') return;
  element.dataset.loading = 'true';
  const originalText = element.innerHTML;
  element.innerHTML = '⏳ Backing up...';
  element.style.color = '#8b949e';

  fetch(`/api/switches/${ip}/backup`, { method: 'POST' })
    .then(r => r.json())
    .then(data => {
      if (data.status === 'ok') {
        element.innerHTML = 'Backed up!';
        element.style.color = '#3fb950';
      } else {
        alert('Failed to backup config: ' + (data.error || 'Unknown error'));
        element.innerHTML = 'Failed';
        element.style.color = '#ff7b72';
      }
    })
    .catch(err => {
      alert('Error during backup: ' + err.message);
      element.innerHTML = 'Failed';
      element.style.color = '#ff7b72';
    })
    .finally(() => {
      setTimeout(() => {
        element.innerHTML = originalText;
        element.style.color = '#58a6ff';
        delete element.dataset.loading;
      }, 3000);
    });
}

export async function saveHost(input, mac) {
  const host = input.value;
  try {
    await fetch('/api/clients/update_host', {
      method: 'POST',
      headers: {'Content-Type': 'application/json'},
      body: JSON.stringify({mac: mac, host: host})
    });
  } catch (e) {
    console.error("Failed to save client host nickname:", e);
  }
  setPaused(false);
  refreshDashboard();
}

export function pausePolling() {
  setPaused(true);
}

export async function saveNote(input) {
  const key = input.dataset.key;
  const note = input.value;
  await fetch('/api/notes', {
    method: 'POST',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({key, note})
  });
  setPaused(false);
  refreshDashboard();
}

export function toggleSpeedUnit(event) {
  if (event) event.stopPropagation();
  setSpeedUnit(speedUnit === 'bps' ? 'Bps' : 'bps');
  refreshDashboard();
}

/* ---- Reset ---- */
export function showResetConfirm() {
  document.getElementById('confirm-overlay').classList.add('open');
}

export function hideResetConfirm() {
  document.getElementById('confirm-overlay').classList.remove('open');
}

export async function doReset() {
  hideResetConfirm();
  await fetch('/api/reset', { method: 'POST' });
  refreshDashboard();
}

/* ---- Font size ---- */
export function setFontSize(size) {
  document.body.dataset.fontSize = size;
  document.querySelectorAll('.fs-btn').forEach(b => (b as HTMLElement).classList.toggle('fs-active', (b as HTMLElement).dataset.size === size));
  fetch('/api/settings', {
    method: 'POST',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({font_size: size})
  });
}

export async function loadSettings() {
  try {
    const r = await fetch('/api/settings');
    const s = await r.json();
    if (s.font_size) {
      setFontSize(s.font_size);
    }
  } catch(e) {}
}
