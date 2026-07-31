import { describe, it, expect, beforeEach } from 'vitest';
import { t, setLang, getLang } from './i18n';

beforeEach(() => {
  setLang('en');
});

describe('i18n', () => {
  it('default language is set', () => {
    expect(['en', 'ja']).toContain(getLang());
  });

  it('returns English translations', () => {
    setLang('en');
    expect(t('nav.dashboard')).toBe('Dashboard');
    expect(t('nav.map')).toBe('Map');
    expect(t('status.live')).toBe('Live');
    expect(t('btn.save')).toBe('Save');
  });

  it('returns Japanese translations', () => {
    setLang('ja');
    expect(t('nav.dashboard')).toBe('ダッシュボード');
    expect(t('nav.map')).toBe('ネットワークマップ');
    expect(t('status.live')).toBe('ライブ');
    expect(t('btn.save')).toBe('保存');
  });

  it('returns key for unknown translations', () => {
    expect(t('unknown.key')).toBe('unknown.key');
  });

  it('substitutes parameters in English', () => {
    setLang('en');
    expect(t('status.update_in', { n: 30 })).toBe('update in 30s');
    expect(t('sfp.title', { p: 5 })).toBe('Port 5 SFP+ Transceiver Status');
    expect(t('graph.title', { label: 'Port 1', name: 'Switch' })).toBe('Bandwidth - Port 1 (Switch)');
  });

  it('substitutes parameters in Japanese', () => {
    setLang('ja');
    expect(t('sfp.title', { p: 5 })).toBe('ポート5 SFP+ トランシーバー診断');
    expect(t('mac.entries', { n: 3 })).toBe('3件');
  });

  it('switches language with getLang/setLang', () => {
    setLang('en');
    expect(getLang()).toBe('en');
    setLang('ja');
    expect(getLang()).toBe('ja');
  });

  it('handles map translations in English', () => {
    setLang('en');
    expect(t('map.title')).toBe('Network Map');
    expect(t('map.switch')).toBe('Switch');
    expect(t('map.client')).toBe('Client');
    expect(t('map.bulk_rename')).toBe('Bulk Rename');
  });

  it('handles map translations in Japanese', () => {
    setLang('ja');
    expect(t('map.title')).toBe('ネットワークマップ');
    expect(t('map.switch')).toBe('スイッチ');
    expect(t('map.client')).toBe('クライアント');
  });
});
