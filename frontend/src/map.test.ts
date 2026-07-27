import { describe, it, expect, beforeEach } from 'vitest';
import { t, setLang } from './i18n';

beforeEach(() => {
  setLang('en');
});

describe('map translations', () => {
  it('returns English map translations', () => {
    setLang('en');
    expect(t('map.nickname')).toBe('Nickname');
    expect(t('map.bulk_rename')).toBe('Bulk Rename');
    expect(t('map.bulk_rename_title')).toBe('Edit Client Names');
    expect(t('map.bulk_rename_saved')).toBe('Names saved');
    expect(t('map.bulk_save')).toBe('Save All');
    expect(t('map.bulk_close')).toBe('Close');
    expect(t('toast.failed')).toBe('Failed');
  });

  it('returns Japanese map translations', () => {
    setLang('ja');
    expect(t('map.bulk_rename')).toBe('一括リネーム');
    expect(t('map.bulk_rename_title')).toBe('クライアント名の編集');
    expect(t('map.bulk_rename_saved')).toBe('名前を保存しました');
    expect(t('map.bulk_save')).toBe('すべて保存');
    expect(t('map.bulk_close')).toBe('閉じる');
    expect(t('toast.failed')).toBe('失敗しました');
  });
});
