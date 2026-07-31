import { t, setLang, getLang } from './i18n';
import { showToast } from './dashboard-utils';

interface MapNode {
  id: string;
  name: string;
  type: string;
  status: string;
  ip?: string;
  mac?: string;
  model?: string;
  vendor?: string;
  host?: string;
  device_type?: string;
  x?: number;
  y?: number;
  level?: number;
  last_seen_ip?: string;
  last_seen_port?: string;
  last_seen_time?: number;
}

interface MapLink {
  source: string;
  target: string;
  source_port?: string;
  target_port?: string;
  speed?: string;
  tx_bps?: number;
  rx_bps?: number;
  type: string;
}

const MAP_REFRESH_SECONDS = 10;

let rawNodes: MapNode[] = [];
let rawLinks: MapLink[] = [];
let positionCache: Record<string, { x: number; y: number }> = {};
let panX = 0, panY = 0, zoom = 1;
let dragNode: MapNode | null = null;
let dragOffX = 0, dragOffY = 0;
let panStartX = 0, panStartY = 0;
let isPanning = false;
let selectedNode: MapNode | null = null;
let showClients = true;
let deviceTypes: Record<string, { label: string; icon?: string; path?: string }> = {};
let iconPaths: Record<string, string> = {};
let countdownSeconds = MAP_REFRESH_SECONDS;
let countdownInterval: ReturnType<typeof setInterval> | null = null;
let loading = false;

const DEFAULT_TYPES: Record<string, string> = {
  laptop: 'laptop', smartphone: 'smartphone', server: 'server', pc_desktop: 'pc_desktop',
  nas: 'nas', ipcam: 'ipcam', tv: 'tv', nvr: 'nvr', smart_switch: 'smart_switch',
  smart_plug: 'smart_plug', sensore: 'sensore', audiovideo: 'audiovideo',
  vacuum_robot: 'vacuum_robot', air_conditioner: 'air_conditioner',
  dehumidifier: 'dehumidifier', three_d_printer: 'three_d_printer', dryer: 'dryer',
};

function init() {
  document.title = t('map.title');
  loadPositionCache();
  loadDeviceTypes();
  fetchTopology();
  setupEventListeners();
  startCountdown();
}

async function loadDeviceTypes() {
  try {
    const r = await fetch('/api/device_types');
    deviceTypes = await r.json();
    await loadIcons();
    render();
  } catch {}
}

function iconifyName(key: string): string {
  const dt = deviceTypes[key];
  if (dt && dt.icon) return dt.icon;
  return '';
}

