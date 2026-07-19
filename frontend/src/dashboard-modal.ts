// @ts-nocheck
import { REFRESH_SECONDS, currentGraphIp, currentGraphPort, currentGraphRange, currentGraphSwitchName, currentGraphPortLabel, currentTransceiverIp, currentTransceiverPort, currentTransceiverSwitchName, formatBytes, formatBps, speedUnit } from './dashboard-utils';

export function refreshTransceiver() {
  if (currentTransceiverIp && currentTransceiverPort) {
    openTransceiver(currentTransceiverIp, currentTransceiverPort, currentTransceiverSwitchName);
  }
}

export function openGraph(ip, port, name, label, totalTx = 0, totalRx = 0) {
  currentGraphIp = ip;
  currentGraphPort = port;
  currentGraphSwitchName = name;
  currentGraphPortLabel = label;
  document.getElementById('graph-title').textContent = `Bandwidth - ${label} (${name})`;
  document.getElementById('graph-stats').innerHTML = `
    <div class="info-item"><div class="label">Peak TX</div><div class="value" id="graph-peak-tx">-</div></div>
    <div class="info-item"><div class="label">Peak RX</div><div class="value" id="graph-peak-rx">-</div></div>
    <div class="info-item"><div class="label">Current TX</div><div class="value" id="graph-now-tx">-</div></div>
    <div class="info-item"><div class="label">Current RX</div><div class="value" id="graph-now-rx">-</div></div>
    <div class="info-item"><div class="label">Total TX</div><div class="value" id="graph-total-tx">${formatBytes(totalTx)}</div></div>
    <div class="info-item"><div class="label">Total RX</div><div class="value" id="graph-total-rx">${formatBytes(totalRx)}</div></div>`;
  document.getElementById('graph-sub').textContent = `Cumulative traffic for port ${port} - updated every ${REFRESH_SECONDS}s`;
  document.getElementById('graph-overlay').classList.add('active');
  renderGraph(ip, port);
}

export function setGraphRange(range) {
  currentGraphRange = range;
  document.querySelectorAll('.graph-tab').forEach(t => t.classList.remove('active'));
  const tab = document.querySelector(`.graph-tab[data-range="${range}"]`);
  if (tab) tab.classList.add('active');
  if (currentGraphIp && currentGraphPort) renderGraph(currentGraphIp, currentGraphPort);
}

export function closeGraph() {
  document.getElementById('graph-overlay').classList.remove('active');
}

export function toggleMaximizeGraph() {
  const box = document.getElementById('graph-box');
  const btn = document.getElementById('graph-maximize-btn');
  if (box.style.maxWidth === 'none') {
    box.style.maxWidth = '700px';
    btn.textContent = '+';
  } else {
    box.style.maxWidth = 'none';
    box.style.width = '95vw';
    btn.textContent = '-';
  }
}

export function renderGraph(ip, port) {
  const range = currentGraphRange;
  const svg = document.getElementById('graph-svg');
  const W = svg.clientWidth || 600;
  const H = 200;
  const pad = { top: 10, right: 10, bottom: 20, left: 50 };
  const plotW = W - pad.left - pad.right;
  const plotH = H - pad.top - pad.bottom;

  fetch(`/api/history?ip=${ip}&port=${port}&range=${range}`)
    .then(r => r.json())
    .then(data => {
      const tx = data.tx || [];
      const rx = data.rx || [];
      const ts = data.timestamps || [];

      let html = `<rect width="${W}" height="${H}" fill="#0d1117" rx="8"/>
        <g transform="translate(${pad.left},${pad.top})">
        <line x1="0" y1="0" x2="0" y2="${plotH}" stroke="#30363d" stroke-width="1"/>
        <line x1="0" y1="${plotH}" x2="${plotW}" y2="${plotH}" stroke="#30363d" stroke-width="1"/>`;

      if (!tx.length) {
        html += `<text x="${plotW/2}" y="${plotH/2}" text-anchor="middle" fill="#b1bac4" font-size="14">No history data available for this range yet</text>`;
        svg.innerHTML = html + '</g></svg>';
        return;
      }

      const allVals = [...tx, ...rx].filter(v => v > 0);
      const maxVal = allVals.length ? Math.max(...allVals) * 1.1 : 1;
      const stepX = plotW / (ts.length - 1 || 1);

      // Grid lines
      for (let i = 0; i <= 4; i++) {
        const y = plotH - (plotH * i / 4);
        html += `<line x1="0" y1="${y}" x2="${plotW}" y2="${y}" stroke="#21262d" stroke-width="1"/>`;
        html += `<text x="-8" y="${y+3}" text-anchor="end" fill="#8b949e" font-size="9">${formatBps(maxVal * i / 4, speedUnit)}</text>`;
      }

      // TX line
      let txPath = '';
      let rxPath = '';
      for (let i = 0; i < ts.length; i++) {
        const x = i * stepX;
        const txy = plotH - (tx[i] / maxVal) * plotH;
        const rxy = plotH - (rx[i] / maxVal) * plotH;
        txPath += (i === 0 ? 'M' : 'L') + x + ',' + txy;
        rxPath += (i === 0 ? 'M' : 'L') + x + ',' + rxy;
      }

      html += `<path d="${txPath}" fill="none" stroke="#58a6ff" stroke-width="2"/>`;
      html += `<path d="${rxPath}" fill="none" stroke="#3fb950" stroke-width="2"/>`;

      // Area fill
      html += `<path d="${txPath} L${(ts.length-1)*stepX},${plotH} L0,${plotH} Z" fill="rgba(88,166,255,0.08)"/>`;
      html += `<path d="${rxPath} L${(ts.length-1)*stepX},${plotH} L0,${plotH} Z" fill="rgba(63,185,80,0.08)"/>`;

      html += '</g>';

      // Stats
      const peakTX = Math.max(...tx);
      const peakRX = Math.max(...rx);
      const nowTX = tx[tx.length - 1] || 0;
      const nowRX = rx[rx.length - 1] || 0;

      document.getElementById('graph-peak-tx').textContent = formatBps(peakTX, speedUnit);
      document.getElementById('graph-peak-rx').textContent = formatBps(peakRX, speedUnit);
      document.getElementById('graph-now-tx').textContent = formatBps(nowTX, speedUnit);
      document.getElementById('graph-now-rx').textContent = formatBps(nowRX, speedUnit);

      svg.innerHTML = html;
    })
    .catch(() => {
      svg.innerHTML = `<rect width="${W}" height="${H}" fill="#0d1117" rx="8"/>
        <text x="${W/2}" y="${H/2}" text-anchor="middle" fill="#f85149" font-size="14">Failed to load history</text>`;
    });
}

