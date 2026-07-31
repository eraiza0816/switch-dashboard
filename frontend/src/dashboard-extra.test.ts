import { describe, it, expect, beforeEach } from 'vitest';

// jsdom doesn't provide CSS.escape — polyfill it
if (typeof CSS === 'undefined') {
  (globalThis as any).CSS = {};
}
if (!CSS.escape) {
  CSS.escape = (s: string) => s.replace(/[!"#$%&'()*+,./:;<=>?@[\]^`{|}~ ]/g, '\\$&');
}

// Minimal DOM setup for getExtraSectionState / toggleExtraSection
function setupDom(ip: string, section: string) {
  document.body.innerHTML = `
    <div class="${section}-content-${ip}" style="display: none;"></div>
    <span class="${section}-toggle-arrow-${ip}">&gt;</span>
  `;
}

function cleanupDom() {
  document.body.innerHTML = '';
  localStorage.clear();
}

// Replicate the functions from dashboard.ts for testing
function getExtraSectionState(ip: string, section: string) {
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

function toggleExtraSection(ip: string, section: string) {
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

describe('getExtraSectionState', () => {
  beforeEach(() => {
    localStorage.clear();
  });

  it('returns default collapsed state for new IP/section', () => {
    const state = getExtraSectionState('10.0.0.1', 'eee');
    expect(state.expanded).toBe(false);
  });

  it('returns saved expanded state', () => {
    localStorage.setItem('extra_section_states', JSON.stringify({
      '10.0.0.1': { eee: { expanded: true } },
    }));
    const state = getExtraSectionState('10.0.0.1', 'eee');
    expect(state.expanded).toBe(true);
  });

  it('isolates per-IP and per-section state', () => {
    localStorage.setItem('extra_section_states', JSON.stringify({
      '10.0.0.1': { eee: { expanded: true } },
      '10.0.0.2': { vlan: { expanded: true } },
    }));
    expect(getExtraSectionState('10.0.0.1', 'eee').expanded).toBe(true);
    expect(getExtraSectionState('10.0.0.1', 'vlan').expanded).toBe(false);
    expect(getExtraSectionState('10.0.0.2', 'vlan').expanded).toBe(true);
    expect(getExtraSectionState('10.0.0.2', 'eee').expanded).toBe(false);
  });
});

describe('toggleExtraSection', () => {
  beforeEach(() => {
    cleanupDom();
  });

  it('toggles state from false to true', () => {
    setupDom('10.0.0.1', 'eee');
    expect(getExtraSectionState('10.0.0.1', 'eee').expanded).toBe(false);
    toggleExtraSection('10.0.0.1', 'eee');
    expect(getExtraSectionState('10.0.0.1', 'eee').expanded).toBe(true);
  });

  it('toggles state from true to false', () => {
    setupDom('10.0.0.1', 'eee');
    toggleExtraSection('10.0.0.1', 'eee');
    expect(getExtraSectionState('10.0.0.1', 'eee').expanded).toBe(true);
    toggleExtraSection('10.0.0.1', 'eee');
    expect(getExtraSectionState('10.0.0.1', 'eee').expanded).toBe(false);
  });

  it('persists expanded state to localStorage', () => {
    setupDom('10.0.0.1', 'vlan');
    toggleExtraSection('10.0.0.1', 'vlan');
    const stored = JSON.parse(localStorage.getItem('extra_section_states')!);
    expect(stored['10.0.0.1']['vlan'].expanded).toBe(true);
  });

  it('toggles DOM visibility when elements exist', () => {
    setupDom('10.0.0.1', 'lag');
    const content = document.querySelector(`.lag-content-${CSS.escape('10.0.0.1')}`) as HTMLElement;
    const arrow = document.querySelector(`.lag-toggle-arrow-${CSS.escape('10.0.0.1')}`) as HTMLElement;

    expect(content.style.display).toBe('none');
    expect(arrow.style.transform).toBe('');

    toggleExtraSection('10.0.0.1', 'lag');
    expect(content.style.display).toBe('block');
    expect(arrow.style.transform).toBe('rotate(90deg)');

    toggleExtraSection('10.0.0.1', 'lag');
    expect(content.style.display).toBe('none');
    expect(arrow.style.transform).toBe('rotate(0deg)');
  });

  it('handles missing DOM elements gracefully', () => {
    cleanupDom(); // no DOM elements
    expect(() => toggleExtraSection('10.0.0.1', 'eee')).not.toThrow();
    expect(getExtraSectionState('10.0.0.1', 'eee').expanded).toBe(true);
  });
});