async function loadIcons() {
  const fallbacks: Record<string, string> = {
    switch: 'M6 3h12a1 1 0 0 1 1 1v16a1 1 0 0 1-1 1H6a1 1 0 0 1-1-1V4a1 1 0 0 1 1-1zm1 3v3h2V6H7zm4 0v3h2V6h-2zm4 0v3h2V6h-2zM7 11v2h2v-2H7zm4 0v2h2v-2h-2zm4 0v2h2v-2h-2zM7 15v2h2v-2H7zm4 0v2h2v-2h-2zm4 0v2h2v-2h-2z',
    router: 'M12 2a5 5 0 0 1 4.9 4H19a2 2 0 0 1 2 2v12a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h2.1A5 5 0 0 1 12 2zm0 2a3 3 0 0 0-2.9 2h5.8A3 3 0 0 0 12 4zm-4 8v2h2v-2H8zm6 0v2h2v-2h-2zm-3 0v2h2v-2h-2zm-6 3v3h2v-3H5zm6 0v3h2v-3h-2zm4 0v3h2v-3h-2z',
    client: 'M16 11c1.66 0 3-1.34 3-3s-1.34-3-3-3-3 1.34-3 3 1.34 3 3 3zm-8 0c1.66 0 3-1.34 3-3S9.66 5 8 5 5 6.34 5 8s1.34 3 3 3zm0 2c-2.33 0-7 1.17-7 3.5V19h14v-2.5c0-2.33-4.67-3.5-7-3.5zm8 0c-.29 0-.62.02-.97.05 1.16.84 1.97 1.97 1.97 3.45V19h6v-2.5c0-2.33-4.67-3.5-7-3.5z',
    internet: 'M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-1 17.93c-3.95-.49-7-3.85-7-7.93 0-.62.08-1.21.21-1.79L9 15v1c0 1.1.9 2 2 2v1.93zm6.9-2.54c-.26-.81-1-1.39-1.9-1.39h-1v-3c0-.55-.45-1-1-1H8v-2h2c.55 0 1-.45 1-1V7h2c1.1 0 2-.9 2-2v-.41c2.93 1.19 5 4.06 5 7.41 0 2.08-.8 3.97-2.1 5.39z',
    unmanaged_switch: 'M6 3h12a1 1 0 0 1 1 1v16a1 1 0 0 1-1 1H6a1 1 0 0 1-1-1V4a1 1 0 0 1 1-1z',
  };

  // Gather iconify names grouped by prefix
  const byPrefix: Record<string, Set<string>> = {};
  const keysByIcon: Record<string, string[]> = {};
  for (const [key, val] of Object.entries(deviceTypes)) {
    if (val.path) {
      iconPaths[key] = val.path;
      continue;
    }
    if (!val.icon) continue;
    const parts = val.icon.split(':');
    const prefix = parts.length > 1 ? parts[0] : '';
    const name = parts.length > 1 ? parts[1] : val.icon;
    if (!byPrefix[prefix]) byPrefix[prefix] = new Set();
    byPrefix[prefix].add(name);
    if (!keysByIcon[val.icon]) keysByIcon[val.icon] = [];
    keysByIcon[val.icon].push(key);
  }

  await Promise.all(Object.entries(byPrefix).map(async ([prefix, names]) => {
    try {
      const r = await fetch(`https://api.iconify.design/${prefix}.json?icons=${Array.from(names).join(',')}`);
      if (!r.ok) return;
      const data = await r.json();
      if (data && data.icons) {
        for (const name of names) {
          const body = data.icons[name];
          if (!body) continue;
          const full = `${prefix}:${name}`;
          for (const key of keysByIcon[full] || []) {
            iconPaths[key] = body;
          }
        }
      }
    } catch {}
  }));

  // Fallbacks for types without a resolved icon
  for (const [key, val] of Object.entries(deviceTypes)) {
    if (!iconPaths[key]) iconPaths[key] = '';
  }
  iconPaths.switch = iconPaths.switch || fallbacks.switch;
  iconPaths.router = iconPaths.router || fallbacks.router;
  iconPaths.internet = iconPaths.internet || fallbacks.internet;
  iconPaths.unmanaged_switch = iconPaths.unmanaged_switch || fallbacks.unmanaged_switch;
  iconPaths.client = iconPaths.client || fallbacks.client;
  iconPaths.other = iconPaths.other || fallbacks.client;
}

function loadPositionCache() {
  fetch('/api/layout_positions')
    .then(r => r.json())
    .then(data => {
      if (data && typeof data === 'object') positionCache = data;
    })
    .catch(() => {
      try {
        const saved = localStorage.getItem('map_positions');
        if (saved) positionCache = JSON.parse(saved);
      } catch {}
    });
}

function savePositionCache() {
  fetch('/api/layout_positions', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(positionCache),
  }).catch(() => {
    try { localStorage.setItem('map_positions', JSON.stringify(positionCache)); } catch {}
  });
}

async function fetchTopology() {
  if (loading) return;
  loading = true;
  try {
    const r = await fetch('/api/topology');
    const data = await r.json();
    const prevSelectedId = selectedNode ? selectedNode.id : null;
    rawNodes = data.nodes || [];
    rawLinks = data.links || [];
    computeLayout();
    if (prevSelectedId) {
      selectedNode = rawNodes.find(n => n.id === prevSelectedId) || null;
    }
    render();
    const loadingEl = document.getElementById('map-loading');
    if (loadingEl) loadingEl.style.display = 'none';
  } catch (e) {
    document.getElementById('map-canvas')!.innerHTML = `<div style="padding:40px;text-align:center;color:#f85149;">${t('map.failed_load')}</div>`;
  } finally {
    loading = false;
  }
}

function startCountdown() {
  if (countdownInterval) clearInterval(countdownInterval);
  countdownInterval = setInterval(() => {
    const text = document.getElementById('map-live-text');
    const dot = document.getElementById('map-live-dot');
    const countdown = document.getElementById('map-countdown');
    if (!text || !dot) return;
    countdownSeconds--;
    if (countdownSeconds <= 0) {
      text.textContent = t('map.loading');
      dot.style.background = '#d29922';
      countdownSeconds = MAP_REFRESH_SECONDS;
      fetchTopology();
    } else {
      text.textContent = t('map.live');
      dot.style.background = '#3fb950';
      if (countdown) countdown.textContent = t('map.refresh_in', { n: countdownSeconds });
    }
  }, 1000);
}

