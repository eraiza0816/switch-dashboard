import { enabledColumns, PORTS_WRAP_THRESHOLD, formatBytes, formatBps, formatPkts, speedClass, formatTime } from './dashboard-utils';
import { speedUnit } from './dashboard-graph';

let refreshDashboard: () => void = () => {};
export function setRefreshDashboard(fn: () => void) {
  refreshDashboard = fn;
}

const defaultOrder = ['port', 'status', 'speed', 'packets', 'bytes', 'info', 'notes'];
let columnOrder = [];

const serverColumnWidths = {};
const serverColumnOrder = [];

export function initColumnOrder() {
  let savedOrder = serverColumnOrder && serverColumnOrder.length > 0 ? serverColumnOrder : null;
  
  if (!savedOrder) {
    savedOrder = defaultOrder;
  }
  
  // Filter savedOrder to keep only enabled columns
  columnOrder = savedOrder.filter(id => enabledColumns.includes(id));
  
  // Append any enabled columns that are not in columnOrder
  enabledColumns.forEach(id => {
    if (!columnOrder.includes(id)) {
      columnOrder.push(id);
    }
  });
}

// Initialize Column Order
initColumnOrder();

let columnWidths = serverColumnWidths || {};

async function saveColumnSettings() {
  try {
    await fetch('/api/config/settings', {
      method: 'POST',
      headers: {'Content-Type': 'application/json'},
      body: JSON.stringify({
        column_widths: columnWidths,
        column_order: columnOrder
      })
    });
  } catch (e) {
    console.error("Failed to save column settings to server:", e);
  }
}

function saveColumnWidth(colId, width) {
  columnWidths[colId] = width;
  saveColumnSettings();
}

function loadColumnWidths() {
  columnWidths = serverColumnWidths || {};
}

loadColumnWidths();

// Resize Columns Logic
let resizingColId = null;
let resizingStartX = 0;
let resizingStartWidth = 0;
let resizingTh = null;

export function handleResizeStart(event, colId) {
  event.stopPropagation();
  event.preventDefault();
  resizingColId = colId;
  resizingTh = event.target.closest('th');
  resizingStartX = event.clientX;
  resizingStartWidth = resizingTh.offsetWidth;
  
  document.addEventListener('mousemove', handleResizeMove);
  document.addEventListener('mouseup', handleResizeEnd);
  resizingTh.classList.add('resizing');
  
  // Temporarily disable draggable on all headers so browser doesn't trigger drag-and-drop
  document.querySelectorAll('.port-table th, .mac-table th').forEach(th => {
    th.setAttribute('draggable', 'false');
  });
}

function handleResizeMove(event) {
  if (!resizingColId || !resizingTh) return;
  const diff = event.clientX - resizingStartX;
  const newWidth = Math.max(30, resizingStartWidth + diff);
  
  if (resizingColId.startsWith('mac-')) {
    // Resize all MAC tables across all switch cards to keep layout aligned
    const ths = document.querySelectorAll(`.mac-table th[data-col-id="${resizingColId}"]`);
    ths.forEach(th => {
      (th as HTMLElement).style.width = newWidth + 'px';
    });
  } else {
    // Set the width on all corresponding th elements across all switch cards to keep layout aligned
    const ths = document.querySelectorAll(`.port-table th[data-col-id="${resizingColId}"]`);
    ths.forEach(th => {
      (th as HTMLElement).style.width = newWidth + 'px';
    });
  }
}

function handleResizeEnd(event) {
  if (resizingColId && resizingTh) {
    resizingTh.classList.remove('resizing');
    const finalWidth = resizingTh.offsetWidth;
    saveColumnWidth(resizingColId, finalWidth);
  }
  
  resizingColId = null;
  resizingTh = null;
  document.removeEventListener('mousemove', handleResizeMove);
  document.removeEventListener('mouseup', handleResizeEnd);
  
  // Restore draggable on all headers
  document.querySelectorAll('.port-table th, .mac-table th').forEach(th => {
    th.setAttribute('draggable', 'true');
  });
}

const COLUMN_DEFS = {
  port: {
    label: 'Port',
    thStyle: 'width:55px',
    tdClass: 'port-num',
    tdTitleFn: (p) => {
      let parts = [];
      if (p.vm_name) parts.push(p.vm_name);
      if (p.interface) parts.push(p.interface);
      return parts.length > 0 ? parts.join(' / ') : '';
    },
    render: (p, sw) => p.port
  },
  status: {
    label: 'Status',
    thStyle: 'width:145px',
    render: (p, sw) => {
      const status = (p.status || 'unknown').toLowerCase();
      // DHCP Snooping port trust badges
      let trustBadgeHtml = '';
      if (sw.dhcp_snooping && sw.dhcp_snooping.enabled) {
        const portsMap = sw.dhcp_snooping.ports || {};
        const cleanPortName = p.port.split('/')[0];
        const trustState = portsMap[p.port] || portsMap[cleanPortName];
        if (trustState === 'Trusted') {
          trustBadgeHtml = `<span class="trust-badge trusted" title="DHCP Snooping: Trusted Port">Trusted</span>`;
        } else if (trustState === 'Untrusted') {
          trustBadgeHtml = `<span class="trust-badge untrusted" title="DHCP Snooping: Untrusted Port">Untrusted</span>`;
        }
      }
      let duplexText = p.duplex || 'Auto';
      if (duplexText.toLowerCase() === 'full') {
        duplexText = 'Full Duplex';
      } else if (duplexText.toLowerCase() === 'half') {
        duplexText = 'Half Duplex';
      }
      return `
        <div style="display: flex; flex-direction: column; gap: 4px; align-items: flex-start;">
          <div style="display: flex; align-items: center; gap: 6px;">
            <span class="status-badge ${status}" onclick="openGraph('${sw.ip}','${p.port}','${sw.name}','Port ${p.port}', ${p.cum_tx || p.tx_bytes || 0}, ${p.cum_rx || p.rx_bytes || 0})"><span class="status-dot ${status}"></span>${status.toUpperCase()}</span>
            ${trustBadgeHtml}
          </div>
          <div style="font-size: 10px; color: #8b949e; padding-left: 4px; font-weight: 500;">${duplexText}</div>
        </div>
      `;
    }
  },
  speed: {
    label: 'Speed',
    thStyle: 'width:55px',
    tdClassFn: (p) => speedClass(p.speed, p.status),
    tdTitleFn: (p) => p.duplex || '',
    render: (p, sw) => {
      const isSfp = p.is_sfp;
      const speed = p.speed || 'Auto';
      if (isSfp) {
        return `<span class="sfp-speed-link" onclick="openTransceiver('${sw.ip}','${p.port}','${sw.name}')" title="Click to view SFP+ Transceiver Diagnostics">${speed} <svg viewBox="0 0 24 24" style="width: 10px; height: 10px; fill: currentColor; display: inline-block;"><path d="M12 9a3 3 0 1 0 0 6 3 3 0 0 0 0-6zm0-7C7 2 2.73 5.11 1 9.5 2.73 13.89 7 17 12 17s9.27-3.11 11-7.5C21.27 5.11 17 2 12 2zm0 13a5.5 5.5 0 1 1 0-11 5.5 5.5 0 0 1 0 11z"/></svg></span>`;
      } else {
        return speed;
      }
    }
  },
  packets: {
    label: 'Traffic<br><span style="font-weight:400;color:#8b949e">(packets)</span>',
    thStyle: 'width:90px',
    tdClass: 'mono',
    tdTitle: 'TX / RX packets',
    render: (p, sw) => {
      const tx = p.tx_packets || 0;
      const rx = p.rx_packets || 0;
      return `TX ${formatPkts(tx)}<br>RX ${formatPkts(rx)}`;
    }
  },
  bytes: {
    label: 'Traffic<br><span style="font-weight:400;color:#8b949e">(bytes)</span>',
    thStyle: 'width:90px',
    tdClass: 'mono',
    tdTitle: 'TX / RX bytes formatted',
    render: (p, sw) => {
      const cumTx = p.cum_tx || p.tx_bytes || 0;
      const cumRx = p.cum_rx || p.rx_bytes || 0;
      return `TX ${formatBytes(cumTx)}<br>RX ${formatBytes(cumRx)}`;
    }
  },
  info: {
    getLabel: () => `<span id="speed-unit-toggle" class="unit-toggle" onclick="toggleSpeedUnit(event)" title="Toggle speed unit">${speedUnit === 'bps' ? 'BPS' : 'B/s'}</span>`,
    tdClass: 'col-info mono',
    render: (p, sw) => {
      const tx = p.speed_tx_bps || 0;
      const rx = p.speed_rx_bps || 0;
      return `TX ${formatBps(tx, speedUnit)}<br>RX ${formatBps(rx, speedUnit)}`;
    }
  },
  notes: {
    label: 'Notes',
    tdClass: 'col-note',
    render: (p, sw) => {
      return `<input class="note-input" type="text" placeholder="-" value="${p.note || ''}" data-key="${sw.ip}:${p.port}" onfocus="pausePolling()" onblur="saveNote(this)">`;
    }
  }
};

