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

// Test map translations (English)
setLang('en');
assert(t('map.title') === 'Network Map', 'en: map.title');
assert(t('map.switch') === 'Switch', 'en: map.switch');
assert(t('map.client') === 'Client', 'en: map.client');
assert(t('map.internet') === 'Internet', 'en: map.internet');
assert(t('map.router') === 'Router', 'en: map.router');
assert(t('map.repeater') === 'Repeater', 'en: map.repeater');
assert(t('map.unmanaged_switch') === 'Unmanaged Switch', 'en: map.unmanaged_switch');
assert(t('map.search_placeholder') === 'Search nodes...', 'en: map.search_placeholder');
assert(t('map.details') === 'Details', 'en: map.details');
assert(t('map.links') === 'Links', 'en: map.links');
assert(t('map.nickname') === 'Nickname', 'en: map.nickname');
assert(t('map.toggle_clients') === 'Toggle Clients', 'en: map.toggle_clients');
assert(t('map.reset_layout') === 'Reset Layout', 'en: map.reset_layout');
assert(t('map.failed_load') === 'Failed to load topology', 'en: map.failed_load');
assert(t('map.loading') === 'Loading topology...', 'en: map.loading');
assert(t('map.ip') === 'IP', 'en: map.ip');
assert(t('map.bulk_rename') === 'Bulk Rename', 'en: map.bulk_rename');
assert(t('map.bulk_rename_title') === 'Edit Client Names', 'en: map.bulk_rename_title');
assert(t('map.bulk_rename_saved') === 'Names saved', 'en: map.bulk_rename_saved');
assert(t('map.bulk_save') === 'Save All', 'en: map.bulk_save');
assert(t('map.bulk_close') === 'Close', 'en: map.bulk_close');

// Test map translations (Japanese)
setLang('ja');
assert(t('map.title') === 'ネットワークマップ', 'ja: map.title');
assert(t('map.switch') === 'スイッチ', 'ja: map.switch');
assert(t('map.client') === 'クライアント', 'ja: map.client');
assert(t('map.internet') === 'インターネット', 'ja: map.internet');
assert(t('map.router') === 'ルーター', 'ja: map.router');
assert(t('map.repeater') === 'リピーター', 'ja: map.repeater');
assert(t('map.unmanaged_switch') === 'アンマネージドスイッチ', 'ja: map.unmanaged_switch');
assert(t('map.search_placeholder') === 'ノードを検索...', 'ja: map.search_placeholder');
assert(t('map.details') === '詳細', 'ja: map.details');
assert(t('map.links') === 'リンク', 'ja: map.links');
assert(t('map.nickname') === 'ニックネーム', 'ja: map.nickname');
assert(t('map.toggle_clients') === 'クライアント表示切替', 'ja: map.toggle_clients');
assert(t('map.reset_layout') === 'レイアウトリセット', 'ja: map.reset_layout');
assert(t('map.failed_load') === 'トポロジーの読み込みに失敗しました', 'ja: map.failed_load');
assert(t('map.loading') === 'トポロジーを読み込み中...', 'ja: map.loading');
assert(t('map.ip') === 'IP', 'ja: map.ip');
assert(t('map.bulk_rename') === '一括リネーム', 'ja: map.bulk_rename');
assert(t('map.bulk_rename_title') === 'クライアント名の編集', 'ja: map.bulk_rename_title');
assert(t('map.bulk_rename_saved') === '名前を保存しました', 'ja: map.bulk_rename_saved');
assert(t('map.bulk_save') === 'すべて保存', 'ja: map.bulk_save');
assert(t('map.bulk_close') === '閉じる', 'ja: map.bulk_close');

// Test language switching
setLang('en');
assert(getLang() === 'en', 'getLang returns en after set');
setLang('ja');
assert(getLang() === 'ja', 'getLang returns ja after set');

console.log('ALL TESTS PASSED');