function computeLayout() {
  const W = window.innerWidth;
  const H = window.innerHeight - 60;

  const adj: Record<string, string[]> = {};
  for (const n of rawNodes) adj[n.id] = [];
  for (const l of rawLinks) {
    if (adj[l.source]) adj[l.source].push(l.target);
    if (adj[l.target]) adj[l.target].push(l.source);
  }

  const visited = new Set<string>();
  const levels: Record<number, MapNode[]> = {};
  let queue: MapNode[] = [];

  let root = rawNodes.find(n => n.type === 'switch') || rawNodes.find(n => n.type === 'router');
  if (!root) root = rawNodes[0];
  if (!root) return;

  root.level = 0;
  queue = [root];
  visited.add(root.id);
  levels[0] = [root];

  while (queue.length) {
    const cur = queue.shift()!;
    for (const nid of adj[cur.id] || []) {
      if (visited.has(nid)) continue;
      visited.add(nid);
      const node = rawNodes.find(n => n.id === nid);
      if (!node) continue;
      const lvl = (cur.level || 0) + 1;
      node.level = lvl;
      if (!levels[lvl]) levels[lvl] = [];
      levels[lvl].push(node);
      queue.push(node);
    }
  }

  for (const n of rawNodes) {
    if (n.level === undefined) {
      n.level = 1;
      if (!levels[1]) levels[1] = [];
      levels[1].push(n);
    }
  }

  const sortedLevels = Object.keys(levels).map(Number).sort((a, b) => a - b);
  const levelHeight = Math.min(180, (H - 100) / Math.max(sortedLevels.length, 1));
  const margin = 120;

  for (const lvl of sortedLevels) {
    const nodes = levels[lvl];
    const y = 80 + lvl * levelHeight;
    const spacing = Math.min(200, (W - margin * 2) / Math.max(nodes.length, 1));

    for (let i = 0; i < nodes.length; i++) {
      const n = nodes[i];
      if (positionCache[n.id]) {
        n.x = positionCache[n.id].x;
        n.y = positionCache[n.id].y;
      } else {
        n.x = margin + i * spacing + spacing / 2;
        n.y = y + (i % 2 === 0 ? -15 : 15);
      }
    }
  }
}

function nodeIconMarkup(node: MapNode): string {
  let key = 'other';
  if (node.type === 'client') key = node.device_type || 'client';
  else if (node.type === 'switch') key = 'switch';
  else if (node.type === 'internet') key = 'internet';
  else if (node.type === 'router') key = 'router';
  else if (node.type === 'unmanaged_switch') key = 'unmanaged_switch';
  else if (node.type === 'repeater') key = 'repeater';

  const path = iconPaths[key];
  if (path) return `<path d="${path}" fill="currentColor"/>`;
  // Fallback: single letter
  const letters: Record<string, string> = { switch: 'S', internet: 'W', router: 'R', repeater: 'A', unmanaged_switch: 'U', client: 'C' };
  return `<text x="12" y="16" text-anchor="middle" dominant-baseline="middle" font-size="11" font-weight="700" fill="currentColor">${letters[key] || 'C'}</text>`;
}

