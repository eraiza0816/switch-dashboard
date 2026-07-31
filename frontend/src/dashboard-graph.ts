import { REFRESH_SECONDS, formatBytes, formatBps } from './dashboard-utils';

export let currentGraphIp: string | null = null;
export let currentGraphPort: string | null = null;
export let currentGraphRange = 'live';
export let currentGraphSwitchName = '';
export let currentGraphPortLabel = '';
export let speedUnit = 'Bps';

export function setSpeedUnit(unit: string) {
  speedUnit = unit;
}

/* ---- Graph ---- */
export function openGraph(ip, port, name, label, totalTx = 0, totalRx = 0) {
  currentGraphIp = ip;
  currentGraphPort = port;
  currentGraphSwitchName = name;
  currentGraphPortLabel = label;
  
  document.getElementById('graph-title').textContent = label + ' Speed (' + name + ')';
  document.getElementById('graph-overlay').classList.add('open');
  
  // Update cumulative traffic display
  document.getElementById('graph-total-tx').innerHTML = `${totalTx.toLocaleString()} bytes <span style="font-weight:400;color:#8b949e">(${formatBytes(totalTx)})</span>`;
  document.getElementById('graph-total-rx').innerHTML = `${totalRx.toLocaleString()} bytes <span style="font-weight:400;color:#8b949e">(${formatBytes(totalRx)})</span>`;
  
  setGraphRange(currentGraphRange); // trigger render
}

export function setGraphRange(range) {
  currentGraphRange = range;
  
  // Update Active Tab Button styling
  document.querySelectorAll('.graph-tab-btn').forEach(btn => btn.classList.remove('active'));
  if (range === 'live') {
    document.getElementById('tab-live').classList.add('active');
    document.getElementById('graph-sub').textContent = `Real-time speed at ${REFRESH_SECONDS}s polling intervals`;
  } else if (range === '1h') {
    document.getElementById('tab-1h').classList.add('active');
    document.getElementById('graph-sub').textContent = `Last 1 hour speed history (at ${REFRESH_SECONDS}s polling intervals)`;
  } else if (range === '24h') {
    document.getElementById('tab-24h').classList.add('active');
    document.getElementById('graph-sub').textContent = 'Last 24 hours speed history (15-minute averages)';
  }
  
  renderGraph(currentGraphIp, currentGraphPort);
}

export function closeGraph() {
  document.getElementById('graph-overlay').classList.remove('open');
  const box = document.getElementById('graph-box');
  if (box.classList.contains('maximized')) {
    box.classList.remove('maximized');
    const btn = document.getElementById('graph-maximize-btn');
    btn.innerHTML = `<svg width="14" height="14" viewBox="0 0 16 16" fill="currentColor"><path d="M1.5 1a.5.5 0 0 0-.5.5v3a.5.5 0 0 0 1 0v-2h2a.5.5 0 0 0 0-1h-2.5zM11.5 1a.5.5 0 0 0 0 1h2v2a.5.5 0 0 0 1 0v-3a.5.5 0 0 0-.5-.5h-2.5zM1 11.5a.5.5 0 0 0 .5.5h2a.5.5 0 0 0 0-1h-2v-2a.5.5 0 0 0-1 0v3zM14 11.5a.5.5 0 0 0-.5-.5h-2a.5.5 0 0 0 0 1h2v2a.5.5 0 0 0 1 0v-3z"/></svg>`;
    btn.title = "Toggle Fullscreen";
  }
  currentGraphIp = null;
  currentGraphPort = null;
  document.getElementById('graph-tooltip').style.display = 'none';
}