// Drag and drop variables and handlers
let draggedColumnId = null;

export function handleDragStart(event, colId) {
  draggedColumnId = colId;
  event.dataTransfer.effectAllowed = 'move';
  event.dataTransfer.setData('text/plain', colId);
}

export function handleDragOver(event) {
  event.preventDefault();
  event.dataTransfer.dropEffect = 'move';
}

export function handleDragEnter(event, colId) {
  event.preventDefault();
  const th = event.target.closest('th');
  if (th && draggedColumnId !== colId) {
    th.classList.add('drag-over');
  }
}

export function handleDragLeave(event) {
  const th = event.target.closest('th');
  if (th) {
    th.classList.remove('drag-over');
  }
}

export function handleDragEnd(event) {
  document.querySelectorAll('.port-table th').forEach(th => {
    th.classList.remove('drag-over');
  });
}

export function handleDrop(event, targetColId) {
  event.preventDefault();
  const th = event.target.closest('th');
  if (th) {
    th.classList.remove('drag-over');
  }
  if (draggedColumnId && draggedColumnId !== targetColId) {
    const fromIndex = columnOrder.indexOf(draggedColumnId);
    const toIndex = columnOrder.indexOf(targetColId);
    if (fromIndex !== -1 && toIndex !== -1) {
      columnOrder.splice(fromIndex, 1);
      columnOrder.splice(toIndex, 0, draggedColumnId);
      saveColumnSettings();
      refreshDashboard();
    }
  }
}
let macTableStates = {};
try {
  const stored = localStorage.getItem('macTableStates');
  if (stored) {
    macTableStates = JSON.parse(stored);
  }
} catch (e) {
  console.error("Failed to load macTableStates from localStorage", e);
}

export function saveMacStatesToStorage() {
  const prunedStates = {};
  for (const ip in macTableStates) {
    if (macTableStates.hasOwnProperty(ip)) {
      const s = macTableStates[ip];
      prunedStates[ip] = {
        expanded: s.expanded,
        filters: s.filters || { mac: '', vendor: '', type: '', port: '', vlan: '' },
        sortBy: s.sortBy || 'port',
        sortAsc: s.sortAsc !== undefined ? s.sortAsc : true,
        scrollTop: s.scrollTop || 0,
        height: s.height || '450px'
      };
    }
  }
  localStorage.setItem('macTableStates', JSON.stringify(prunedStates));
}

export function getMacState(ip) {
  if (!macTableStates[ip]) {
    macTableStates[ip] = {};
  }
  const state = macTableStates[ip];
  if (state.expanded === undefined) state.expanded = false;
  if (!state.filters) {
    state.filters = {
      mac: '',
      host: '',
      vendor: '',
      type: '',
      port: '',
      vlan: ''
    };
  }
  if (state.sortBy === undefined) state.sortBy = 'port';
  if (state.sortAsc === undefined) state.sortAsc = true;
  if (!state.data) state.data = [];
  if (state.scrollTop === undefined) state.scrollTop = 0;
  if (state.height === undefined) state.height = '450px';
  return state;
}

function parsePortForSort(portStr) {
  if (!portStr) return 0;
  if (!isNaN(portStr)) return parseInt(portStr);
  const m = portStr.match(/\d+/);
  if (m) return parseInt(m[0]);
  return portStr.toString().toLowerCase();
}

function parseVlanForSort(vlanStr) {
  if (!vlanStr) return 0;
  const parsed = parseInt(vlanStr);
  return isNaN(parsed) ? vlanStr.toString().toLowerCase() : parsed;
}
let scrollSaveTimeout = null;
export function saveMacScroll(ip, element) {
  const state = getMacState(ip);
  state.scrollTop = element.scrollTop;
  
  clearTimeout(scrollSaveTimeout);
  scrollSaveTimeout = setTimeout(() => {
    saveMacStatesToStorage();
  }, 500);
}

export function saveMacHeight(ip, element) {
  if (element && element.style.height) {
    const state = getMacState(ip);
    state.height = element.style.height;
    saveMacStatesToStorage();
  }
}

export function toggleMacTable(ip) {
  const state = getMacState(ip);
  state.expanded = !state.expanded;
  saveMacStatesToStorage();
  
  const content = document.querySelector(`.mac-content-${CSS.escape(ip)}`) as HTMLElement | null;
  const arrow = document.querySelector(`.mac-toggle-arrow-${CSS.escape(ip)}`) as HTMLElement | null;
  
  if (content && arrow) {
    if (state.expanded) {
      content.style.display = 'block';
      arrow.style.transform = 'rotate(90deg)';
      renderMacTable(ip);
    } else {
      content.style.display = 'none';
      arrow.style.transform = 'rotate(0deg)';
    }
  }
}

function getIgmpState(ip) {
  let states = {};
  try {
    const stored = localStorage.getItem('igmp_states');
    if (stored) states = JSON.parse(stored);
  } catch (e) {
    console.error("Failed to load igmp_states from localStorage", e);
  }
  if (!states[ip]) {
    states[ip] = { expanded: false };
  }
  return states[ip];
}