function render() {
  const canvas = document.getElementById('map-canvas');
  if (!canvas) return;
  const W = window.innerWidth;
  const H = window.innerHeight - 60;

  let html = `<svg width="${W}" height="${H}" id="map-svg">
    <defs>
      <marker id="arrow" viewBox="0 0 10 10" refX="10" refY="5" markerWidth="6" markerHeight="6" orient="auto">
        <path d="M 0 0 L 10 5 L 0 10 z" fill="#30363d"/>
      </marker>
    </defs>
    <rect width="${W}" height="${H}" fill="#0d1117" id="map-bg"/>
    <g id="viewport" transform="translate(${panX},${panY}) scale(${zoom})">`;

  const filteredLinks = showClients ? rawLinks : rawLinks.filter(l => l.type !== 'client');

  for (const link of filteredLinks) {
    const src = rawNodes.find(n => n.id === link.source);
    const tgt = rawNodes.find(n => n.id === link.target);
    if (!src || !tgt || src.x === undefined || tgt.x === undefined) continue;

    const x1 = src.x, y1 = src.y + 20;
    const x2 = tgt.x, y2 = tgt.y - 20;
    const midY = (y1 + y2) / 2;
    const color = linkSpeedColor(link.speed);

    html += `<g class="link-group" data-source="${link.source}" data-target="${link.target}" style="cursor:pointer;">`;
    html += `<path d="M ${x1} ${y1} C ${x1} ${midY}, ${x2} ${midY}, ${x2} ${y2}" fill="none" stroke="${color}" stroke-width="2" marker-end="url(#arrow)"/>`;
    if (link.source_port) {
      const mx = (x1 + x2) / 2 - 30;
      const my = midY - 8;
      html += `<text x="${mx}" y="${my}" fill="#8b949e" font-size="9" font-family="monospace">${link.source_port}</text>`;
    }
    html += `</g>`;
  }

  const displayNodes = showClients ? rawNodes : rawNodes.filter(n => n.type !== 'client');

  for (const node of displayNodes) {
    if (node.x === undefined) continue;
    const isSelected = selectedNode?.id === node.id;
    const fill = nodeColor(node);
    const w = node.type === 'switch' ? 180 : 140;
    const h = node.type === 'switch' ? 60 : 44;
    const offline = node.status === 'offline';

    html += `<g class="node-group ${offline ? 'node-offline' : ''}" data-id="${node.id}" transform="translate(${node.x - w/2},${node.y - h/2})" style="cursor:grab">`;
    html += `<rect width="${w}" height="${h}" rx="8" fill="${fill.bg}" stroke="${isSelected ? '#58a6ff' : fill.border}" stroke-width="${isSelected ? 2 : 1}"/>`;

    // Status dot
    const dotColor = node.status === 'online' ? '#3fb950' : (node.status === 'disable' ? '#57606a' : '#f85149');
    html += `<circle cx="${w - 12}" cy="12" r="4" fill="${dotColor}"/>`;

    // Icon
    html += `<g transform="translate(12,12)"><svg width="22" height="22" viewBox="0 0 24 24" style="display:block;color:${fill.text}">${nodeIconMarkup(node)}</svg></g>`;

    // Name
    html += `<text x="42" y="20" fill="#f0f6fc" font-size="12" font-weight="500">${truncate(node.name, 18)}</text>`;
    if (node.type === 'switch' && node.model) {
      html += `<text x="42" y="34" fill="#8b949e" font-size="9">${node.model}</text>`;
    }
    if (node.type === 'client') {
      const sub = node.vendor || (node.device_type && deviceTypes[node.device_type] ? deviceTypes[node.device_type].label : '');
      html += `<text x="42" y="34" fill="#8b949e" font-size="9">${truncate(sub, 18)}</text>`;
    }

    if (node.type === 'switch') {
      html += `<text x="${w/2}" y="${h + 14}" fill="#8b949e" font-size="9" text-anchor="middle">${node.ip || ''}</text>`;
    }

    // Highlight pulse
    if (isSelected) {
      html += `<circle class="pulse-ring" cx="${w/2}" cy="${h/2}" r="20" fill="none" stroke="#58a6ff" stroke-width="2" style="display:none;" id="pulse-${node.id}"/>`;
    }

    html += `</g>`;
  }

  html += `</g></svg>`;
  canvas.innerHTML = html;

  document.querySelectorAll('.node-group').forEach(el => {
    el.addEventListener('mousedown', onNodeMouseDown);
  });
  document.querySelectorAll('.link-group').forEach(el => {
    el.addEventListener('mousemove', onLinkMouseMove);
    el.addEventListener('mouseleave', hideLinkTooltip);
  });
  document.getElementById('map-bg')?.addEventListener('mousedown', onBgMouseDown);
}

function nodeColor(node: MapNode) {
  switch (node.type) {
    case 'switch': return { bg: 'rgba(88,166,255,0.1)', border: '#58a6ff', text: '#58a6ff' };
    case 'internet': return { bg: 'rgba(35,134,54,0.1)', border: '#3fb950', text: '#3fb950' };
    case 'router': return { bg: 'rgba(35,134,54,0.1)', border: '#3fb950', text: '#3fb950' };
    case 'repeater': return { bg: 'rgba(242,175,36,0.1)', border: '#d29922', text: '#d29922' };
    case 'unmanaged_switch': return { bg: 'rgba(56,139,253,0.1)', border: '#38bdf8', text: '#38bdf8' };
    default: return { bg: 'rgba(139,148,158,0.08)', border: '#8b949e', text: '#8b949e' };
  }
}

function linkSpeedColor(speed?: string): string {
  if (!speed) return '#30363d';
  const s = speed.toLowerCase();
  if (s.includes('10g')) return '#1f6feb';
  if (s.includes('2.5g') || s.includes('2500')) return '#58a6ff';
  if (s.includes('1g') || s.includes('1000')) return '#3fb950';
  if (s.includes('100m')) return '#d29922';
  return '#30363d';
}