export function toggleMaximizeGraph() {
  const box = document.getElementById('graph-box');
  box.classList.toggle('maximized');
  
  const btn = document.getElementById('graph-maximize-btn');
  if (box.classList.contains('maximized')) {
    btn.innerHTML = `<svg width="14" height="14" viewBox="0 0 16 16" fill="currentColor"><path d="M5.5 1.5a.5.5 0 0 0-1 0v2h-2a.5.5 0 0 0 0 1h2.5a.5.5 0 0 0 .5-.5v-2.5zM11.5 1.5a.5.5 0 0 0-1 0v2.5a.5.5 0 0 0 .5.5h2.5a.5.5 0 0 0 0-1h-2v-2zM4.5 11.5v2a.5.5 0 0 0 1 0v-2.5a.5.5 0 0 0-.5-.5h-2.5a.5.5 0 0 0 0 1h2zM12.5 11.5h-2a.5.5 0 0 0-.5.5v2.5a.5.5 0 0 0 1 0v-2h2a.5.5 0 0 0 0-1z"/></svg>`;
    btn.title = "Restore Size";
  } else {
    btn.innerHTML = `<svg width="14" height="14" viewBox="0 0 16 16" fill="currentColor"><path d="M1.5 1a.5.5 0 0 0-.5.5v3a.5.5 0 0 0 1 0v-2h2a.5.5 0 0 0 0-1h-2.5zM11.5 1a.5.5 0 0 0 0 1h2v2a.5.5 0 0 0 1 0v-3a.5.5 0 0 0-.5-.5h-2.5zM1 11.5a.5.5 0 0 0 .5.5h2a.5.5 0 0 0 0-1h-2v-2a.5.5 0 0 0-1 0v3zM14 11.5a.5.5 0 0 0-.5-.5h-2a.5.5 0 0 0 0 1h2v2a.5.5 0 0 0 1 0v-3z"/></svg>`;
    btn.title = "Toggle Fullscreen";
  }
  
  if (currentGraphIp && currentGraphPort) {
    renderGraph(currentGraphIp, currentGraphPort);
  }
}

