import { REFRESH_SECONDS, state, formatBytes, formatBps, formatPkts, speedClass, formatTime, formatNumber, parseHexPort } from './dashboard-utils';

export function refreshTransceiver() {
  if (state.currentTransceiverIp && state.currentTransceiverPort) {
    (window as any).openTransceiver(state.currentTransceiverIp, state.currentTransceiverPort, state.currentTransceiverSwitchName);
  }
}

export function openGraph(ip: string, port: string, name: string, label: string, totalTx = 0, totalRx = 0) {
  state.currentGraphIp = ip;
  state.currentGraphPort = port;
  state.currentGraphSwitchName = name;
  state.currentGraphPortLabel = label;
  document.getElementById('graph-title')!.textContent = `Bandwidth - ${label} (${name})`;
  document.getElementById('graph-stats')!.innerHTML = `
    <div class="info-item"><div class="label">Peak TX</div><div class="value" id="graph-peak-tx">-</div></div>
    <div class="info-item"><div class="label">Peak RX</div><div class="value" id="graph-peak-rx">-</div></div>
    <div class="info-item"><div class="label">Current TX</div><div class="value" id="graph-now-tx">-</div></div>
    <div class="info-item"><div class="label">Current RX</div><div class="value" id="graph-now-rx">-</div></div>
    <div class="info-item"><div class="label">Total TX</div><div class="value" id="graph-total-tx">${formatBytes(totalTx)}</div></div>
    <div class="info-item"><div class="label">Total RX</div><div class="value" id="graph-total-rx">${formatBytes(totalRx)}</div></div>`;
  document.getElementById('graph-sub')!.textContent = `Cumulative traffic for port ${port} - updated every ${REFRESH_SECONDS}s`;
  document.getElementById('graph-overlay')!.classList.add('active');
  renderGraph(ip, port);
}

export function setGraphRange(range: string) {
  state.currentGraphRange = range;
  document.querySelectorAll('.graph-tab').forEach(t => t.classList.remove('active'));
  const tab = document.querySelector(`.graph-tab[data-range="${range}"]`);
  if (tab) tab.classList.add('active');
  if (state.currentGraphIp && state.currentGraphPort) renderGraph(state.currentGraphIp, state.currentGraphPort);
}

export function closeGraph() {
  document.getElementById('graph-overlay')!.classList.remove('active');
}

export function toggleMaximizeGraph() {
  const box = document.getElementById('graph-box')!;
  box.classList.toggle('maximized');
  const btn = document.getElementById('graph-maximize-btn')!;
  if (box.classList.contains('maximized')) {
    btn.innerHTML = `<svg width="14" height="14" viewBox="0 0 16 16" fill="currentColor"><path d="M5.5 1.5a.5.5 0 0 0-1 0v2h-2a.5.5 0 0 0 0 1h2.5a.5.5 0 0 0 .5-.5v-2.5zM11.5 1.5a.5.5 0 0 0-1 0v2.5a.5.5 0 0 0 .5.5h2.5a.5.5 0 0 0 0-1h-2v-2zM4.5 11.5v2a.5.5 0 0 0 1 0v-2.5a.5.5 0 0 0-.5-.5h-2.5a.5.5 0 0 0 0 1h2zM12.5 11.5h-2a.5.5 0 0 0-.5.5v2.5a.5.5 0 0 0 1 0v-2h2a.5.5 0 0 0 0-1z"/></svg>`;
    btn.title = "Restore Size";
  } else {
    btn.innerHTML = `<svg width="14" height="14" viewBox="0 0 16 16" fill="currentColor"><path d="M1.5 1a.5.5 0 0 0-.5.5v3a.5.5 0 0 0 1 0v-2h2a.5.5 0 0 0 0-1h-2.5zM11.5 1a.5.5 0 0 0 0 1h2v2a.5.5 0 0 0 1 0v-3a.5.5 0 0 0-.5-.5h-2.5zM1 11.5a.5.5 0 0 0 .5.5h2a.5.5 0 0 0 0-1h-2v-2a.5.5 0 0 0-1 0v3zM14 11.5a.5.5 0 0 0-.5-.5h-2a.5.5 0 0 0 0 1h2v2a.5.5 0 0 0 1 0v-3z"/></svg>`;
    btn.title = "Toggle Fullscreen";
  }
  if (state.currentGraphIp && state.currentGraphPort) {
    renderGraph(state.currentGraphIp, state.currentGraphPort);
  }
}