export function toggleIgmpTable(ip) {
  let states = {};
  try {
    const stored = localStorage.getItem('igmp_states');
    if (stored) states = JSON.parse(stored);
  } catch (e) {
    console.error("Failed to load igmp_states from localStorage", e);
  }
  if (!states[ip]) states[ip] = { expanded: false };
  states[ip].expanded = !states[ip].expanded;
  localStorage.setItem('igmp_states', JSON.stringify(states));
  
  const content = document.querySelector(`.igmp-content-${CSS.escape(ip)}`) as HTMLElement | null;
  const arrow = document.querySelector(`.igmp-toggle-arrow-${CSS.escape(ip)}`) as HTMLElement | null;
  
  if (content && arrow) {
    if (states[ip].expanded) {
      content.style.display = 'block';
      arrow.style.transform = 'rotate(90deg)';
    } else {
      content.style.display = 'none';
      arrow.style.transform = 'rotate(0deg)';
    }
  }
}

const EXTRA_SECTIONS = ['eee', 'vlan', 'lag', 'mirror', 'bandwidth'] as const;
type ExtraSection = typeof EXTRA_SECTIONS[number];

function getExtraSectionState(ip: string, section: ExtraSection) {
  let states: Record<string, Record<string, { expanded: boolean }>> = {};
  try {
    const stored = localStorage.getItem('extra_section_states');
    if (stored) states = JSON.parse(stored);
  } catch (e) {
    console.error("Failed to load extra_section_states from localStorage", e);
  }
  if (!states[ip]) states[ip] = {};
  if (!states[ip][section]) states[ip][section] = { expanded: false };
  return states[ip][section];
}

