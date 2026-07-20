import { t, setLang, getLang } from './i18n';

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
let searchResults: MapNode[] = [];
const ICON_PATHS: Record<string, string> = {};

function init() {
  document.title = t('map.title');
  loadPositionCache();
  loadDeviceTypes();
  fetchTopology();
  setupEventListeners();
}

async function loadDeviceTypes() {
  try {
    const r = await fetch('/api/device_types');
    deviceTypes = await r.json();
  } catch {}
}

function loadPositionCache() {
  try {
    const saved = localStorage.getItem('map_positions');
    if (saved) positionCache = JSON.parse(saved);
  } catch {}
}

function savePositionCache() {
  try { localStorage.setItem('map_positions', JSON.stringify(positionCache)); } catch {}
}

async function fetchTopology() {
  try {
    const r = await fetch('/api/topology');
    const data = await r.json();
    rawNodes = data.nodes || [];
    rawLinks = data.links || [];
    computeLayout();
    render();
  } catch (e) {
    document.getElementById('map-canvas').innerHTML = `<div style="padding:40px;text-align:center;color:#f85149;">${t('map.failed_load')}</div>`;
  }
}

function computeLayout() {
  const W = window.innerWidth;
  const H = window.innerHeight - 60;

  // Build adjacency
  const adj: Record<string, string[]> = {};
  for (const n of rawNodes) adj[n.id] = [];
  for (const l of rawLinks) {
    if (adj[l.source]) adj[l.source].push(l.target);
    if (adj[l.target]) adj[l.target].push(l.source);
  }

  // BFS levels
  const visited = new Set<string>();
  const levels: Record<number, MapNode[]> = {};
  let queue: MapNode[] = [];

  // Find root: switch or first node
  let root = rawNodes.find(n => n.type === 'switch');
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

  // Position unvisited nodes
  for (const n of rawNodes) {
    if (n.level === undefined) {
      n.level = 1;
      if (!levels[1]) levels[1] = [];
      levels[1].push(n);
    }
  }

  // Position by level
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

function render() {
  const canvas = document.getElementById('map-canvas');
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

  // Links
  const filteredLinks = showClients ? rawLinks : rawLinks.filter(l => l.type !== 'client');

  for (const link of filteredLinks) {
    const src = rawNodes.find(n => n.id === link.source);
    const tgt = rawNodes.find(n => n.id === link.target);
    if (!src || !tgt || src.x === undefined || tgt.x === undefined) continue;

    const x1 = src.x, y1 = src.y + 20;
    const x2 = tgt.x, y2 = tgt.y - 20;
    const midY = (y1 + y2) / 2;
    const color = linkSpeedColor(link.speed);

    html += `<g class="link-group">`;
    html += `<path d="M ${x1} ${y1} C ${x1} ${midY}, ${x2} ${midY}, ${x2} ${y2}" fill="none" stroke="${color}" stroke-width="2" marker-end="url(#arrow)"/>`;
    if (link.source_port) {
      const mx = (x1 + x2) / 2 - 30;
      const my = midY - 8;
      html += `<text x="${mx}" y="${my}" fill="#8b949e" font-size="9" font-family="monospace">${link.source_port}</text>`;
    }
    html += `</g>`;
  }

  // Nodes
  const displayNodes = showClients ? rawNodes : rawNodes.filter(n => n.type !== 'client');

  for (const node of displayNodes) {
    if (node.x === undefined) continue;
    const isSelected = selectedNode?.id === node.id;
    const fill = nodeColor(node);
    const w = node.type === 'switch' ? 180 : 140;
    const h = node.type === 'switch' ? 60 : 44;

    html += `<g class="node-group" data-id="${node.id}" transform="translate(${node.x - w/2},${node.y - h/2})" style="cursor:grab">`;
    html += `<rect width="${w}" height="${h}" rx="8" fill="${fill.bg}" stroke="${isSelected ? '#58a6ff' : fill.border}" stroke-width="${isSelected ? 2 : 1}"/>`;

    // Status dot
    const dotColor = node.status === 'online' ? '#3fb950' : '#f85149';
    html += `<circle cx="${w - 12}" cy="12" r="4" fill="${dotColor}"/>`;

    // Icon or letter
    html += `<text x="14" y="28" fill="${fill.text}" font-size="14" font-weight="600" text-anchor="middle">${nodeIcon(node)}</text>`;

    // Name
    html += `<text x="28" y="20" fill="#f0f6fc" font-size="12" font-weight="500">${truncate(node.name, 16)}</text>`;
    if (node.type === 'switch' && node.model) {
      html += `<text x="28" y="34" fill="#8b949e" font-size="9">${node.model}</text>`;
    }
    if (node.type === 'client' && node.vendor) {
      html += `<text x="28" y="34" fill="#8b949e" font-size="9">${truncate(node.vendor, 18)}</text>`;
    }

    // Port label for switch nodes
    if (node.type === 'switch') {
      html += `<text x="${w/2}" y="${h + 14}" fill="#8b949e" font-size="9" text-anchor="middle">${node.ip || ''}</text>`;
    }

    html += `</g>`;
  }

  html += `</g></svg>`;
  canvas.innerHTML = html;

  // Bind events
  document.querySelectorAll('.node-group').forEach(el => {
    el.addEventListener('mousedown', onNodeMouseDown);
  });
  document.getElementById('map-bg')?.addEventListener('mousedown', onBgMouseDown);
}

function nodeColor(node: MapNode) {
  switch (node.type) {
    case 'switch': return { bg: 'rgba(88,166,255,0.1)', border: '#58a6ff', text: '#58a6ff' };
    case 'internet': return { bg: 'rgba(35,134,54,0.1)', border: '#3fb950', text: '#3fb950' };
    case 'router': return { bg: 'rgba(35,134,54,0.1)', border: '#3fb950', text: '#3fb950' };
    case 'repeater': return { bg: 'rgba(242,175,36,0.1)', border: '#d29922', text: '#d29922' };
    default: return { bg: 'rgba(139,148,158,0.08)', border: '#8b949e', text: '#8b949e' };
  }
}

function nodeIcon(node: MapNode): string {
  switch (node.type) {
    case 'switch': return 'S';
    case 'internet': return 'W';
    case 'router': return 'R';
    case 'repeater': return 'A';
    case 'unmanaged_switch': return 'U';
    default: return 'C';
  }
}

function linkSpeedColor(speed?: string): string {
  if (!speed) return '#30363d';
  const s = speed.toLowerCase();
  if (s.includes('10g')) return '#58a6ff';
  if (s.includes('2.5g')) return '#8b949e';
  if (s.includes('1g') || s.includes('1000')) return '#3fb950';
  if (s.includes('100m')) return '#d29922';
  return '#30363d';
}

function truncate(s: string, n: number): string {
  return s.length > n ? s.slice(0, n - 1) + '\u2026' : s;
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
  if (n.vendor) html += `<div class="info-row"><span class="label">${t('mac.vendor')}</span><span class="value">${n.vendor}</span></div>`;

  if (links.length) {
    html += `<h4 style="font-size:12px;color:#f0f6fc;margin:12px 0 6px;">${t('map.links')}</h4>`;
    for (const l of links) {
      const peer = rawNodes.find(n => n.id === (l.source === n.id ? l.target : l.source));
      const port = l.source === n.id ? l.source_port : l.target_port;
      html += `<div class="info-row"><span class="value">${peer?.name || '?'}${port ? ' (' + port + ')' : ''}</span></div>`;
    }
  }

  if (n.type === 'client') {
    html += `<div style="margin-top:12px;display:flex;gap:6px;">
      <input id="rename-input" value="${n.host || n.name}" style="flex:1;padding:4px 8px;background:#0d1117;border:1px solid #30363d;border-radius:6px;color:#c9d1d9;font-size:12px;" placeholder="${t('map.nickname')}"/>
      <button class="btn btn-secondary" onclick="renameClient()" style="padding:4px 10px;font-size:11px;">${t('btn.save')}</button>
    </div>`;
  }

  html += `</div>`;
  sidebar.innerHTML = html;
}

function setupEventListeners() {
  document.addEventListener('mousemove', onMouseMove);
  document.addEventListener('mouseup', onMouseUp);
  document.addEventListener('wheel', onWheel, { passive: false });
  document.getElementById('search-input')?.addEventListener('input', onSearch);

  // Sidebar close
  document.getElementById('sidebar-close')?.addEventListener('click', () => {
    selectedNode = null;
    updateSidebar();
  });
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
    (n.mac && n.mac.toLowerCase().includes(q)) ||
    (n.ip && n.ip.includes(q)) ||
    (n.vendor && n.vendor.toLowerCase().includes(q))
  ).slice(0, 5);

  if (!matches.length) { results!.innerHTML = ''; return; }

  results!.innerHTML = matches.map(m =>
    `<div class="search-item" onclick="searchNavigate('${m.id}')">
      <span style="color:${nodeColor(m).text}">${nodeIcon(m)}</span>
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
  if (!host) return;
  fetch('/api/clients/update_host', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ mac: selectedNode.mac, host }),
  }).then(r => {
    if (!r.ok) throw new Error('rename failed');
    selectedNode.name = host;
    selectedNode.host = host;
    updateSidebar();
    render();
  }).catch(() => {});
}

function toggleClients() {
  showClients = !showClients;
  render();
}

window.addEventListener('load', init);
window.addEventListener('resize', () => { render(); });
(window as any).setLang = setLang;
(window as any).renameClient = renameClient;
(window as any).toggleClients = toggleClients;
(window as any).resetLayout = resetLayout;
(window as any).searchNavigate = searchNavigate;