function renderGraph(ip: string, port: string) {
  const svg = document.getElementById('graph-svg')!;
  const box = document.getElementById('graph-box')!;
  const isMax = box.classList.contains('maximized');
  const W = Math.max(500, box.clientWidth - (isMax ? 64 : 48));
  let H = window.innerHeight * 0.35;
  if (isMax) {
    H = window.innerHeight - 250;
  }
  H = Math.max(240, H);
  svg.setAttribute('viewBox', `0 0 ${W} ${H}`);
  svg.innerHTML = '';

  const pad = { top: 20, bottom: 40, left: 70, right: 30 };
  const plotW = W - pad.left - pad.right;
  const plotH = H - pad.top - pad.bottom;

  fetch(`/api/history?ip=${ip}&port=${port}&range=${state.currentGraphRange}`)
    .then(r => r.json())
    .then(data => {
      if (!data || !data.tx || data.tx.length < 2) {
        svg.innerHTML = `<text x="${W/2}" y="${H/2}" text-anchor="middle" fill="#b1bac4" font-size="14">No history data available for this range yet</text>`;
        document.getElementById('graph-peak-tx')!.textContent = '-';
        document.getElementById('graph-peak-rx')!.textContent = '-';
        document.getElementById('graph-now-tx')!.textContent = '-';
        document.getElementById('graph-now-rx')!.textContent = '-';
        return;
      }

      const txPoints: number[] = data.tx;
      const rxPoints: number[] = data.rx;
      const tss: number[] = data.timestamps;
      const len = txPoints.length;

      let maxVal = 0;
      for (let i = 0; i < len; i++) {
        if (txPoints[i] > maxVal) maxVal = txPoints[i];
        if (rxPoints[i] > maxVal) maxVal = rxPoints[i];
      }
      if (maxVal === 0) maxVal = 1000;
      maxVal *= 1.15;

      const peakTx = Math.max(...txPoints);
      const peakRx = Math.max(...rxPoints);
      const nowTx = txPoints[len - 1];
      const nowRx = rxPoints[len - 1];

      document.getElementById('graph-peak-tx')!.textContent = formatBps(peakTx, state.speedUnit);
      document.getElementById('graph-peak-rx')!.textContent = formatBps(peakRx, state.speedUnit);
      document.getElementById('graph-now-tx')!.textContent = formatBps(nowTx, state.speedUnit);
      document.getElementById('graph-now-rx')!.textContent = formatBps(nowRx, state.speedUnit);

      const toX = (i: number) => pad.left + (i / (len - 1)) * plotW;
      const toY = (v: number) => pad.top + plotH - (v / maxVal) * plotH;

      let yTicks = 4;
      let gridSvg = '';
      for (let i = 0; i <= yTicks; i++) {
        const v = (maxVal / yTicks) * i;
        const y = toY(v);
        gridSvg += `<line x1="${pad.left}" y1="${y}" x2="${W - pad.right}" y2="${y}" stroke="#21262d" stroke-width="1" stroke-dasharray="${i === 0 ? '0' : '4 4'}"/>`;
        gridSvg += `<text x="${pad.left - 10}" y="${y + 4}" text-anchor="end" fill="#b1bac4" font-size="11" font-family="'SF Mono', Monaco, monospace">${formatBps(v, state.speedUnit)}</text>`;
      }

      let xTicks: number[] = [];
      const tickCount = Math.min(5, len);
      const step = Math.max(1, Math.floor((len - 1) / (tickCount - 1)));
      for (let i = 0; i < len; i += step) {
        xTicks.push(i);
      }
      if (xTicks[xTicks.length - 1] !== len - 1) {
        xTicks.push(len - 1);
      }

      const xTicksSvg = xTicks.map(idx => {
        const x = toX(idx);
        const date = new Date(tss[idx] * 1000);
        let timeStr = date.toLocaleTimeString([], {hour: '2-digit', minute: '2-digit', second: '2-digit'});
        if (state.currentGraphRange === '24h') {
          timeStr = date.toLocaleTimeString([], {hour: '2-digit', minute: '2-digit'}) + '<br>' + (date.getMonth()+1) + '/' + date.getDate();
        }
        return `<text x="${x}" y="${H - 5}" text-anchor="middle" fill="#b1bac4" font-size="10" font-family="'SF Mono', Monaco, monospace">${timeStr}</text>`;
      }).join('');

      const lineLayer = `<path d="${txPoints.map((v, i) => `${i === 0 ? 'M' : 'L'}${toX(i)},${toY(v)}`).join(' ')}" fill="none" stroke="#58a6ff" stroke-width="2"/>`;
      const lineLayerRx = `<path d="${rxPoints.map((v, i) => `${i === 0 ? 'M' : 'L'}${toX(i)},${toY(v)}`).join(' ')}" fill="none" stroke="#3fb950" stroke-width="2"/>`;

      let tooltipRect = '';
      if (len > 0) {
        tooltipRect = `<rect id="hover-overlay-rect" x="${pad.left}" y="0" width="${plotW}" height="${H}" fill="transparent" style="cursor:crosshair"/>
          <line id="hover-line" x1="0" y1="${pad.top}" x2="0" y2="${pad.top + plotH}" stroke="#58a6ff" stroke-width="1" style="display:none"/>
          <circle id="hover-dot-tx" r="4" fill="#58a6ff" style="display:none"/>
          <circle id="hover-dot-rx" r="4" fill="#3fb950" style="display:none"/>`;
      }

      svg.innerHTML = gridSvg + lineLayer + lineLayerRx + xTicksSvg + tooltipRect;
    })
    .catch(err => {
      svg.innerHTML = `<text x="${W/2}" y="${H/2}" text-anchor="middle" fill="#f85149" font-size="14">Error: ${err.message}</text>`;
    });
}
