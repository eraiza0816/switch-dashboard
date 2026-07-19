import { t, setLang, getLang } from './i18n';

// Test translation engine
function assert(condition: boolean, msg: string) {
  if (!condition) throw new Error(msg);
  console.log('PASS:', msg);
}

// Test default language
assert(getLang() === 'en' || getLang() === 'ja', 'default language is set');

// Test English translation
setLang('en');
assert(t('nav.dashboard') === 'Dashboard', 'en: nav.dashboard');
assert(t('nav.map') === 'Map', 'en: nav.map');
assert(t('status.live') === 'Live', 'en: status.live');
assert(t('btn.save') === 'Save', 'en: btn.save');
assert(t('unknown.key') === 'unknown.key', 'en: unknown key returns key');

// Test Japanese translation
setLang('ja');
assert(t('nav.dashboard') === 'ダッシュボード', 'ja: nav.dashboard');
assert(t('nav.map') === 'ネットワークマップ', 'ja: nav.map');
assert(t('status.live') === 'ライブ', 'ja: status.live');
assert(t('btn.save') === '保存', 'ja: btn.save');
assert(t('unknown.key') === 'unknown.key', 'ja: unknown key returns key');

// Test parameter substitution
setLang('en');
assert(t('status.update_in', { n: 30 }) === 'update in 30s', 'en: parameter substitution');
assert(t('sfp.title', { p: 5 }) === 'Port 5 SFP+ Transceiver Status', 'en: SFP title with port');
assert(t('graph.title', { label: 'Port 1', name: 'Switch' }) === 'Bandwidth - Port 1 (Switch)', 'en: graph title');

// Test Japanese parameter substitution
setLang('ja');
assert(t('sfp.title', { p: 5 }) === 'ポート5 SFP+ トランシーバー診断', 'ja: SFP title with port');
assert(t('mac.entries', { n: 3 }) === '3件', 'ja: MAC entries count');

// Test language switching
setLang('en');
assert(getLang() === 'en', 'getLang returns en after set');
setLang('ja');
assert(getLang() === 'ja', 'getLang returns ja after set');

console.log('ALL TESTS PASSED');