export function toggleExtraSection(ip: string, section: ExtraSection) {
  let states: Record<string, Record<string, { expanded: boolean }>> = {};
  try {
    const stored = localStorage.getItem('extra_section_states');
    if (stored) states = JSON.parse(stored);
  } catch (e) {
    console.error("Failed to load extra_section_states from localStorage", e);
  }
  if (!states[ip]) states[ip] = {};
  if (!states[ip][section]) states[ip][section] = { expanded: false };
  states[ip][section].expanded = !states[ip][section].expanded;
  localStorage.setItem('extra_section_states', JSON.stringify(states));

  const content = document.querySelector(`.${section}-content-${CSS.escape(ip)}`) as HTMLElement | null;
  const arrow = document.querySelector(`.${section}-toggle-arrow-${CSS.escape(ip)}`) as HTMLElement | null;
  if (content && arrow) {
    content.style.display = states[ip][section].expanded ? 'block' : 'none';
    arrow.style.transform = states[ip][section].expanded ? 'rotate(90deg)' : 'rotate(0deg)';
  }
}
export function filterMacTable(ip) {
  const state = getMacState(ip);
  const fMac = document.querySelector(`.mac-filter-mac-${CSS.escape(ip)}`) as HTMLInputElement | null;
  const fHost = document.querySelector(`.mac-filter-host-${CSS.escape(ip)}`) as HTMLInputElement | null;
  const fVendor = document.querySelector(`.mac-filter-vendor-${CSS.escape(ip)}`) as HTMLInputElement | null;
  const fType = document.querySelector(`.mac-filter-type-${CSS.escape(ip)}`) as HTMLInputElement | null;
  const fPort = document.querySelector(`.mac-filter-port-${CSS.escape(ip)}`) as HTMLInputElement | null;
  const fVlan = document.querySelector(`.mac-filter-vlan-${CSS.escape(ip)}`) as HTMLInputElement | null;
  
  if (fMac) state.filters.mac = fMac.value.trim().toLowerCase();
  if (fHost) state.filters.host = fHost.value.trim().toLowerCase();
  if (fVendor) state.filters.vendor = fVendor.value.trim().toLowerCase();
  if (fType) state.filters.type = fType.value.trim().toLowerCase();
  if (fPort) state.filters.port = fPort.value.trim().toLowerCase();
  if (fVlan) state.filters.vlan = fVlan.value.trim().toLowerCase();
  
  saveMacStatesToStorage();
  renderMacTable(ip);
}
export function sortMac(ip, col) {
  const state = getMacState(ip);
  if (state.sortBy === col) {
    state.sortAsc = !state.sortAsc;
  } else {
    state.sortBy = col;
    state.sortAsc = true;
  }
  saveMacStatesToStorage();
  renderMacTable(ip);
}
export function manualRefreshMac(ip) {
  const state = getMacState(ip);
  const spinner = document.querySelector(`.mac-spinner-${CSS.escape(ip)}`);
  const btnText = document.querySelector(`.mac-refresh-btn-${CSS.escape(ip)} span:not(.mac-spinner-${CSS.escape(ip)})`);
  
  if (spinner) (spinner as HTMLElement).style.display = 'inline-block';
  if (btnText) btnText.textContent = 'Refreshing...';
  
  fetch(`/api/switches/${ip}/refresh_mac`, { method: 'POST' })
    .then(r => r.json())
    .then(data => {
      if (data.status === 'ok') {
        state.data = data.mac_table || [];
        const headerTime = document.querySelector(`.mac-time-${CSS.escape(ip)}`);
        if (headerTime) {
          headerTime.textContent = 'Last Scraped: ' + formatTime(Date.now() / 1000);
        }
        const titleCount = document.querySelector(`.mac-count-${CSS.escape(ip)}`);
        if (titleCount) {
          titleCount.textContent = `${state.data.length} entries`;
        }
        renderMacTable(ip);
      } else {
        alert('Failed to refresh MAC table: ' + (data.error || 'Unknown error'));
      }
    })
    .catch(err => {
      alert('Error refreshing MAC table: ' + err.message);
    })
    .finally(() => {
      if (spinner) (spinner as HTMLElement).style.display = 'none';
      if (btnText) btnText.textContent = 'Refresh MACs';
    });
}
export function renderMacTable(ip) {
  const state = getMacState(ip);
  const tbody = document.querySelector(`.mac-tbody-${CSS.escape(ip)}`);
  if (!tbody) return;
  
  ['mac', 'host', 'vendor', 'type', 'port', 'vlan'].forEach(col => {
    const indicator = document.querySelector(`.sort-indicator-${CSS.escape(ip)}-${col}`);
    if (indicator) {
      if (state.sortBy === col) {
        indicator.textContent = state.sortAsc ? ' ▴' : ' ▾';
        (indicator as HTMLElement).style.color = '#58a6ff';
      } else {
        indicator.textContent = '';
      }
    }
  });

  let rows = state.data;
  const filters = state.filters || { mac: '', host: '', vendor: '', type: '', port: '', vlan: '' };
  if (filters.mac || filters.host || filters.vendor || filters.type || filters.port || filters.vlan) {
    rows = rows.filter(r => {
      const matchMac = !filters.mac || (r.mac || '').toLowerCase().includes(filters.mac);
      const matchHost = !filters.host || (r.host || '').toLowerCase().includes(filters.host);
      const matchVendor = !filters.vendor || (r.vendor || '').toLowerCase().includes(filters.vendor);
      const matchType = !filters.type || (r.type || '').toLowerCase().includes(filters.type);
      const matchPort = !filters.port || (r.port || '').toString().toLowerCase().includes(filters.port);
      const matchVlan = !filters.vlan || (r.vlan || '').toString().toLowerCase().includes(filters.vlan);
      return matchMac && matchHost && matchVendor && matchType && matchPort && matchVlan;
    });
  }

  rows.sort((a, b) => {
    let valA = a[state.sortBy] || '';
    let valB = b[state.sortBy] || '';
    
    if (state.sortBy === 'port') {
      valA = parsePortForSort(valA);
      valB = parsePortForSort(valB);
    } else if (state.sortBy === 'vlan') {
      valA = parseVlanForSort(valA);
      valB = parseVlanForSort(valB);
    } else {
      valA = valA.toString().toLowerCase();
      valB = valB.toString().toLowerCase();
    }
    
    if (valA < valB) return state.sortAsc ? -1 : 1;
    if (valA > valB) return state.sortAsc ? 1 : -1;
    return 0;
  });

  if (rows.length === 0) {
    tbody.innerHTML = `<tr><td colspan="6" style="padding: 12px; text-align: center; color: #8b949e;">No MAC entries found</td></tr>`;
    const container = document.querySelector(`.mac-scroll-container-${CSS.escape(ip)}`);
    if (container) {
      container.scrollTop = state.scrollTop || 0;
    }
    return;
  }

  tbody.innerHTML = rows.map(r => {
    const vendorHtml = r.vendor ? r.vendor : `<a href="/config#vendor-editor-card" style="color: #8b949e; text-decoration: underline;" title="Edit custom vendor list">Unknown</a>`;
    const hostInputHtml = `
      <input type="text" value="${r.host || ''}" 
             placeholder="Enter nickname..." 
             onfocus="pausePolling()" 
             onblur="saveHost(this, '${r.mac}')" 
             style="width: 100%; max-width: 100%; background: #161b22; border: 1px solid #30363d; border-radius: 4px; padding: 4px 8px; color: #e1e4e8; font-size: 11px; outline: none; transition: border-color 0.2s;" 
             onfocus="this.style.borderColor='#58a6ff'" 
             onblur="this.style.borderColor='#30363d'">
    `;
    return `
      <tr style="border-bottom: 1px solid #21262d; transition: background 0.15s;" onmouseover="this.style.background='#1c2128'" onmouseout="this.style.background='transparent'">
        <td style="padding: 8px 12px; font-family: 'SF Mono', Monaco, monospace; color: #c9d1d9;">${r.mac}</td>
        <td style="padding: 4px 12px; color: #b1bac4;">${hostInputHtml}</td>
        <td style="padding: 8px 12px; color: #b1bac4;">${vendorHtml}</td>
        <td style="padding: 8px 12px; color: #8b949e;">${r.type}</td>
        <td style="padding: 8px 12px; font-weight: 600; color: #58a6ff;">${r.port}</td>
        <td style="padding: 8px 12px; color: #b1bac4;">VLAN ${r.vlan}</td>
      </tr>
    `;
  }).join('');

  const container = document.querySelector(`.mac-scroll-container-${CSS.escape(ip)}`);
  if (container) {
    container.scrollTop = state.scrollTop || 0;
  }
}
export function renderSwitch(sw) {
  const renderMacTh = (colId, label) => {
    const savedWidth = columnWidths[`mac-${colId}`];
    const styleStr = `position: relative; padding: 8px 12px; cursor: pointer; color: #8b949e; user-select: none; font-weight: 600;${savedWidth ? ` width: ${savedWidth}px;` : ''}`;
    return `
      <th style="${styleStr}" onclick="sortMac('${sw.ip}', '${colId}')" data-col-id="mac-${colId}">
        ${label} <span class="sort-indicator-${sw.ip}-${colId}"></span>
        <div class="resizer" onmousedown="handleResizeStart(event, 'mac-${colId}')" style="position: absolute; top: 0; right: 0; width: 6px; height: 100%; cursor: col-resize; user-select: none; z-index: 10;"></div>
      </th>
    `;
  };

  const getPortColor = (p) => {
    if (!p) return '#57606a';
    const status = (p.status || '').toLowerCase();
    if (status === 'disable' || status === 'disabled') {
      return '#353c45'; // Disabled (darker grey)
    }
    if (status !== 'up') {
      return '#57606a'; // Down (grey)
    }
    const speed = (p.speed || '').toLowerCase();
    if (speed.includes('10g')) return '#1f6feb'; // 10G (dark blue)
    if (speed.includes('2.5g') || speed.includes('2500') || speed.includes('2g') || speed.includes('2000')) {
      return '#58a6ff'; // 2.5G/2G (medium blue)
    }
    if (speed.includes('1g') || speed.includes('1000')) {
      return '#3fb950'; // 1G (green)
    }
    if (speed.includes('100') || speed.includes('10')) {
      return '#f08a24'; // 100M/10M (orange)
    }
    return '#57606a'; // Down/fallback (grey)
  };

  const wrapThreshold = typeof PORTS_WRAP_THRESHOLD !== 'undefined' ? Number(PORTS_WRAP_THRESHOLD) : 0;
  const rawPorts = sw.ports || [];
  
  const portHtmlList = rawPorts.map(p => {
    const isSfp = p.is_sfp;
    const color = getPortColor(p);
    
    let tooltipParts = [];
    if (p.vm_name) {
      tooltipParts.push(p.vm_name);
    }
    if (p.interface && p.interface !== p.port) {
      tooltipParts.push(p.interface);
    }
    const details = tooltipParts.length > 0 ? ` (${tooltipParts.join(' / ')})` : '';
    const titleText = `Port ${p.port}${details}: ${p.status.toUpperCase()} ${p.speed || ''} (${p.duplex || ''}) - Click to view graph`;
    const path = isSfp 
      ? 'M 4,5 h 3 v 3 h 10 v -3 h 3 v 14 h -16 z' // SFP (Fiber)
      : 'M 4,8 h 5 v -3 h 6 v 3 h 5 v 11 h -16 z'; // RJ45 (Copper)
      
    return `
      <div style="display: flex; flex-direction: column; align-items: center; gap: 2px; cursor: pointer;" 
           title="${titleText}"
           onclick="openGraph('${sw.ip}','${p.port}','${sw.name}','Port ${p.port}', ${p.cum_tx || p.tx_bytes || 0}, ${p.cum_rx || p.rx_bytes || 0})">
        <span style="font-size: 10px; color: #8b949e; font-weight: 700; margin-bottom: 2px;">${p.port}</span>
        <svg viewBox="0 0 24 24" style="width: 18px; height: 18px; fill: ${color}; filter: drop-shadow(0 1px 2px rgba(0,0,0,0.4)); transition: all 0.15s ease-in-out; transform: scale(1);"
             onmouseover="this.style.transform='scale(1.2)';"
             onmouseout="this.style.transform='scale(1)';">
          <path d="${path}" />
        </svg>
      </div>
    `;
  });

  let rowsHtml = '';
  if (wrapThreshold > 0 && portHtmlList.length > wrapThreshold) {
    for (let i = 0; i < portHtmlList.length; i += wrapThreshold) {
      const chunk = portHtmlList.slice(i, i + wrapThreshold);
      rowsHtml += `
        <div style="display: flex; gap: 10px; align-items: center; justify-content: center;">
          ${chunk.join('')}
        </div>
      `;
    }
  } else {
    rowsHtml = `
      <div style="display: flex; gap: 10px; align-items: center; justify-content: center;">
        ${portHtmlList.join('')}
      </div>
    `;
  }

  const portBar = (rawPorts.length > 0) ? `
    <div style="background: rgba(13, 17, 23, 0.75); border: 1px solid #30363d; border-radius: 20px; padding: 10px 18px; display: inline-flex; flex-direction: column; gap: 10px; align-items: center; justify-content: center; box-shadow: inset 0 2px 4px rgba(0,0,0,0.4), 0 4px 10px rgba(0,0,0,0.3); margin: 4px 0;">
      ${rowsHtml}
    </div>
  ` : '';

  const activeColumnOrder = columnOrder.filter(colId => COLUMN_DEFS[colId]);

  const theadCols = activeColumnOrder.map(colId => {
    const colDef = COLUMN_DEFS[colId];
    const label = colDef.getLabel ? colDef.getLabel() : colDef.label;
    const savedWidth = columnWidths[colId];
    const styleWidth = savedWidth ? `width: ${savedWidth}px;` : (colDef.thStyle || '');
    
    return `<th draggable="true"
                ondragstart="handleDragStart(event, '${colId}')"
                ondragover="handleDragOver(event)"
                ondragenter="handleDragEnter(event, '${colId}')"
                ondragleave="handleDragLeave(event)"
                ondragend="handleDragEnd(event)"
                ondrop="handleDrop(event, '${colId}')"
                data-col-id="${colId}"
                style="cursor: move; user-select: none; ${styleWidth}"
                class="${colDef.thClass || ''}">
              ${label}
              <div class="resizer" onmousedown="handleResizeStart(event, '${colId}')"></div>
            </th>`;
  }).join('');

  const portRows = (sw.ports || []).map(p => {
    const cells = activeColumnOrder.map(colId => {
      const colDef = COLUMN_DEFS[colId];
      const content = colDef.render(p, sw);
      const tdClass = colDef.tdClassFn ? colDef.tdClassFn(p) : (colDef.tdClass || '');
      const tdTitle = colDef.tdTitleFn ? colDef.tdTitleFn(p) : (colDef.tdTitle || '');
      const classAttr = tdClass ? `class="${tdClass}"` : '';
      const titleAttr = tdTitle ? `title="${tdTitle}"` : '';
      
      return `<td ${classAttr} ${titleAttr}>${content}</td>`;
    }).join('');
    return `<tr>${cells}</tr>`;
  }).join('');

  const errorBlock = sw.error ? `<div class="error-card">Error: ${sw.error}</div>` : '';
  const uptime = sw.uptime ? `<span><span class="label">Uptime:</span> <span class="value">${sw.uptime}</span></span>` : '';
  const mac = sw.mac ? `<span><span class="label">MAC:</span> <span class="value">${sw.mac}</span></span>` : '';
  const fw = sw.firmware ? `<span><span class="label">Firmware:</span> <span class="value">${sw.firmware}</span></span>` : '';

  let dhcpBadge = '';
  if (sw.dhcp_snooping) {
    const isActive = sw.dhcp_snooping.enabled;
    dhcpBadge = `<span><span class="label">DHCP Snooping:</span> <span class="value" style="color: ${isActive ? '#56d364' : '#8b949e'}">${isActive ? '🟢 Active' : '⚪ Inactive'}</span></span>`;
  }

  let igmpBadge = '';
  if (sw.igmp) {
    const isActive = sw.igmp.enabled;
    igmpBadge = `<span><span class="label">IGMP Snooping:</span> <span class="value" style="color: ${isActive ? '#56d364' : '#8b949e'}">${isActive ? '🟢 Active' : '⚪ Inactive'}</span></span>`;
  }

  let jumboBadge = '';
  if (sw.jumbo_frame) {
    const isActive = sw.jumbo_frame.enabled;
    const sizeVal = sw.jumbo_frame.size;
    jumboBadge = `<span><span class="label">Jumbo Frame:</span> <span class="value" style="color: ${isActive ? '#56d364' : '#8b949e'}">${isActive ? '🟢 Active (' + sizeVal + ')' : '⚪ Inactive'}</span></span>`;
  }

  const macState = getMacState(sw.ip);
  const isExpanded = macState.expanded;
  const displayStyle = isExpanded ? 'block' : 'none';
  const arrowTransform = isExpanded ? 'rotate(90deg)' : 'rotate(0deg)';
  const macCount = sw.mac_table ? sw.mac_table.length : 0;
  const lastScrapedStr = sw.mac_timestamp ? formatTime(sw.mac_timestamp) : 'Never';

  const macSectionHtml = `
    <div class="mac-section" style="border-top: 1px solid #30363d; padding: 12px 20px; background: rgba(22, 27, 34, 0.2);">
      <div class="mac-header" style="display: flex; justify-content: space-between; align-items: center; cursor: pointer; user-select: none;" onclick="toggleMacTable('${sw.ip}')">
        <h4 style="font-size: 12px; font-weight: 600; color: #8b949e; display: flex; align-items: center; gap: 6px; margin: 0;">
          <span class="mac-toggle-arrow-${sw.ip}" style="display: inline-block; transition: transform 0.2s; transform: ${arrowTransform}; font-size: 10px;">&gt;</span>
          <span>MAC Address Table (<span class="mac-count-${sw.ip}">${macCount} entries</span>)</span>
        </h4>
        <span class="mac-time-${sw.ip}" style="font-size: 10px; color: #8b949e;">Last Scraped: ${lastScrapedStr}</span>
      </div>
      
      <div class="mac-content-${sw.ip}" style="display: ${displayStyle}; margin-top: 12px;">
        <div style="display: flex; justify-content: flex-end; margin-bottom: 10px;">
          <button class="mac-refresh-btn-${sw.ip}" onclick="manualRefreshMac('${sw.ip}')" style="background: #21262d; border: 1px solid #30363d; border-radius: 6px; padding: 6px 12px; color: #c9d1d9; font-size: 11px; font-weight: 600; cursor: pointer; transition: all 0.2s; display: flex; align-items: center; gap: 6px;" onmouseover="this.style.background='#30363d'; this.style.borderColor='#8b949e';" onmouseout="this.style.background='#21262d'; this.style.borderColor='#30363d';">
            <span class="mac-spinner-${sw.ip}" style="display: none; width: 12px; height: 12px; border: 2px solid transparent; border-top-color: #58a6ff; border-radius: 50%; animation: spin .8s linear infinite; box-sizing: border-box;"></span>
            <span>Refresh MACs</span>
          </button>
        </div>
        <div class="mac-scroll-container-${sw.ip}" onscroll="saveMacScroll('${sw.ip}', this)" onmouseup="saveMacHeight('${sw.ip}', this)" ontouchend="saveMacHeight('${sw.ip}', this)" style="height: ${macState.height || '450px'}; min-height: 150px; max-height: 1200px; resize: vertical; overflow-y: auto; border: 1px solid #30363d; border-radius: 8px; background: #0d1117; box-shadow: inset 0 2px 4px rgba(0,0,0,0.5);">
          <table class="mac-table" style="width: 100%; border-collapse: collapse; font-size: 11px; text-align: left; table-layout: fixed;">
            <thead>
              <tr style="background: #161b22; position: sticky; top: 0; z-index: 2; border-bottom: 1px solid #30363d;">
                ${renderMacTh('mac', 'MAC Address')}
                ${renderMacTh('host', 'Host')}
                ${renderMacTh('vendor', 'Vendor')}
                ${renderMacTh('type', 'Type')}
                ${renderMacTh('port', 'Port')}
                ${renderMacTh('vlan', 'VLAN')}
              </tr>
              <tr style="background: #0d1117; position: sticky; top: 31px; z-index: 2; border-bottom: 1px solid #30363d;">
                <th style="padding: 4px 6px;"><input type="text" class="mac-filter-mac-${sw.ip}" placeholder="Filter MAC..." value="${macState.filters ? macState.filters.mac : ''}" onfocus="focusMacSearch()" onblur="blurMacSearch()" oninput="filterMacTable('${sw.ip}')" style="width: 100%; background: #161b22; border: 1px solid #30363d; border-radius: 4px; padding: 4px 8px; color: #e1e4e8; font-size: 11px; outline: none; transition: border-color 0.2s;" onfocus="this.style.borderColor='#58a6ff'" onblur="this.style.borderColor='#30363d'"></th>
                <th style="padding: 4px 6px;"><input type="text" class="mac-filter-host-${sw.ip}" placeholder="Filter Host..." value="${macState.filters ? macState.filters.host : ''}" onfocus="focusMacSearch()" onblur="blurMacSearch()" oninput="filterMacTable('${sw.ip}')" style="width: 100%; background: #161b22; border: 1px solid #30363d; border-radius: 4px; padding: 4px 8px; color: #e1e4e8; font-size: 11px; outline: none; transition: border-color 0.2s;" onfocus="this.style.borderColor='#58a6ff'" onblur="this.style.borderColor='#30363d'"></th>
                <th style="padding: 4px 6px;"><input type="text" class="mac-filter-vendor-${sw.ip}" placeholder="Filter Vendor..." value="${macState.filters ? macState.filters.vendor : ''}" onfocus="focusMacSearch()" onblur="blurMacSearch()" oninput="filterMacTable('${sw.ip}')" style="width: 100%; background: #161b22; border: 1px solid #30363d; border-radius: 4px; padding: 4px 8px; color: #e1e4e8; font-size: 11px; outline: none; transition: border-color 0.2s;" onfocus="this.style.borderColor='#58a6ff'" onblur="this.style.borderColor='#30363d'"></th>
                <th style="padding: 4px 6px;"><input type="text" class="mac-filter-type-${sw.ip}" placeholder="Filter Type..." value="${macState.filters ? macState.filters.type : ''}" onfocus="focusMacSearch()" onblur="blurMacSearch()" oninput="filterMacTable('${sw.ip}')" style="width: 100%; background: #161b22; border: 1px solid #30363d; border-radius: 4px; padding: 4px 8px; color: #e1e4e8; font-size: 11px; outline: none; transition: border-color 0.2s;" onfocus="this.style.borderColor='#58a6ff'" onblur="this.style.borderColor='#30363d'"></th>
                <th style="padding: 4px 6px;"><input type="text" class="mac-filter-port-${sw.ip}" placeholder="Port..." value="${macState.filters ? macState.filters.port : ''}" onfocus="focusMacSearch()" onblur="blurMacSearch()" oninput="filterMacTable('${sw.ip}')" style="width: 100%; background: #161b22; border: 1px solid #30363d; border-radius: 4px; padding: 4px 8px; color: #e1e4e8; font-size: 11px; outline: none; transition: border-color 0.2s;" onfocus="this.style.borderColor='#58a6ff'" onblur="this.style.borderColor='#30363d'"></th>
                <th style="padding: 4px 6px;"><input type="text" class="mac-filter-vlan-${sw.ip}" placeholder="VLAN..." value="${macState.filters ? macState.filters.vlan : ''}" onfocus="focusMacSearch()" onblur="blurMacSearch()" oninput="filterMacTable('${sw.ip}')" style="width: 100%; background: #161b22; border: 1px solid #30363d; border-radius: 4px; padding: 4px 8px; color: #e1e4e8; font-size: 11px; outline: none; transition: border-color 0.2s;" onfocus="this.style.borderColor='#58a6ff'" onblur="this.style.borderColor='#30363d'"></th>
              </tr>
            </thead>
            <tbody class="mac-tbody-${sw.ip}">
              <!-- Populate via Javascript -->
            </tbody>
          </table>
        </div>
      </div>
    </div>
  `;

  const igmpState = getIgmpState(sw.ip);
  const isIgmpExpanded = igmpState.expanded;
  const igmpDisplayStyle = isIgmpExpanded ? 'block' : 'none';
  const igmpArrowTransform = isIgmpExpanded ? 'rotate(90deg)' : 'rotate(0deg)';
  const igmpEntries = sw.igmp && sw.igmp.entries ? sw.igmp.entries : [];
  const igmpCount = igmpEntries.length;

  let igmpRows = '<tr><td colspan="3" style="padding: 12px; text-align: center; color: #8b949e;">No active multicast groups detected</td></tr>';
  if (igmpEntries.length > 0) {
    igmpRows = igmpEntries.map(e => `
      <tr style="border-bottom: 1px solid #21262d; transition: background 0.15s;" onmouseover="this.style.background='#1c2128'" onmouseout="this.style.background='transparent'">
        <td style="padding: 8px 12px; color: #58a6ff; font-weight: 600; font-family: 'SF Mono', Monaco, monospace;">${e.ip}</td>
        <td style="padding: 8px 12px; color: #c9d1d9; font-weight: 600;">${e.ports}</td>
        <td style="padding: 8px 12px; color: #8b949e;">VLAN ${e.vlan}</td>
      </tr>
    `).join('');
  }

  const igmpSectionHtml = `
    <div class="igmp-section" style="border-top: 1px solid #30363d; padding: 12px 20px; background: rgba(22, 27, 34, 0.2);">
      <div class="igmp-header" style="display: flex; justify-content: space-between; align-items: center; cursor: pointer; user-select: none;" onclick="toggleIgmpTable('${sw.ip}')">
        <h4 style="font-size: 12px; font-weight: 600; color: #8b949e; display: flex; align-items: center; gap: 6px; margin: 0;">
          <span class="igmp-toggle-arrow-${sw.ip}" style="display: inline-block; transition: transform 0.2s; transform: ${igmpArrowTransform}; font-size: 10px;">&gt;</span>
          <span>IGMP Multicast Groups (<span class="igmp-count-${sw.ip}">${igmpCount} entries</span>)</span>
        </h4>
      </div>
      
      <div class="igmp-content-${sw.ip}" style="display: ${igmpDisplayStyle}; margin-top: 12px;">
        <div style="border: 1px solid #30363d; border-radius: 8px; background: #0d1117; overflow: hidden; box-shadow: inset 0 2px 4px rgba(0,0,0,0.5);">
          <table style="width: 100%; border-collapse: collapse; font-size: 11px; text-align: left;">
            <thead>
              <tr style="background: #161b22; border-bottom: 1px solid #30363d;">
                <th style="padding: 8px 12px; color: #8b949e; font-weight: 600; width: 45%;">Multicast IP</th>
                <th style="padding: 8px 12px; color: #8b949e; font-weight: 600; width: 35%;">Member Ports</th>
                <th style="padding: 8px 12px; color: #8b949e; font-weight: 600; width: 20%;">VLAN</th>
              </tr>
            </thead>
            <tbody>
              ${igmpRows}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  `;

  // EEE section
  const eeeState = getExtraSectionState(sw.ip, 'eee');
  const isEeeExpanded = eeeState.expanded;
  const eeeData = sw.eee || [];
  const eeeSectionHtml = `
    <div class="eee-section" style="border-top: 1px solid #30363d; padding: 12px 20px; background: rgba(22, 27, 34, 0.2);">
      <div class="eee-header" style="display: flex; justify-content: space-between; align-items: center; cursor: pointer; user-select: none;" onclick="toggleExtraSection('${sw.ip}', 'eee')">
        <h4 style="font-size: 12px; font-weight: 600; color: #8b949e; display: flex; align-items: center; gap: 6px; margin: 0;">
          <span class="eee-toggle-arrow-${sw.ip}" style="display: inline-block; transition: transform 0.2s; transform: ${isEeeExpanded ? 'rotate(90deg)' : 'rotate(0deg)'}; font-size: 10px;">&gt;</span>
          <span>EEE Status (${eeeData.length} ports)</span>
        </h4>
      </div>
      <div class="eee-content-${sw.ip}" style="display: ${isEeeExpanded ? 'block' : 'none'}; margin-top: 12px;">
        <div style="border: 1px solid #30363d; border-radius: 8px; background: #0d1117; overflow: hidden;">
          <table style="width: 100%; border-collapse: collapse; font-size: 11px; text-align: left;">
            <thead>
              <tr style="background: #161b22; border-bottom: 1px solid #30363d;">
                <th style="padding: 8px 12px; color: #8b949e; font-weight: 600;">Port</th>
                <th style="padding: 8px 12px; color: #8b949e; font-weight: 600;">Active</th>
                <th style="padding: 8px 12px; color: #8b949e; font-weight: 600;">Status</th>
                <th style="padding: 8px 12px; color: #8b949e; font-weight: 600;">LP Status</th>
              </tr>
            </thead>
            <tbody>
              ${eeeData.length === 0 ? '<tr><td colspan="4" style="padding: 12px; text-align: center; color: #8b949e;">No EEE data</td></tr>' : eeeData.map(e => `
                <tr style="border-bottom: 1px solid #21262d;">
                  <td style="padding: 8px 12px; font-weight: 600; color: #58a6ff;">${e.port}</td>
                  <td style="padding: 8px 12px; color: ${e.active ? '#3fb950' : '#8b949e'};">${e.active ? 'Active' : 'Inactive'}</td>
                  <td style="padding: 8px 12px; color: #c9d1d9;">${e.status || '-'}</td>
                  <td style="padding: 8px 12px; color: #c9d1d9;">${e.lp_status || '-'}</td>
                </tr>
              `).join('')}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  `;

  // VLAN section
  const vlanState = getExtraSectionState(sw.ip, 'vlan');
  const isVlanExpanded = vlanState.expanded;
  const vlanData = sw.vlan_list || [];
  const vlanSectionHtml = `
    <div class="vlan-section" style="border-top: 1px solid #30363d; padding: 12px 20px; background: rgba(22, 27, 34, 0.2);">
      <div class="vlan-header" style="display: flex; justify-content: space-between; align-items: center; cursor: pointer; user-select: none;" onclick="toggleExtraSection('${sw.ip}', 'vlan')">
        <h4 style="font-size: 12px; font-weight: 600; color: #8b949e; display: flex; align-items: center; gap: 6px; margin: 0;">
          <span class="vlan-toggle-arrow-${sw.ip}" style="display: inline-block; transition: transform 0.2s; transform: ${isVlanExpanded ? 'rotate(90deg)' : 'rotate(0deg)'}; font-size: 10px;">&gt;</span>
          <span>VLAN List (${vlanData.length} entries)</span>
        </h4>
      </div>
      <div class="vlan-content-${sw.ip}" style="display: ${isVlanExpanded ? 'block' : 'none'}; margin-top: 12px;">
        <div style="border: 1px solid #30363d; border-radius: 8px; background: #0d1117; overflow: hidden;">
          <table style="width: 100%; border-collapse: collapse; font-size: 11px; text-align: left;">
            <thead>
              <tr style="background: #161b22; border-bottom: 1px solid #30363d;">
                <th style="padding: 8px 12px; color: #8b949e; font-weight: 600;">VLAN ID</th>
                <th style="padding: 8px 12px; color: #8b949e; font-weight: 600;">Name</th>
              </tr>
            </thead>
            <tbody>
              ${vlanData.length === 0 ? '<tr><td colspan="2" style="padding: 12px; text-align: center; color: #8b949e;">No VLANs configured</td></tr>' : vlanData.map(v => `
                <tr style="border-bottom: 1px solid #21262d;">
                  <td style="padding: 8px 12px; font-weight: 600; color: #58a6ff;">${v.id}</td>
                  <td style="padding: 8px 12px; color: #c9d1d9;">${v.name || '-'}</td>
                </tr>
              `).join('')}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  `;

  // LAG section
  const lagState = getExtraSectionState(sw.ip, 'lag');
  const isLagExpanded = lagState.expanded;
  const lagData = sw.lag || [];
  const lagSectionHtml = `
    <div class="lag-section" style="border-top: 1px solid #30363d; padding: 12px 20px; background: rgba(22, 27, 34, 0.2);">
      <div class="lag-header" style="display: flex; justify-content: space-between; align-items: center; cursor: pointer; user-select: none;" onclick="toggleExtraSection('${sw.ip}', 'lag')">
        <h4 style="font-size: 12px; font-weight: 600; color: #8b949e; display: flex; align-items: center; gap: 6px; margin: 0;">
          <span class="lag-toggle-arrow-${sw.ip}" style="display: inline-block; transition: transform 0.2s; transform: ${isLagExpanded ? 'rotate(90deg)' : 'rotate(0deg)'}; font-size: 10px;">&gt;</span>
          <span>LAG (${lagData.length} groups)</span>
        </h4>
      </div>
      <div class="lag-content-${sw.ip}" style="display: ${isLagExpanded ? 'block' : 'none'}; margin-top: 12px;">
        <div style="border: 1px solid #30363d; border-radius: 8px; background: #0d1117; overflow: hidden;">
          <table style="width: 100%; border-collapse: collapse; font-size: 11px; text-align: left;">
            <thead>
              <tr style="background: #161b22; border-bottom: 1px solid #30363d;">
                <th style="padding: 8px 12px; color: #8b949e; font-weight: 600;">Group</th>
                <th style="padding: 8px 12px; color: #8b949e; font-weight: 600;">Members</th>
                <th style="padding: 8px 12px; color: #8b949e; font-weight: 600;">Hash</th>
              </tr>
            </thead>
            <tbody>
              ${lagData.length === 0 ? '<tr><td colspan="3" style="padding: 12px; text-align: center; color: #8b949e;">No LAGs configured</td></tr>' : lagData.map(l => `
                <tr style="border-bottom: 1px solid #21262d;">
                  <td style="padding: 8px 12px; font-weight: 600; color: #58a6ff;">${l.number}</td>
                  <td style="padding: 8px 12px; color: #c9d1d9;">${l.members || '-'}</td>
                  <td style="padding: 8px 12px; color: #c9d1d9;">${l.hash || '-'}</td>
                </tr>
              `).join('')}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  `;

  // Mirror section
  const mirrorState = getExtraSectionState(sw.ip, 'mirror');
  const isMirrorExpanded = mirrorState.expanded;
  const mirror = sw.mirror;
  const mirrorSectionHtml = `
    <div class="mirror-section" style="border-top: 1px solid #30363d; padding: 12px 20px; background: rgba(22, 27, 34, 0.2);">
      <div class="mirror-header" style="display: flex; justify-content: space-between; align-items: center; cursor: pointer; user-select: none;" onclick="toggleExtraSection('${sw.ip}', 'mirror')">
        <h4 style="font-size: 12px; font-weight: 600; color: #8b949e; display: flex; align-items: center; gap: 6px; margin: 0;">
          <span class="mirror-toggle-arrow-${sw.ip}" style="display: inline-block; transition: transform 0.2s; transform: ${isMirrorExpanded ? 'rotate(90deg)' : 'rotate(0deg)'}; font-size: 10px;">&gt;</span>
          <span>Port Mirroring</span>
        </h4>
      </div>
      <div class="mirror-content-${sw.ip}" style="display: ${isMirrorExpanded ? 'block' : 'none'}; margin-top: 12px;">
        <div style="border: 1px solid #30363d; border-radius: 8px; background: #0d1117; overflow: hidden; padding: 12px;">
          ${!mirror ? '<div style="color: #8b949e; font-size: 11px; text-align: center;">No mirror configuration</div>' : `
            <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 8px; font-size: 11px;">
              <div><span style="color: #8b949e;">Enabled</span></div>
              <div style="color: ${mirror.enabled ? '#3fb950' : '#8b949e'};">${mirror.enabled ? 'Active' : 'Inactive'}</div>
              <div><span style="color: #8b949e;">Mirror Port</span></div>
              <div style="color: #c9d1d9;">${mirror.port || '-'}</div>
              <div><span style="color: #8b949e;">RX Mirror</span></div>
              <div style="color: #c9d1d9;">${mirror.mirror_rx || '-'}</div>
              <div><span style="color: #8b949e;">TX Mirror</span></div>
              <div style="color: #c9d1d9;">${mirror.mirror_tx || '-'}</div>
            </div>
          `}
        </div>
      </div>
    </div>
  `;

  // Bandwidth section
  const bwState = getExtraSectionState(sw.ip, 'bandwidth');
  const isBwExpanded = bwState.expanded;
  const bwData = sw.bandwidth || [];
  const bwSectionHtml = `
    <div class="bandwidth-section" style="border-top: 1px solid #30363d; padding: 12px 20px; background: rgba(22, 27, 34, 0.2);">
      <div class="bandwidth-header" style="display: flex; justify-content: space-between; align-items: center; cursor: pointer; user-select: none;" onclick="toggleExtraSection('${sw.ip}', 'bandwidth')">
        <h4 style="font-size: 12px; font-weight: 600; color: #8b949e; display: flex; align-items: center; gap: 6px; margin: 0;">
          <span class="bandwidth-toggle-arrow-${sw.ip}" style="display: inline-block; transition: transform 0.2s; transform: ${isBwExpanded ? 'rotate(90deg)' : 'rotate(0deg)'}; font-size: 10px;">&gt;</span>
          <span>Bandwidth Control (${bwData.length} ports)</span>
        </h4>
      </div>
      <div class="bandwidth-content-${sw.ip}" style="display: ${isBwExpanded ? 'block' : 'none'}; margin-top: 12px;">
        <div style="border: 1px solid #30363d; border-radius: 8px; background: #0d1117; overflow: hidden;">
          <table style="width: 100%; border-collapse: collapse; font-size: 11px; text-align: left;">
            <thead>
              <tr style="background: #161b22; border-bottom: 1px solid #30363d;">
                <th style="padding: 8px 12px; color: #8b949e; font-weight: 600;">Port</th>
                <th style="padding: 8px 12px; color: #8b949e; font-weight: 600;">Ingress Limit</th>
                <th style="padding: 8px 12px; color: #8b949e; font-weight: 600;">Ingress BW</th>
                <th style="padding: 8px 12px; color: #8b949e; font-weight: 600;">Egress Limit</th>
                <th style="padding: 8px 12px; color: #8b949e; font-weight: 600;">Egress BW</th>
              </tr>
            </thead>
            <tbody>
              ${bwData.length === 0 ? '<tr><td colspan="5" style="padding: 12px; text-align: center; color: #8b949e;">No bandwidth limits configured</td></tr>' : bwData.map(b => `
                <tr style="border-bottom: 1px solid #21262d;">
                  <td style="padding: 8px 12px; font-weight: 600; color: #58a6ff;">${b.port}</td>
                  <td style="padding: 8px 12px; color: ${b.in_limited ? '#d29922' : '#8b949e'};">${b.in_limited ? 'Limited' : 'No Limit'}</td>
                  <td style="padding: 8px 12px; color: #c9d1d9;">${b.in_bw || '-'}</td>
                  <td style="padding: 8px 12px; color: ${b.out_limited ? '#d29922' : '#8b949e'};">${b.out_limited ? 'Limited' : 'No Limit'}</td>
                  <td style="padding: 8px 12px; color: #c9d1d9;">${b.out_bw || '-'}</td>
                </tr>
              `).join('')}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  `;

  return `<div class="switch-card">
    <div class="switch-header">
      <div style="display: flex; align-items: center; gap: 12px;">
        <img src="/api/switches/${sw.ip}/image" alt="Switch" style="height: 24px; width: auto; object-fit: contain; filter: drop-shadow(0 2px 4px rgba(0,0,0,0.5));">
        <div>
          <h2>${sw.name}</h2>
          <span style="font-size:12px;color:#b1bac4;display:flex;align-items:center;gap:8px;">
            <a href="http://${sw.ip}/" target="_blank" style="color:#58a6ff;text-decoration:none" title="Open switch web UI">${sw.ip}</a>
            <span style="color:#30363d">|</span>
            <a href="#" onclick="backupConfig('${sw.ip}', this); return false;" style="color:#58a6ff;text-decoration:none;font-weight:600;display:inline-flex;align-items:center;gap:4px;" title="Backup configuration"> Backup</a>
            <a href="#" onclick="switchConsole('${sw.ip}'); return false;" style="color:#8b949e;text-decoration:none;font-weight:600;font-size:11px;" title="Send CLI command"> Console</a>
          </span>
        </div>
      </div>
      
      <!-- Port Graphic -->
      ${portBar}
      
      <div class="meta">${sw.model || ''}</div>
    </div>
    <div class="device-bar">
      ${uptime}
      ${fw}
      ${mac}
      ${dhcpBadge}
      ${igmpBadge}
      ${jumboBadge}
    </div>
    <table class="port-table">
      <thead><tr>
        ${theadCols}
      </tr></thead>
      <tbody>${portRows || `<tr><td colspan="${activeColumnOrder.length}" style="padding:20px;text-align:center;color:#b1bac4">No data</td></tr>`}</tbody>
    </table>
    ${errorBlock}
    ${eeeSectionHtml}
    ${vlanSectionHtml}
    ${lagSectionHtml}
    ${mirrorSectionHtml}
    ${bwSectionHtml}
    ${macSectionHtml}
    ${igmpSectionHtml}
    <div class="last-update">Last update: ${formatTime(sw.timestamp)}</div>
  </div>`;
}
