let currentTransceiverIp: string | null = null;
let currentTransceiverPort: string | null = null;
let currentTransceiverSwitchName = '';

/* ---- Transceiver ---- */
export function closeTransceiver() {
  document.getElementById('transceiver-overlay').classList.remove('open');
}

export function openTransceiver(ip, port, switchName) {
  currentTransceiverIp = ip;
  currentTransceiverPort = port;
  currentTransceiverSwitchName = switchName;

  const overlay = document.getElementById('transceiver-overlay');
  const title = document.getElementById('transceiver-title');
  const sub = document.getElementById('transceiver-sub');
  const content = document.getElementById('transceiver-content');
  
  title.textContent = `Port ${port} SFP+ Transceiver Status`;
  sub.textContent = `DDMI Diagnostics & Telemetry (${switchName})`;
  
  // Show loading state
  content.innerHTML = `
    <div style="display: flex; flex-direction: column; align-items: center; justify-content: center; padding: 40px; gap: 12px; color: #8b949e;">
      <div style="width: 24px; height: 24px; border: 2px solid transparent; border-top-color: #58a6ff; border-radius: 50%; animation: spin .8s linear infinite; box-sizing: border-box;"></div>
      <span>Fetching transceiver data from switch...</span>
    </div>
  `;
  
  overlay.classList.add('open');
  
  fetch(`/api/switches/${ip}/transceiver`)
    .then(r => {
      if (!r.ok) {
        return r.json().then(err => { throw new Error(err.error || 'Request failed'); });
      }
      return r.json();
    })
    .then(data => {
      renderTransceiverData(data, content);
    })
    .catch(err => {
      content.innerHTML = `
        <div class="error-card" style="margin: 20px 0; font-size: 13px;">
          <strong>No SFP module detected or telemetry unavailable</strong><br>
          <span style="opacity: 0.8; font-size: 11px;">${err.message}</span>
        </div>
      `;
    });
}

