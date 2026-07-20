export const REFRESH_SECONDS = (window.__DATA__ && window.__DATA__.refresh) || 30;
export const enabledColumns = (window.__DATA__ && window.__DATA__.columns) || ['port', 'status', 'speed', 'packets', 'bytes', 'info', 'notes'];
export const PORTS_WRAP_THRESHOLD = (window.__DATA__ && window.__DATA__.portWrap) || 0;

export const state = {
  currentGraphIp: null as string | null,
  currentGraphPort: null as string | null,
  currentGraphRange: 'live' as string,
  currentGraphSwitchName: '',
  currentGraphPortLabel: '',
  speedUnit: 'Bps' as string,
  pollingPaused: false,
  countdownSeconds: 0,
  countdownInterval: null as any,
  currentTransceiverIp: null as string | null,
  currentTransceiverPort: null as string | null,
  currentTransceiverSwitchName: null as string | null,
};

export function formatBytes(n: number): string {
  if (n >= 1e12) return (n / 1e12).toFixed(2) + ' TB';
  if (n >= 1e9) return (n / 1e9).toFixed(2) + ' GB';
  if (n >= 1e6) return (n / 1e6).toFixed(2) + ' MB';
  if (n >= 1e3) return (n / 1e3).toFixed(2) + ' KB';
  return n + ' B';
}

export function formatBps(n: number, unit: string): string {
  if (!n || n === 0) return '0 ' + unit;
  const d = unit === 'bps' ? 1 : 8;
  const v = n / d;
  if (v >= 1e9) return (v / 1e9).toFixed(2) + ' G' + unit;
  if (v >= 1e6) return (v / 1e6).toFixed(2) + ' M' + unit;
  if (v >= 1e3) return (v / 1e3).toFixed(2) + ' K' + unit;
  return v.toFixed(1) + ' ' + unit;
}

export function formatPkts(n: number): string {
  if (!n || n === 0) return '0';
  if (n >= 1e9) return (n / 1e9).toFixed(2) + 'G';
  if (n >= 1e6) return (n / 1e6).toFixed(2) + 'M';
  if (n >= 1e3) return (n / 1e3).toFixed(2) + 'K';
  return '' + n;
}

export function speedClass(speed: string, status: string): string {
  if (status === 'disable') return '';
  const s = (speed || '').toLowerCase();
  if (s.includes('10g')) return 'speed-10g';
  if (s.includes('2.5g')) return 'speed-25g';
  if (s.includes('1g') || s.includes('1000')) return 'speed-1g';
  if (s.includes('100m')) return 'speed-100m';
  return '';
}

export function formatTime(ts: number): string {
  if (!ts) return '';
  const d = new Date(ts * 1000);
  const now = Date.now();
  const diff = now - d.getTime();
  if (diff < 60000) return 'just now';
  if (diff < 3600000) return Math.floor(diff / 60000) + 'm ago';
  if (diff < 86400000) return Math.floor(diff / 3600000) + 'h ago';
  return d.toLocaleDateString();
}

export function formatNumber(n: number): string {
  if (!n && n !== 0) return '-';
  return n.toLocaleString();
}

export function parseHexPort(portStr: string): string | number {
  const v = parseInt(portStr);
  return isNaN(v) ? portStr : v;
}
