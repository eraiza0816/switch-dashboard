import { t, setLang } from './i18n';

function assert(condition: boolean, msg: string) {
  if (!condition) throw new Error(msg);
  console.log('PASS:', msg);
}

setLang('en');
assert(t('map.nickname') === 'Nickname', 'en: map.nickname');
assert(t('btn.save') === 'Save', 'en: btn.save');
assert(t('map.bulk_rename') === 'Bulk Rename', 'en: map.bulk_rename');
assert(t('map.bulk_rename_title') === 'Edit Client Names', 'en: map.bulk_rename_title');
assert(t('map.bulk_rename_saved') === 'Names saved', 'en: map.bulk_rename_saved');
assert(t('map.bulk_save') === 'Save All', 'en: map.bulk_save');
assert(t('map.bulk_close') === 'Close', 'en: map.bulk_close');
assert(t('toast.failed') === 'Failed', 'en: map.toast.failed');

setLang('ja');
assert(t('map.bulk_rename') === '一括リネーム', 'ja: map.bulk_rename');
assert(t('map.bulk_rename_title') === 'クライアント名の編集', 'ja: map.bulk_rename_title');
assert(t('map.bulk_rename_saved') === '名前を保存しました', 'ja: map.bulk_rename_saved');
assert(t('map.bulk_save') === 'すべて保存', 'ja: map.bulk_save');
assert(t('map.bulk_close') === '閉じる', 'ja: map.bulk_close');
assert(t('toast.failed') === '失敗しました', 'ja: map.toast.failed');

console.log('ALL MAP TESTS PASSED');