function truncate(s: string, n: number): string {
  if (!s) return '';
  return s.length > n ? s.slice(0, n - 1) + '\u2026' : s;
}

function formatTraffic(bps?: number): string {
  if (!bps) return '0 bps';
  if (bps >= 1e9) return (bps / 1e9).toFixed(1) + ' Gbps';
  if (bps >= 1e6) return (bps / 1e6).toFixed(1) + ' Mbps';
  if (bps >= 1e3) return (bps / 1e3).toFixed(1) + ' Kbps';
  return bps + ' bps';
}

function onLinkMouseMove(e: MouseEvent) {
  const el = e.currentTarget as SVGGElement;
  const source = el.getAttribute('data-source') || '';
  const target = el.getAttribute('data-target') || '';
  const link = rawLinks.find(l => l.source === source && l.target === target);
  if (!link) return;
  const src = rawNodes.find(n => n.id === link.source);
  const tgt = rawNodes.find(n => n.id === link.target);
  if (!src || !tgt) return;

  const speedColor = linkSpeedColor(link.speed);
  let trafficHtml = '';
  if (link.tx_bps || link.rx_bps) {
    trafficHtml = `
      <div class="row" style="border-top:1px dashed rgba(48,54,61,0.4);margin-top:4px;padding-top:4px;">
        <span>TX Speed:</span><span class="val blue">${formatTraffic(link.tx_bps)}</span>
      </div>
      <div class="row">
        <span>RX Speed:</span><span class="val green">${formatTraffic(link.rx_bps)}</span>
      </div>`;
  }

  const tooltip = document.getElementById('link-tooltip')!;
  tooltip.innerHTML = `
    <div class="title">
      <span>Connection Link</span>
      <span style="color:${speedColor}">${link.speed || 'Unknown'}</span>
    </div>
    <div class="row"><span>Source:</span><span class="val">${src.name}</span></div>
    ${link.source_port ? `<div class="row"><span>Port:</span><span class="val">${link.source_port}</span></div>` : ''}
    <div class="row"><span>Target:</span><span class="val">${tgt.name}</span></div>
    ${link.target_port ? `<div class="row"><span>Port:</span><span class="val">${link.target_port}</span></div>` : ''}
    ${trafficHtml}
  `;
  tooltip.style.display = 'block';
  tooltip.style.left = (e.clientX + 14) + 'px';
  tooltip.style.top = (e.clientY + 14) + 'px';
}

function hideLinkTooltip() {
  const tooltip = document.getElementById('link-tooltip');
  if (tooltip) tooltip.style.display = 'none';
}

function onNodeMouseDown(e: MouseEvent) {
  const el = e.currentTarget as SVGGElement;
  const id = el.getAttribute('data-id');
  const node = rawNodes.find(n => n.id === id);
  if (!node) return;

  if (e.button === 0) {
    dragNode = node;
    const rect = el.getBoundingClientRect();
    dragOffX = e.clientX - rect.left;
    dragOffY = e.clientY - rect.top;
    el.style.cursor = 'grabbing';
    selectNode(node);
  }
}

function onBgMouseDown(e: MouseEvent) {
  if (e.target === e.currentTarget) {
    selectedNode = null;
    updateSidebar();
    isPanning = true;
    panStartX = e.clientX - panX;
    panStartY = e.clientY - panY;
  }
}

function selectNode(node: MapNode) {
  selectedNode = node;
  updateSidebar();
  render();
  // Pulse highlight
  const pulse = document.getElementById(`pulse-${node.id}`);
  if (pulse) pulse.style.display = 'block';
}

