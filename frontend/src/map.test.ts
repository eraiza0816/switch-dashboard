import { t, setLang, getLang } from './i18n';

setLang('en');

function assert(condition: boolean, msg: string) {
  if (!condition) throw new Error(msg);
  console.log('PASS:', msg);
}

assert(t('map.nickname') === 'Nickname', 'en: map.nickname');
assert(t('btn.save') === 'Save', 'en: btn.save');

console.log('ALL MAP TESTS PASSED');
