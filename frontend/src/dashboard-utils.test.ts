import { describe, it, expect } from 'vitest';
import {
  formatBytes,
  formatBps,
  formatPkts,
  speedClass,
  formatTime,
  formatNumber,
  parseHexPort,
} from './dashboard-utils';

describe('formatBytes', () => {
  it('returns empty for null/undefined', () => {
    expect(formatBytes(null as any)).toBe('');
    expect(formatBytes(undefined as any)).toBe('');
  });

  it('returns B for small values', () => {
    expect(formatBytes(0)).toBe('0 B');
    expect(formatBytes(500)).toBe('500 B');
    expect(formatBytes(999)).toBe('999 B');
  });

  it('returns KB', () => {
    expect(formatBytes(1000)).toBe('1.0 KB');
    expect(formatBytes(1500)).toBe('1.5 KB');
  });

  it('returns MB', () => {
    expect(formatBytes(1e6)).toBe('1.0 MB');
    expect(formatBytes(2.5e6)).toBe('2.5 MB');
  });

  it('returns GB', () => {
    expect(formatBytes(1e9)).toBe('1.0 GB');
  });

  it('returns TB', () => {
    expect(formatBytes(1e12)).toBe('1.0 TB');
    expect(formatBytes(2e12)).toBe('2.0 TB');
  });
});

describe('formatBps', () => {
  it('returns 0 for falsy input', () => {
    expect(formatBps(0, 'Bps')).toBe('0 B/s');
    expect(formatBps(0, 'bps')).toBe('0 bps');
    expect(formatBps(null as any, 'Bps')).toBe('0 B/s');
  });

  it('formats in Bps mode (divides by 8)', () => {
    expect(formatBps(8000, 'Bps')).toBe('1.0 KB/s');
    expect(formatBps(8e6, 'Bps')).toBe('1.0 MB/s');
    expect(formatBps(8e9, 'Bps')).toBe('1.0 GB/s');
  });

  it('formats in bps mode (no division)', () => {
    expect(formatBps(8000, 'bps')).toBe('8.0 Kbps');
    expect(formatBps(8e6, 'bps')).toBe('8.0 Mbps');
    expect(formatBps(8e9, 'bps')).toBe('8.0 Gbps');
  });
});

describe('formatPkts', () => {
  it('returns 0 for falsy input', () => {
    expect(formatPkts(0)).toBe('0');
    expect(formatPkts(null as any)).toBe('0');
  });

  it('formats K', () => {
    expect(formatPkts(1500)).toBe('1.5K');
  });

  it('formats M', () => {
    expect(formatPkts(1e6)).toBe('1.0M');
    expect(formatPkts(2e6)).toBe('2.0M');
  });

  it('returns raw number below 1k', () => {
    expect(formatPkts(999)).toBe(999);
  });
});

describe('speedClass', () => {
  it('returns speed-down for non-up status', () => {
    expect(speedClass('10G', 'disable')).toBe('speed-down');
    expect(speedClass('10G', 'down')).toBe('speed-down');
    expect(speedClass('10G', '')).toBe('speed-down');
  });

  it('returns speed-10g', () => {
    expect(speedClass('10G', 'up')).toBe('speed-10g');
    expect(speedClass('10g', 'up')).toBe('speed-10g');
  });

  it('returns speed-2500m', () => {
    expect(speedClass('2.5G', 'up')).toBe('speed-2500m');
  });

  it('returns speed-1000m', () => {
    expect(speedClass('1G', 'up')).toBe('speed-1000m');
    expect(speedClass('1000', 'up')).toBe('speed-1000m');
  });

  it('returns speed-100m', () => {
    expect(speedClass('100M', 'up')).toBe('speed-100m');
  });

  it('returns empty for unknown speed', () => {
    expect(speedClass('Auto', 'up')).toBe('');
  });
});

describe('formatTime', () => {
  it('returns empty for falsy timestamp', () => {
    expect(formatTime(0)).toBe('');
    expect(formatTime(undefined as any)).toBe('');
  });

  it('formats timestamp as local time string', () => {
    const ts = new Date(2024, 0, 15, 10, 30, 0).getTime() / 1000;
    expect(formatTime(ts)).toContain(':');
  });
});

describe('formatNumber', () => {
  it('returns dash for null/undefined', () => {
    expect(formatNumber(null as any)).toBe('-');
  });

  it('formats with locale separators', () => {
    expect(formatNumber(0)).toBe('0');
    expect(formatNumber(1000)).toBe('1,000');
    expect(formatNumber(1000000)).toBe('1,000,000');
  });
});

describe('parseHexPort', () => {
  it('returns the number for a numeric string', () => {
    expect(parseHexPort('5')).toBe(5);
    expect(parseHexPort('0')).toBe(0);
  });

  it('returns the string for non-numeric input', () => {
    expect(parseHexPort('abc')).toBe('abc');
    expect(parseHexPort('')).toBe('');
  });
});