function updateSidebar() {
  const sidebar = document.getElementById('sidebar');
  const content = document.getElementById('sidebar-content');
  if (!sidebar || !content) return;
  if (!selectedNode) { sidebar.style.display = 'none'; content.innerHTML = ''; return; }
  sidebar.style.display = 'block';

  const n = selectedNode;
  const links = rawLinks.filter(l => l.source === n.id || l.target === n.id);

  const typeLabel = t('map.' + n.type) || n.type;
  const statusLabel = t('status.' + n.status) || n.status;
  let html = `<div style="padding:16px;">
    <h3 style="font-size:15px;font-weight:600;color:#f0f6fc;margin-bottom:4px;">${n.name}</h3>
    <p style="font-size:11px;color:#8b949e;margin-bottom:12px;">${typeLabel} &middot; ${statusLabel}</p>`;

  if (n.ip) html += `<div class="info-row"><span class="label">${t('map.ip')}</span><span class="value">${n.ip}</span></div>`;
  if (n.mac) html += `<div class="info-row"><span class="label">${t('mac.mac')}</span><span class="value">${n.mac}</span></div>`;
  if (n.model) html += `<div class="info-row"><span class="label">${t('config.model')}</span><span class="value">${n.model}</span></div>`;
  if (n.vendor) html += `<div class="info-row"><span class="label">${t('map.vendor')}</span><span class="value">${n.vendor}</span></div>`;
  if (n.last_seen_ip) html += `<div class="info-row"><span class="label">${t('map.last_location')}</span><span class="value">${n.last_seen_ip}:${n.last_seen_port || ''}</span></div>`;
  if (n.last_seen_time) html += `<div class="info-row"><span class="label">${t('map.last_active')}</span><span class="value">${new Date(n.last_seen_time * 1000).toLocaleString()}</span></div>`;

  if (links.length) {
    html += `<h4 style="font-size:12px;color:#f0f6fc;margin:12px 0 6px;">${t('map.links')}</h4>`;
    for (const l of links) {
      const peer = rawNodes.find(n => n.id === (l.source === n.id ? l.target : l.source));
      const port = l.source === n.id ? l.source_port : l.target_port;
      html += `<div class="info-row"><span class="value">${peer?.name || '?'}${port ? ' (' + port + ')' : ''}</span></div>`;
    }
  }

  if (n.type === 'client') {
    const clientNames = [...new Set(rawNodes
      .filter(c => c.type === 'client' && (c.host || c.name))
      .map(c => c.host || c.name)
    )];

    const typesToRender = Object.keys(deviceTypes).length > 0 ? deviceTypes : DEFAULT_TYPES;
    let options = '';
    for (const [key, val] of Object.entries(typesToRender)) {
      const label = typeof val === 'string' ? DEFAULT_TYPES[key] || key : (val.label || key);
      const selected = n.device_type === key ? 'selected' : '';
      options += `<option value="${key}" ${selected}>${label}</option>`;
    }

    html += `<div style="margin-top:12px;display:flex;gap:6px;flex-direction:column;">
      <div style="display:flex;gap:6px;">
        <input id="rename-input" list="client-name-suggestions" value="${n.host || n.name}" style="flex:1;padding:4px 8px;background:#0d1117;border:1px solid #30363d;border-radius:6px;color:#c9d1d9;font-size:12px;" placeholder="${t('map.nickname')}"/>
        <button class="btn btn-secondary" onclick="renameClient()" style="padding:4px 10px;font-size:11px;">${t('btn.save')}</button>
      </div>
      <datalist id="client-name-suggestions">${clientNames.map(cn => `<option value="${cn}">`).join('')}</datalist>
      <div style="display:flex;gap:6px;margin-top:4px;">
        <select id="inspect-type-select" onchange="updateClientType(this.value)" style="flex:1;padding:4px 8px;background:#0d1117;border:1px solid #30363d;border-radius:6px;color:#c9d1d9;font-size:12px;">
          <option value="">${t('map.device_type')}...</option>
          ${options}
        </select>
      </div>
      <button class="btn btn-danger" onclick="forgetClient()" style="margin-top:6px;padding:4px 10px;font-size:11px;">${t('map.forget')}</button>
    </div>`;
  }

  html += `</div>`;
  content.innerHTML = html;

  const typeSelect = document.getElementById('inspect-type-select') as HTMLSelectElement | null;
  if (typeSelect && n.device_type) {
    for (const opt of Array.from(typeSelect.options)) {
      if (opt.value === n.device_type) opt.selected = true;
    }
  }
}

function setupEventListeners() {
  document.addEventListener('mousemove', onMouseMove);
  document.addEventListener('mouseup', onMouseUp);
  document.addEventListener('wheel', onWheel, { passive: false });
  document.getElementById('search-input')?.addEventListener('input', onSearch);

  document.getElementById('sidebar-close')?.addEventListener('click', () => {
    selectedNode = null;
    updateSidebar();
  });

  const fileInput = document.getElementById('csv-file-input') as HTMLInputElement | null;
  if (fileInput) {
    fileInput.addEventListener('change', onCSVFileSelected);
  }
}

function onMouseMove(e: MouseEvent) {
  if (isPanning) {
    panX = e.clientX - panStartX;
    panY = e.clientY - panStartY;
    render();
  } else if (dragNode) {
    dragNode.x = (e.clientX - panX) / zoom - dragOffX / zoom;
    dragNode.y = (e.clientY - panY) / zoom - dragOffY / zoom;
    render();
  }
}