export function closeTransceiver() {
  document.getElementById('transceiver-overlay').classList.remove('open');
}

export function openTransceiver(ip, port, switchName) {
  currentTransceiverIp = ip;
  currentTransceiverPort = port;
  currentTransceiverSwitchName = switchName;

  const overlay = document.getElementById('transceiver-overlay');
  document.getElementById('transceiver-title').textContent = `Port ${port} SFP+ Transceiver Status`;
  document.getElementById('transceiver-sub').textContent = `DDMI Diagnostics & Telemetry (${switchName})`;
  document.getElementById('transceiver-content').innerHTML = `<div style="display:flex;flex-direction:column;align-items:center;padding:40px;gap:12px;color:#8b949e;"><div class="spinner"></div><span>Loading transceiver data...</span></div>`;
  overlay.classList.add('open');

  fetch(`/api/switches/${ip}/transceiver`)
    .then(r => { if (!r.ok) throw new Error('Request failed'); return r.json(); })
    .then(data => renderTransceiverData(data, document.getElementById('transceiver-content')))
    .catch(err => {
      document.getElementById('transceiver-content').innerHTML =
        `<div style="margin:20px 0;font-size:13px;color:#8b949e;"><strong>No SFP module detected</strong><br><span style="opacity:0.8;font-size:11px;">${err.message}</span></div>`;
    });
}

export function renderTransceiverData(data, container) {
  function parseVal(valStr) {
    if (!valStr) return 0;
    const m = valStr.toString().match(/[-+]?[0-9]*\.?[0-9]+/);
    return m ? parseFloat(m[0]) : 0;
  }

  const temp = parseVal(data.temperature);
  const volt = parseVal(data.voltage);
  const curr = parseVal(data.current);
  let txDbm = -99, rxDbm = -99;
  if (data.tx_power && !data.tx_power.includes('-inf')) txDbm = parseVal(data.tx_power);
  if (data.rx_power && !data.rx_power.includes('-inf')) rxDbm = parseVal(data.rx_power);

  const bar = (val, max, color) =>
    `<div class="telemetry-bar"><div class="telemetry-fill" style="width:${Math.min(100, Math.max(0, (val/max)*100))}%;background:${color}"></div></div>`;

  container.innerHTML = `
    <div class="sfp-grid">
      <div class="sfp-item"><div class="label">Vendor</div><div class="value">${data.vendor_name || '-'}</div></div>
      <div class="sfp-item"><div class="label">Model</div><div class="value">${data.vendor_pn || '-'}</div></div>
      <div class="sfp-item"><div class="label">Serial</div><div class="value" style="font-size:11px;">${data.vendor_sn || '-'}</div></div>
      <div class="sfp-item"><div class="label">Type</div><div class="value">${data.transceiver_type || '-'}</div></div>
      <div class="sfp-item"><div class="label">Temperature</div><div class="value" style="color:${temp > 65 ? '#f85149' : temp > 55 ? '#d29922' : '#3fb950'}">${data.temperature || '-'}</div>${bar(temp, 80, temp > 65 ? '#f85149' : temp > 55 ? '#d29922' : '#3fb950')}</div>
      <div class="sfp-item"><div class="label">Voltage</div><div class="value" style="color:${volt > 3.5 || volt < 3.1 ? '#f85149' : '#3fb950'}">${data.voltage || '-'}</div>${bar(volt - 2.8, 1, volt > 3.5 || volt < 3.1 ? '#f85149' : '#3fb950')}</div>
      <div class="sfp-item"><div class="label">Bias Current</div><div class="value" style="color:${curr > 30 ? '#f85149' : '#3fb950'}">${data.current || '-'}</div>${bar(curr, 40, curr > 30 ? '#f85149' : '#3fb950')}</div>
      <div class="sfp-item"><div class="label">TX Power</div><div class="value" style="color:${txDbm > 3 || txDbm < -10 ? '#f85149' : '#3fb950'}">${data.tx_power || '-'}</div>${bar(txDbm + 10, 15, txDbm > 3 || txDbm < -10 ? '#f85149' : '#3fb950')}</div>
      <div class="sfp-item"><div class="label">RX Power</div><div class="value" style="color:${rxDbm > 3 || rxDbm < -15 ? '#f85149' : '#3fb950'}">${data.rx_power || '-'}</div>${bar(rxDbm + 15, 20, rxDbm > 3 || rxDbm < -15 ? '#f85149' : '#3fb950')}</div>
    </div>
    <p style="font-size:10px;color:#8b949e;margin-top:12px;">DDMI telemetry reads directly from SFP module registers.</p>`;
}
