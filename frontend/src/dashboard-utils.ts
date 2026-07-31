export const REFRESH_SECONDS = (window.__DATA__ && window.__DATA__.refresh) || 30;
export const enabledColumns = (window.__DATA__ && window.__DATA__.columns) || ['port', 'status', 'speed', 'packets', 'bytes', 'info', 'notes'];
export const PORTS_WRAP_THRESHOLD = (window.__DATA__ && window.__DATA__.portWrap) || 0;

export function formatBytes(n: number): string {
  if (n === undefined || n === null) return '';
  if (n >= 1e12) return (n / 1e12).toFixed(1) + ' TB';
  if (n >= 1e9) return (n / 1e9).toFixed(1) + ' GB';
  if (n >= 1e6) return (n / 1e6).toFixed(1) + ' MB';
  if (n >= 1e3) return (n / 1e3).toFixed(1) + ' KB';
  return n + ' B';
}

export function formatBps(n: number, unit: string): string {
  if (!n) return unit === 'Bps' ? '0 B/s' : '0 bps';
  const v = unit === 'Bps' ? n / 8 : n;
  const suf = unit === 'Bps' ? ['B/s', 'KB/s', 'MB/s', 'GB/s'] : ['bps', 'Kbps', 'Mbps', 'Gbps'];
  const thresholds = [1e9, 1e6, 1e3];
  if (v >= thresholds[0]) return (v / thresholds[0]).toFixed(1) + ' ' + suf[3];
  if (v >= thresholds[1]) return (v / thresholds[1]).toFixed(1) + ' ' + suf[2];
  if (v >= thresholds[2]) return (v / thresholds[2]).toFixed(1) + ' ' + suf[1];
  return v.toFixed(0) + ' ' + suf[0];
}

export function formatPkts(n: number): string | number {
  if (!n) return '0';
  if (n >= 1e6) return (n / 1e6).toFixed(1) + 'M';
  if (n >= 1e3) return (n / 1e3).toFixed(1) + 'K';
  return n;
}

export function speedClass(speed: string, status: string): string {
  if (!status || status.toLowerCase() !== 'up') {
    return 'speed-down';
  }
  if (!speed) return '';
  const s = speed.toLowerCase();
  if (s.includes('10g')) return 'speed-10g';
  if (s.includes('2500') || s.includes('2.5g')) return 'speed-2500m';
  if (s.includes('2000') || s.includes('2g')) return 'speed-2000m';
  if (s.includes('1000') || s.includes('1g')) return 'speed-1000m';
  if (s.includes('100')) return 'speed-100m';
  if (s.includes('10')) return 'speed-10m';
  return '';
}

export function formatTime(ts: number): string {
  if (!ts) return '';
  const d = new Date(ts * 1000);
  return d.toLocaleTimeString();
}

export function formatNumber(n: number): string {
  if (!n && n !== 0) return '-';
  return n.toLocaleString();
}

export function parseHexPort(portStr: string): string | number {
  const v = parseInt(portStr);
  return isNaN(v) ? portStr : v;
}

export function showToast(message: string, type: string, icon = "") {
  const toast = document.getElementById("toast-box")!;
  const toastIcon = document.getElementById("toast-icon")!;
  const toastMsg = document.getElementById("toast-message")!;

  toast.className = `toast ${type}`;
  toastIcon.textContent = icon;
  toastMsg.textContent = message;

  toast.classList.add("show");
  setTimeout(() => {
    toast.classList.remove("show");
  }, 4000);
}