export function renderGraph(ip, port) {
  const svg = document.getElementById('graph-svg');
  const box = document.getElementById('graph-box');
  const isMax = box.classList.contains('maximized');
  const W = Math.max(500, box.clientWidth - (isMax ? 64 : 48));
  let H = window.innerHeight * 0.35;
  if (isMax) {
    H = window.innerHeight - 250;
  }
  H = Math.max(240, H);
  svg.setAttribute('viewBox', `0 0 ${W} ${H}`);
  svg.innerHTML = ''; // clear

  const pad = { top: 20, bottom: 40, left: 70, right: 30 };
  const plotW = W - pad.left - pad.right;
  const plotH = H - pad.top - pad.bottom;

  fetch(`/api/history?ip=${ip}&port=${port}&range=${currentGraphRange}`)
    .then(r => r.json())
    .then(data => {
      if (!data || !data.tx || data.tx.length < 2) {
        svg.innerHTML = `<text x="${W/2}" y="${H/2}" text-anchor="middle" fill="#b1bac4" font-size="14">No history data available for this range yet</text>`;
        document.getElementById('graph-peak-tx').textContent = '-';
        document.getElementById('graph-peak-rx').textContent = '-';
        document.getElementById('graph-now-tx').textContent = '-';
        document.getElementById('graph-now-rx').textContent = '-';
        return;
      }

      const txPoints = data.tx;
      const rxPoints = data.rx;
      const tss = data.timestamps;
      const len = txPoints.length;

      let maxVal = 0;
      for (let i = 0; i < len; i++) {
        if (txPoints[i] > maxVal) maxVal = txPoints[i];
        if (rxPoints[i] > maxVal) maxVal = rxPoints[i];
      }
      if (maxVal === 0) maxVal = 1000; // fallback scale
      maxVal *= 1.15; // padding top

      const peakTx = Math.max(...txPoints);
      const peakRx = Math.max(...rxPoints);
      const nowTx = txPoints[len - 1];
      const nowRx = rxPoints[len - 1];

      document.getElementById('graph-peak-tx').textContent = formatBps(peakTx, speedUnit);
      document.getElementById('graph-peak-rx').textContent = formatBps(peakRx, speedUnit);
      document.getElementById('graph-now-tx').textContent = formatBps(nowTx, speedUnit);
      document.getElementById('graph-now-rx').textContent = formatBps(nowRx, speedUnit);

      function toX(i) { return pad.left + (i / (len - 1)) * plotW; }
      function toY(v) { return pad.top + plotH - (v / maxVal) * plotH; }

      // Build SVG grid lines
      let yTicks = 4;
      let gridSvg = '';
      for (let i = 0; i <= yTicks; i++) {
        const v = (maxVal / yTicks) * i;
        const y = toY(v);
        gridSvg += `<line x1="${pad.left}" y1="${y}" x2="${W - pad.right}" y2="${y}" stroke="#21262d" stroke-width="1" stroke-dasharray="${i === 0 ? '0' : '4 4'}"/>`;
        gridSvg += `<text x="${pad.left - 10}" y="${y + 4}" text-anchor="end" fill="#b1bac4" font-size="11" font-family="'SF Mono', Monaco, monospace">${formatBps(v, speedUnit)}</text>`;
      }

      // X Axis Time ticks
      let xTicks = [];
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
        if (currentGraphRange === '24h') {
          timeStr = date.toLocaleTimeString([], {hour: '2-digit', minute: '2-digit'}) + '<br>' + (date.getMonth()+1) + '/' + date.getDate();
        }
        return `
          <line x1="${x}" y1="${pad.top}" x2="${x}" y2="${pad.top + plotH}" stroke="#21262d" stroke-dasharray="4 4" stroke-width="0.5"/>
          <text x="${x}" y="${H - 12}" text-anchor="middle" fill="#b1bac4" font-size="10">${timeStr}</text>
        `;
      }).join('');

      // Paths building
      let txD = '', rxD = '';
      let txAreaD = `M ${toX(0)} ${toY(0)} `;
      let rxAreaD = `M ${toX(0)} ${toY(0)} `;

      for (let i = 0; i < len; i++) {
        const x = toX(i).toFixed(1);
        const yTx = toY(txPoints[i]).toFixed(1);
        const yRx = toY(rxPoints[i]).toFixed(1);
        
        if (i === 0) {
          txD += `M ${x} ${yTx}`;
          rxD += `M ${x} ${yRx}`;
          txAreaD = `M ${x} ${pad.top + plotH} L ${x} ${yTx}`;
          rxAreaD = `M ${x} ${pad.top + plotH} L ${x} ${yRx}`;
        } else {
          txD += ` L ${x} ${yTx}`;
          rxD += ` L ${x} ${yRx}`;
          txAreaD += ` L ${x} ${yTx}`;
          rxAreaD += ` L ${x} ${yRx}`;
        }
      }
      txAreaD += ` L ${toX(len - 1).toFixed(1)} ${pad.top + plotH} Z`;
      rxAreaD += ` L ${toX(len - 1).toFixed(1)} ${pad.top + plotH} Z`;

      // Full HTML Render inside SVG
      svg.innerHTML = `
        <defs>
          <linearGradient id="tx-grad" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stop-color="#58a6ff" stop-opacity="0.35"/>
            <stop offset="100%" stop-color="#58a6ff" stop-opacity="0.0"/>
          </linearGradient>
          <linearGradient id="rx-grad" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stop-color="#3fb950" stop-opacity="0.35"/>
            <stop offset="100%" stop-color="#3fb950" stop-opacity="0.0"/>
          </linearGradient>
        </defs>
        
        <!-- Grid and Axes -->
        ${gridSvg}
        ${xTicksSvg}

        <!-- Filled Areas (Gradients) -->
        <path d="${txAreaD}" fill="url(#tx-grad)" />
        <path d="${rxAreaD}" fill="url(#rx-grad)" />

        <!-- Line curves -->
        <path d="${txD}" fill="none" stroke="#58a6ff" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"/>
        <path d="${rxD}" fill="none" stroke="#3fb950" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"/>
        
        <!-- Legend labels -->
        <g transform="translate(${W - pad.right - 90}, ${pad.top + 5})">
          <rect x="0" y="0" width="12" height="12" fill="#58a6ff" rx="3"/>
          <text x="18" y="10" fill="#b1bac4" font-size="11" font-weight="600">TX Speed</text>
        </g>
        <g transform="translate(${W - pad.right - 90}, ${pad.top + 22})">
          <rect x="0" y="0" width="12" height="12" fill="#3fb950" rx="3"/>
          <text x="18" y="10" fill="#b1bac4" font-size="11" font-weight="600">RX Speed</text>
        </g>

        <!-- Hover Guideline / Interactive Crosshair components -->
        <line id="hover-line" x1="0" y1="${pad.top}" x2="0" y2="${pad.top + plotH}" stroke="#b1bac4" stroke-width="1" stroke-dasharray="3 3" style="display:none; pointer-events:none;"/>
        <circle id="hover-dot-tx" r="5" fill="#58a6ff" stroke="#161b22" stroke-width="2" style="display:none; pointer-events:none;"/>
        <circle id="hover-dot-rx" r="5" fill="#3fb950" stroke="#161b22" stroke-width="2" style="display:none; pointer-events:none;"/>

        <!-- Invisible Overlay rect for hover detection -->
        <rect id="hover-overlay-rect" x="${pad.left}" y="${pad.top}" width="${plotW}" height="${plotH}" fill="transparent" style="cursor: crosshair;"/>
      `;

      // Set up mouse events on the overlay
      const rect = document.getElementById('hover-overlay-rect');
      const hoverLine = document.getElementById('hover-line');
      const hoverDotTx = document.getElementById('hover-dot-tx');
      const hoverDotRx = document.getElementById('hover-dot-rx');
      const tooltip = document.getElementById('graph-tooltip');

      rect.addEventListener('mousemove', (e) => {
        // Calculate index closest to mouse position
        const rectBounds = svg.getBoundingClientRect();
        const mouseX = e.clientX - rectBounds.left;
        
        // Translate client mouseX to SVG internal coordinate W
        const internalX = (mouseX / rectBounds.width) * W;
        const relativeX = internalX - pad.left;
        
        // Find index
        let idx = Math.round((relativeX / plotW) * (len - 1));
        if (idx < 0) idx = 0;
        if (idx >= len) idx = len - 1;

        const xPos = toX(idx);
        const yTx = toY(txPoints[idx]);
        const yRx = toY(rxPoints[idx]);

        // Draw guideline and dots
        hoverLine.setAttribute('x1', String(xPos));
        hoverLine.setAttribute('x2', String(xPos));
        hoverLine.style.display = 'block';

        hoverDotTx.setAttribute('cx', String(xPos));
        hoverDotTx.setAttribute('cy', String(yTx));
        hoverDotTx.style.display = 'block';

        hoverDotRx.setAttribute('cx', String(xPos));
        hoverDotRx.setAttribute('cy', String(yRx));
        hoverDotRx.style.display = 'block';

        // Position & Show Tooltip Panel
        const date = new Date(tss[idx] * 1000);
        let timeStr = date.toLocaleTimeString();
        if (currentGraphRange === '24h') {
          timeStr = date.toLocaleDateString() + ' ' + date.toLocaleTimeString([], {hour: '2-digit', minute:'2-digit'});
        }

        tooltip.innerHTML = `
          <div class="time">${timeStr}</div>
          <div class="tx">TX: ${formatBps(txPoints[idx], speedUnit)}</div>
          <div class="rx">RX: ${formatBps(rxPoints[idx], speedUnit)}</div>
        `;

        // Position tooltip relatively to the cursor position
        const svgContainerRect = svg.parentElement.getBoundingClientRect();
        const tipX = e.clientX - svgContainerRect.left + 15;
        const tipY = e.clientY - svgContainerRect.top - 60;
        
        tooltip.style.left = `${tipX}px`;
        tooltip.style.top = `${tipY}px`;
        tooltip.style.display = 'block';
      });

      rect.addEventListener('mouseleave', () => {
        hoverLine.style.display = 'none';
        hoverDotTx.style.display = 'none';
        hoverDotRx.style.display = 'none';
        tooltip.style.display = 'none';
      });
    })
    .catch(e => {
      svg.innerHTML = `<text x="${W/2}" y="${H/2}" text-anchor="middle" fill="#f85149" font-size="14">Error loading history: ${e.message}</text>`;
    });
}