function onMouseUp() {
  if (isPanning) {
    isPanning = false;
    render();
  } else if (dragNode) {
    positionCache[dragNode.id] = { x: dragNode.x!, y: dragNode.y! };
    savePositionCache();
    dragNode = null;
    render();
  }
}

function onWheel(e: WheelEvent) {
  e.preventDefault();
  const delta = e.deltaY > 0 ? 0.9 : 1.1;
  zoom = Math.max(0.15, Math.min(4, zoom * delta));
  render();
}

function onSearch(e: Event) {
  const q = (e.target as HTMLInputElement).value.toLowerCase();
  const results = document.getElementById('search-results');
  if (!q || q.length < 2) { results!.innerHTML = ''; return; }

  const matches = rawNodes.filter(n =>
    n.name.toLowerCase().includes(q) ||
    (n.mac && n.mac.toLowerCase().includes(q.replace(/:/g, ''))) ||
    (n.ip && n.ip.includes(q)) ||
    (n.vendor && n.vendor.toLowerCase().includes(q))
  ).slice(0, 5);

  if (!matches.length) { results!.innerHTML = ''; return; }

  results!.innerHTML = matches.map(m =>
    `<div class="search-item" onclick="searchNavigate('${m.id}')">
      <span style="color:${nodeColor(m).text}">${m.type === 'switch' ? 'S' : m.type === 'router' ? 'R' : 'C'}</span>
      <span>${m.name}</span>
      <span style="color:#8b949e;font-size:10px;">${m.mac ? m.mac.slice(-8) : ''}</span>
    </div>`
  ).join('');
}

function searchNavigate(id: string) {
  const node = rawNodes.find(n => n.id === id);
  if (!node || node.x === undefined) return;
  panX = window.innerWidth / 2 - node.x * zoom;
  panY = (window.innerHeight - 60) / 2 - node.y * zoom;
  zoom = 1;
  document.getElementById('search-results')!.innerHTML = '';
  (document.getElementById('search-input') as HTMLInputElement).value = '';
  selectNode(node);
  render();
}

function resetLayout() {
  if (!confirm(t('map.reset_layout') + '?')) return;
  positionCache = {};
  savePositionCache();
  computeLayout();
  render();
}

function renameClient() {
  if (!selectedNode || !selectedNode.mac) return;
  const input = document.getElementById('rename-input') as HTMLInputElement;
  if (!input) return;
  const host = input.value.trim();
  fetch('/api/clients/update_host', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ mac: selectedNode.mac, host }),
  }).then(r => {
    if (!r.ok) throw new Error('rename failed');
    selectedNode!.name = host;
    selectedNode!.host = host;
    updateSidebar();
    render();
    showToast(t('map.bulk_rename_saved'), 'toast-success');
  }).catch(() => {
    showToast(t('toast.failed'), 'toast-error');
  });
}

async function updateClientType(type: string) {
  if (!selectedNode || !selectedNode.mac) return;
  const mac = selectedNode.mac;
  try {
    const r = await fetch('/api/clients/update_type', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ mac, type }),
    });
    if (!r.ok) throw new Error();
    selectedNode.device_type = type || undefined;
    updateSidebar();
    render();
    showToast(t('toast.saved'), 'toast-success');
  } catch {
    showToast(t('toast.failed'), 'toast-error');
  }
}

function forgetClient() {
  if (!selectedNode || !selectedNode.mac) return;
  if (!confirm(t('map.forget_confirm'))) return;
  const mac = selectedNode.mac;
  fetch('/api/clients/delete', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ mac }),
  }).then(r => {
    if (!r.ok) throw new Error();
    delete positionCache[mac];
    selectedNode = null;
    updateSidebar();
    fetchTopology();
    showToast(t('toast.saved'), 'toast-success');
  }).catch(() => {
    showToast(t('toast.failed'), 'toast-error');
  });
}

function importClientsCSV() {
  const input = document.getElementById('csv-file-input') as HTMLInputElement;
  if (input) input.click();
}

function onCSVFileSelected(e: Event) {
  const input = e.target as HTMLInputElement;
  const file = input.files?.[0];
  if (!file) return;
  const form = new FormData();
  form.append('file', file);
  fetch('/api/clients/import_csv', {
    method: 'POST',
    body: form,
  }).then(r => r.json()).then(data => {
    if (data.status === 'ok') {
      showToast(t('map.import_csv_done', { n: data.imported }), 'toast-success');
    } else {
      showToast(t('map.import_csv_failed'), 'toast-error');
    }
    fetchTopology();
  }).catch(() => {
    showToast(t('map.import_csv_failed'), 'toast-error');
  }).finally(() => {
    input.value = '';
  });
}

