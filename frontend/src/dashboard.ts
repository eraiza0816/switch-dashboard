import './i18n';
import { REFRESH_SECONDS, formatBytes } from './dashboard-utils';
import { currentGraphIp, currentGraphPort, openGraph, renderGraph } from './dashboard-graph';
import { openTransceiver, closeTransceiver } from './dashboard-transceiver';
import {
  toggleMacTable, toggleIgmpTable, toggleExtraSection, sortMac, filterMacTable, manualRefreshMac,
  saveMacScroll, saveMacHeight, handleResizeStart, handleDragStart, handleDragEnter, handleDragOver,
  handleDragLeave, handleDragEnd, handleDrop, getMacState, saveMacStatesToStorage, renderMacTable, renderSwitch,
  setRefreshDashboard,
} from './dashboard-table';
import {
  setPollingController, pausePolling, focusMacSearch, blurMacSearch, switchConsole, backupConfig,
  saveHost, saveNote, toggleSpeedUnit, showResetConfirm, hideResetConfirm, doReset, setFontSize, loadSettings,
} from './dashboard-inputs';

let pollingPaused = false;
let countdownSeconds = REFRESH_SECONDS;
let countdownInterval = null;

function resetCountdown() {
  countdownSeconds = REFRESH_SECONDS;
  const liveText = document.getElementById('live-text');
  if (liveText) {
    if (pollingPaused) {
      liveText.textContent = 'Paused';
    } else {
      liveText.textContent = `Live (update in ${countdownSeconds}s)`;
    }
  }
}

function startCountdown() {
  if (countdownInterval) clearInterval(countdownInterval);
  countdownInterval = setInterval(() => {
    const liveText = document.getElementById('live-text');
    const liveDot = document.getElementById('live-dot');
    if (!liveText || !liveDot) return;

    if (pollingPaused) {
      liveText.textContent = 'Paused';
      liveDot.style.background = '#d29922';
      return;
    }

    countdownSeconds--;
    if (countdownSeconds <= 0) {
      liveText.textContent = 'Updating...';
      liveDot.style.background = '#d29922';
      countdownSeconds = REFRESH_SECONDS;
      updateDashboard();
    } else {
      liveText.textContent = `Live (update in ${countdownSeconds}s)`;
      liveDot.style.background = '#3fb950';
    }
  }, 1000);
}

async function updateDashboard() {
  if (pollingPaused) return;
  const liveDot = document.getElementById('live-dot');
  const liveText = document.getElementById('live-text');
  liveDot.style.background = '#d29922';
  liveText.textContent = 'Updating...';
  try {
    const r = await fetch('/api/switches');
    const data = await r.json();
    const app = document.getElementById('app');
    if (!data || data.length === 0) {
      app.innerHTML = '<div style="text-align:center;padding:40px;color:#b1bac4">No nodes configured. <a href="/config" style="color:#58a6ff">Configure now</a></div>';
      app.className = '';
      liveDot.style.background = '#f85149';
      liveText.textContent = 'No nodes';
      return;
    }

    // Save scroll position and height of active MAC tables before recreating elements
    data.forEach(sw => {
      const container = document.querySelector(`.mac-scroll-container-${CSS.escape(sw.ip)}`) as HTMLElement | null;
      if (container) {
        const state = getMacState(sw.ip);
        state.scrollTop = container.scrollTop;
        if (container.style.height) {
          state.height = container.style.height;
        }
      }
    });
    saveMacStatesToStorage();

    app.className = 'switches';
    app.innerHTML = data.map(renderSwitch).join('');

    // Initialize states and render active MAC tables
    data.forEach(sw => {
      const state = getMacState(sw.ip);
      state.data = sw.mac_table || [];
      if (state.expanded) {
        renderMacTable(sw.ip);
      }
    });

    if (document.getElementById('graph-overlay').classList.contains('open') && currentGraphIp && currentGraphPort) {
      renderGraph(currentGraphIp, currentGraphPort);
      const currentSw = data.find(sw => sw.ip === currentGraphIp);
      if (currentSw) {
        const currentPortObj = (currentSw.ports || []).find(p => p.port === currentGraphPort);
        if (currentPortObj) {
          const totalTx = currentPortObj.cum_tx || currentPortObj.tx_bytes || 0;
          const totalRx = currentPortObj.cum_rx || currentPortObj.rx_bytes || 0;
          document.getElementById('graph-total-tx').innerHTML = `${totalTx.toLocaleString()} bytes <span style="font-weight:400;color:#8b949e">(${formatBytes(totalTx)})</span>`;
          document.getElementById('graph-total-rx').innerHTML = `${totalRx.toLocaleString()} bytes <span style="font-weight:400;color:#8b949e">(${formatBytes(totalRx)})</span>`;
        }
      }
    }
    liveDot.style.background = '#3fb950';
    resetCountdown();
  } catch (e) {
    document.getElementById('app').innerHTML =
      `<div style="text-align:center;padding:40px;color:#f85149">Connection error: ${e.message}</div>`;
    liveDot.style.background = '#f85149';
    liveText.textContent = 'Error';
  }
}

loadSettings();
updateDashboard();
startCountdown();
setRefreshDashboard(updateDashboard);
setPollingController({
  setPaused: (paused: boolean) => {
    pollingPaused = paused;
    const liveDot = document.getElementById('live-dot');
    const liveText = document.getElementById('live-text');
    if (paused) {
      liveDot.style.background = '#d29922';
      liveText.textContent = 'Paused';
    } else {
      liveDot.style.background = '#3fb950';
      resetCountdown();
    }
  },
  refresh: updateDashboard,
});
window.addEventListener('resize', () => {
  if (document.getElementById('graph-overlay').classList.contains('open') && currentGraphIp && currentGraphPort) {
    renderGraph(currentGraphIp, currentGraphPort);
  }
});
window.setFontSize = setFontSize;
window.doReset = doReset;
window.backupConfig = backupConfig;
window.manualRefreshMac = manualRefreshMac;
window.openGraph = openGraph;
window.openTransceiver = openTransceiver;
window.closeTransceiver = closeTransceiver;
window.sortMac = sortMac;
window.toggleIgmpTable = toggleIgmpTable;
window.toggleMacTable = toggleMacTable;
window.toggleSpeedUnit = toggleSpeedUnit;
window.switchConsole = switchConsole;
window.pausePolling = pausePolling;
window.saveNote = saveNote;
window.saveHost = saveHost;
window.focusMacSearch = focusMacSearch;
window.blurMacSearch = blurMacSearch;
window.saveMacHeight = saveMacHeight;
window.handleResizeStart = handleResizeStart;
window.handleDragStart = handleDragStart;
window.handleDragEnter = handleDragEnter;
window.handleDragOver = handleDragOver;
window.handleDragLeave = handleDragLeave;
window.handleDragEnd = handleDragEnd;
window.handleDrop = handleDrop;
window.filterMacTable = filterMacTable;
window.saveMacScroll = saveMacScroll;
(window as any).toggleExtraSection = toggleExtraSection;
