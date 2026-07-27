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
  it('returns B for small values', () => {
    expect(formatBytes(0)).toBe('0 B');
    expect(formatBytes(500)).toBe('500 B');
    expect(formatBytes(999)).toBe('999 B');
  });

  it('returns KB', () => {
    expect(formatBytes(1000)).toBe('1.00 KB');
    expect(formatBytes(1500)).toBe('1.50 KB');
    expect(formatBytes(999999)).toBe('1000.00 KB');
  });

  it('returns MB', () => {
    expect(formatBytes(1e6)).toBe('1.00 MB');
    expect(formatBytes(2.5e6)).toBe('2.50 MB');
    expect(formatBytes(1e9 - 1)).toBe('1000.00 MB');
  });

  it('returns GB', () => {
    expect(formatBytes(1e9)).toBe('1.00 GB');
    expect(formatBytes(1e12 - 1)).toBe('1000.00 GB');
  });

  it('returns TB', () => {
    expect(formatBytes(1e12)).toBe('1.00 TB');
    expect(formatBytes(2e12)).toBe('2.00 TB');
  });
});

describe('formatBps', () => {
  it('returns 0 for falsy input', () => {
    expect(formatBps(0, 'Bps')).toBe('0 Bps');
    expect(formatBps(null as any, 'Bps')).toBe('0 Bps');
  });

  it('formats in Bps mode (divides by 8)', () => {
    expect(formatBps(8000, 'Bps')).toBe('1.00 KBps');
    expect(formatBps(8e6, 'Bps')).toBe('1.00 MBps');
    expect(formatBps(8e9, 'Bps')).toBe('1.00 GBps');
  });

  it('formats in bps mode (no division)', () => {
    expect(formatBps(8000, 'bps')).toBe('8.00 Kbps');
    expect(formatBps(8e6, 'bps')).toBe('8.00 Mbps');
    expect(formatBps(8e9, 'bps')).toBe('8.00 Gbps');
  });
});

describe('formatPkts', () => {
  it('returns 0 for falsy input', () => {
    expect(formatPkts(0)).toBe('0');
    expect(formatPkts(null as any)).toBe('0');
  });

  it('formats K', () => {
    expect(formatPkts(1500)).toBe('1.50K');
  });

  it('formats M', () => {
    expect(formatPkts(1e6)).toBe('1.00M');
    expect(formatPkts(2e6)).toBe('2.00M');
  });

  it('formats G', () => {
    expect(formatPkts(1e9)).toBe('1.00G');
  });
});

describe('speedClass', () => {
  it('returns empty for disable status', () => {
    expect(speedClass('10G', 'disable')).toBe('');
  });

  it('returns speed-10g', () => {
    expect(speedClass('10G', 'up')).toBe('speed-10g');
    expect(speedClass('10g', 'up')).toBe('speed-10g');
  });

  it('returns speed-25g', () => {
    expect(speedClass('2.5G', 'up')).toBe('speed-25g');
  });

  it('returns speed-1g', () => {
    expect(speedClass('1G', 'up')).toBe('speed-1g');
    expect(speedClass('1000', 'up')).toBe('speed-1g');
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
  });

  it('returns "just now" for recent timestamps', () => {
    const now = Date.now() / 1000;
    expect(formatTime(now)).toBe('just now');
  });

  it('returns "m ago" for minutes', () => {
    const fiveMinAgo = (Date.now() - 300000) / 1000;
    expect(formatTime(fiveMinAgo)).toMatch(/m ago/);
  });

  it('returns "h ago" for hours', () => {
    const twoHoursAgo = (Date.now() - 7200000) / 1000;
    expect(formatTime(twoHoursAgo)).toMatch(/h ago/);
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