function toggleClients() {
  showClients = !showClients;
  render();
}

function bulkRename() {
  const existing = document.getElementById('bulk-rename-modal');
  if (existing) {
    existing.classList.toggle('open');
    return;
  }
  const clients = rawNodes.filter(n => n.type === 'client');
  const modal = document.createElement('div');
  modal.id = 'bulk-rename-modal';
  modal.className = 'modal-overlay open';
  modal.onclick = (e) => { if (e.target === modal) modal.classList.remove('open'); };
  modal.innerHTML = `
    <div class="modal" style="max-width:600px;">
      <h3>${t('map.bulk_rename_title')} (${clients.length})</h3>
      <div style="max-height:50vh;overflow-y:auto;">
        ${clients.map(c => `
          <div class="bulk-rename-row" data-mac="${c.mac || c.id}">
            <span class="bulk-rename-mac">${(c.mac || c.id).slice(-8)}</span>
            <input class="bulk-rename-input" value="${c.host || c.name}" placeholder="${t('map.nickname')}"/>
            <button class="btn btn-secondary bulk-rename-save" style="padding:4px 10px;font-size:11px;flex-shrink:0;">${t('btn.save')}</button>
          </div>
        `).join('')}
      </div>
      <div style="margin-top:12px;display:flex;gap:8px;justify-content:flex-end;">
        <button class="btn btn-secondary" onclick="document.getElementById('bulk-rename-modal').classList.remove('open')">${t('map.bulk_close')}</button>
        <button class="btn btn-primary" id="bulk-save-all">${t('map.bulk_save')}</button>
      </div>
    </div>`;
  document.body.appendChild(modal);

  modal.querySelectorAll('.bulk-rename-save').forEach(btn => {
    btn.addEventListener('click', () => {
      const row = btn.closest('.bulk-rename-row') as HTMLElement;
      const mac = row.getAttribute('data-mac') || '';
      const input = row.querySelector('.bulk-rename-input') as HTMLInputElement;
      const host = input.value.trim();
      if (!host) return;
      fetch('/api/clients/update_host', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ mac, host }),
      }).then(r => {
        if (!r.ok) throw new Error();
        const node = rawNodes.find(n => (n.mac || n.id) === mac);
        if (node) { node.name = host; node.host = host; }
        if (selectedNode && (selectedNode.mac || selectedNode.id) === mac) {
          selectedNode.name = host;
          selectedNode.host = host;
          updateSidebar();
        }
        render();
        showToast(t('toast.saved'), 'toast-success');
      }).catch(() => {
        showToast(t('toast.failed'), 'toast-error');
      });
    });
  });

  document.getElementById('bulk-save-all')?.addEventListener('click', () => {
    const rows = modal.querySelectorAll('.bulk-rename-row');
    let pending = 0;
    rows.forEach(row => {
      const mac = (row as HTMLElement).getAttribute('data-mac') || '';
      const input = row.querySelector('.bulk-rename-input') as HTMLInputElement;
      const host = input.value.trim();
      if (!host) return;
      pending++;
      fetch('/api/clients/update_host', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ mac, host }),
      }).then(r => {
        if (!r.ok) throw new Error();
        const node = rawNodes.find(n => (n.mac || n.id) === mac);
        if (node) { node.name = host; node.host = host; }
        if (selectedNode && (selectedNode.mac || selectedNode.id) === mac) {
          selectedNode.name = host;
          selectedNode.host = host;
          updateSidebar();
        }
      }).catch(() => {}).finally(() => {
        pending--;
        if (pending <= 0) {
          render();
          modal.classList.remove('open');
          showToast(t('map.bulk_rename_saved'), 'toast-success');
        }
      });
    });
    if (pending === 0) modal.classList.remove('open');
  });
}

window.addEventListener('load', init);
window.addEventListener('resize', () => { render(); });
document.addEventListener('click', (e) => {
  if (!(e.target as HTMLElement).closest('.search-item') && !(e.target as HTMLElement).closest('#search-input')) {
    const results = document.getElementById('search-results');
    if (results) results.innerHTML = '';
  }
});
(window as any).setLang = setLang;
(window as any).renameClient = renameClient;
(window as any).updateClientType = updateClientType;
(window as any).forgetClient = forgetClient;
(window as any).importClientsCSV = importClientsCSV;
(window as any).toggleClients = toggleClients;
(window as any).resetLayout = resetLayout;
(window as any).searchNavigate = searchNavigate;
(window as any).bulkRename = bulkRename;