export function renderTransceiverData(data, container) {
  // Parsing helpers for telemetry bars
  function parseVal(valStr) {
    if (!valStr) return 0;
    const match = valStr.toString().match(/[-+]?[0-9]*\.?[0-9]+/);
    return match ? parseFloat(match[0]) : 0;
  }
  
  const temp = parseVal(data.temperature);
  const volt = parseVal(data.voltage);
  const curr = parseVal(data.current);
  
  let txDbm = -99;
  let rxDbm = -99;
  if (data.tx_power && !data.tx_power.includes('-inf')) {
    txDbm = parseVal(data.tx_power);
  }
  if (data.rx_power && !data.rx_power.includes('-inf')) {
    rxDbm = parseVal(data.rx_power);
  }

  // Calculate percentages for telemetry bars
  // Temperature range: 0 to 80 C
  const tempPct = Math.min(100, Math.max(0, (temp / 80) * 100));
  // Voltage range: 3.0V to 3.6V (centered around 3.3V)
  const voltPct = Math.min(100, Math.max(0, ((volt - 2.8) / (3.8 - 2.8)) * 100));
  // Current range: 0 to 40 mA (typical is 5-20 mA)
  const currPct = Math.min(100, Math.max(0, (curr / 40) * 100));
  // TX Power range: -20 dBm to +5 dBm
  const txPct = txDbm === -99 ? 0 : Math.min(100, Math.max(0, ((txDbm - (-20)) / (5 - (-20))) * 100));
  // RX Power range: -30 dBm to +5 dBm
  const rxPct = rxDbm === -99 ? 0 : Math.min(100, Math.max(0, ((rxDbm - (-30)) / (5 - (-30))) * 100));

  // Visual status colors
  const tempColor = temp > 65 || temp < 5 ? '#f85149' : (temp > 55 ? '#d29922' : '#3fb950');
  const voltColor = volt > 3.5 || volt < 3.1 ? '#f85149' : '#3fb950';
  const currColor = curr > 30 ? '#f85149' : '#3fb950';
  const txColor = txDbm === -99 ? '#57606a' : (txDbm < -12 ? '#d29922' : '#3fb950');
  const rxColor = rxDbm === -99 ? '#57606a' : (rxDbm < -18 ? '#d29922' : '#3fb950');

  container.innerHTML = `
    <div class="transceiver-grid">
      <!-- SFP Module Details Card -->
      <div class="transceiver-card">
        <h4>
          <svg viewBox="0 0 24 24" style="width: 14px; height: 14px; fill: currentColor; display: inline-block; vertical-align: middle;"><path d="M4,5 h3 v3 h10 v -3 h 3 v 14 h -16 z" /></svg>
          SFP Module Details
        </h4>
        <div class="transceiver-item"><span class="label">OE Present</span><span class="val">${data.oe_present || 'Unknown'}</span></div>
        <div class="transceiver-item"><span class="label">Loss of Signal</span><span class="val">${data.loss_of_signal || 'Unknown'}</span></div>
        <div class="transceiver-item"><span class="label">Transceiver Type</span><span class="val">${data.transceiver_type || 'Unknown'}</span></div>
        <div class="transceiver-item"><span class="label">Connector Type</span><span class="val">${data.connector_type || 'Unknown'}</span></div>
        <div class="transceiver-item"><span class="label">Compliance Codes</span><span class="val">${data.eth_compliance_codes || 'Unknown'}</span></div>
        <div class="transceiver-item"><span class="label">Wavelength</span><span class="val">${data.wavelength || 'Unknown'}</span></div>
        <div class="transceiver-item"><span class="label">Bitrate</span><span class="val">${data.bitrate || 'Unknown'}</span></div>
        <div class="transceiver-item"><span class="label">Vendor Name</span><span class="val">${(data.vendor_name || 'Unknown').trim()}</span></div>
        <div class="transceiver-item"><span class="label">Vendor OUI</span><span class="val">${data.vendor_oui || 'Unknown'}</span></div>
        <div class="transceiver-item"><span class="label">Part Number</span><span class="val">${(data.vendor_pn || 'Unknown').trim()}</span></div>
        <div class="transceiver-item"><span class="label">Revision</span><span class="val">${(data.vendor_revision || 'Unknown').trim()}</span></div>
        <div class="transceiver-item"><span class="label">Serial Number</span><span class="val">${(data.vendor_sn || 'Unknown').trim()}</span></div>
        <div class="transceiver-item"><span class="label">Date Code</span><span class="val">${data.date_code || 'Unknown'}</span></div>
      </div>
      
      <!-- DDMI Real-Time Telemetry Card -->
      <div class="transceiver-card">
        <h4>
          <svg viewBox="0 0 24 24" style="width: 14px; height: 14px; fill: currentColor; display: inline-block; vertical-align: middle;"><path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-6h-2v-2h4v8zm0-10h-2V5h2v2z"/></svg>
          DDMI Diagnostics
        </h4>
        
        <!-- Temperature -->
        <div class="telemetry-row">
          <div class="header-row">
            <span class="label">🌡️ Temperature</span>
            <span class="val" style="color: ${tempColor};">${data.temperature || 'N/A'}</span>
          </div>
          <div class="telemetry-bar-bg">
            <div class="telemetry-bar-fill" style="width: ${tempPct}%; background: ${tempColor};"></div>
          </div>
        </div>
        
        <!-- Voltage -->
        <div class="telemetry-row">
          <div class="header-row">
            <span class="label">⚡ Supply Voltage</span>
            <span class="val" style="color: ${voltColor};">${data.voltage || 'N/A'}</span>
          </div>
          <div class="telemetry-bar-bg">
            <div class="telemetry-bar-fill" style="width: ${voltPct}%; background: ${voltColor};"></div>
          </div>
        </div>
        
        <!-- Current -->
        <div class="telemetry-row">
          <div class="header-row">
            <span class="label">🔋 Bias Current</span>
            <span class="val" style="color: ${currColor};">${data.current || 'N/A'}</span>
          </div>
          <div class="telemetry-bar-bg">
            <div class="telemetry-bar-fill" style="width: ${currPct}%; background: ${currColor};"></div>
          </div>
        </div>
        
        <!-- TX Power -->
        <div class="telemetry-row">
          <div class="header-row">
            <span class="label">📤 Optical TX Power</span>
            <span class="val" style="color: ${txColor};">${data.tx_power || 'N/A'}</span>
          </div>
          <div class="telemetry-bar-bg">
            <div class="telemetry-bar-fill" style="width: ${txPct}%; background: ${txColor};"></div>
          </div>
        </div>
        
        <!-- RX Power -->
        <div class="telemetry-row">
          <div class="header-row">
            <span class="label">📥 Optical RX Power</span>
            <span class="val" style="color: ${rxColor};">${data.rx_power || 'N/A'}</span>
          </div>
          <div class="telemetry-bar-bg">
            <div class="telemetry-bar-fill" style="width: ${rxPct}%; background: ${rxColor};"></div>
          </div>
        </div>
        
        <div style="margin-top: 18px; padding: 10px; background: rgba(88,166,255,0.05); border: 1px solid rgba(88,166,255,0.15); border-radius: 6px; font-size: 11px; color: #8b949e; line-height: 1.4;">
          💡 Telemetry values are updated directly from the switch's internal SFP micro-controller registers using DDMI.
        </div>
      </div>
    </div>
  `;
}

