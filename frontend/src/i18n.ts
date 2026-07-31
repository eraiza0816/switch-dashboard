export type Lang = 'en' | 'ja';

const STORAGE_KEY = 'dashboard_lang';

const translations: Record<Lang, Record<string, string>> = {
  en: {
    'nav.dashboard': 'Dashboard',
    'nav.map': 'Map',
    'nav.config': 'Config',
    'nav.backups': 'Backups',
    'nav.logs': 'Logs',
    'nav.api': 'API',

    'status.live': 'Live',
    'status.paused': 'Paused',
    'status.update_in': 'update in {n}s',
    'status.online': 'online',
    'status.offline': 'offline',
    'status.up': 'up',
    'status.down': 'down',
    'status.disable': 'disable',
    'status.link_up': 'Link Up',
    'status.link_down': 'Link Down',
    'status.disabled': 'Disabled',

    'port.port': 'Port',
    'port.status': 'Status',
    'port.speed': 'Speed',
    'port.duplex': 'Duplex',
    'port.packets': 'Packets',
    'port.bytes': 'Bytes',
    'port.info': 'Info',
    'port.notes': 'Notes',

    'mac.vendor': 'Vendor',
    'mac.mac': 'MAC',
    'mac.type': 'Type',
    'mac.port': 'Port',
    'mac.vlan': 'VLAN',
    'mac.host': 'Host',
    'mac.entries': '{n} entries',
    'mac.table': 'MAC Address Table',
    'mac.no_entries': 'No MAC entries found',
    'mac.search': 'Search MAC table...',

    'btn.backup': 'Backup',
    'btn.console': 'Console',
    'btn.save': 'Save',
    'btn.cancel': 'Cancel',
    'btn.delete': 'Delete',
    'btn.download': 'Download',
    'btn.clear': 'Clear',
    'btn.yes': 'Yes',
    'btn.config_save': 'Save Configuration',

    'sfp.title': 'Port {p} SFP+ Transceiver Status',
    'sfp.subtitle': 'DDMI Diagnostics & Telemetry ({name})',
    'sfp.vendor': 'Vendor',
    'sfp.model': 'Model',
    'sfp.serial': 'Serial',
    'sfp.type': 'Type',
    'sfp.temperature': 'Temperature',
    'sfp.voltage': 'Voltage',
    'sfp.bias': 'Bias Current',
    'sfp.tx_power': 'TX Power',
    'sfp.rx_power': 'RX Power',
    'sfp.loading': 'Loading transceiver data...',
    'sfp.no_module': 'No SFP module detected',
    'sfp.note': 'DDMI telemetry reads directly from SFP module registers.',

    'graph.title': 'Bandwidth - {label} ({name})',
    'graph.live': 'Live',
    'graph.1h': '1 Hour',
    'graph.24h': '24 Hours',
    'graph.peak_tx': 'Peak TX',
    'graph.peak_rx': 'Peak RX',
    'graph.current_tx': 'Current TX',
    'graph.current_rx': 'Current RX',
    'graph.total_tx': 'Total TX',
    'graph.total_rx': 'Total RX',
    'graph.no_data': 'No history data available for this range yet',
    'graph.load_failed': 'Failed to load history',
    'graph.cumulative': 'Cumulative traffic - updated every {n}s',

    'backup.title': 'Configuration Backups',
    'backup.no_backups': 'No backups yet',
    'backup.description': 'Trigger a backup from the switch card on the dashboard.',
    'backup.confirm_delete': 'Delete backup {name}?',
    'backup.saved': 'Backed up!',
    'backup.failed': 'Failed',

    'logs.title': 'System Logs',
    'logs.no_logs': 'No logs yet',
    'logs.waiting': 'Logs will appear here as the server runs.',
    'logs.level': 'Log Level',
    'logs.auto_scroll_paused': 'Auto-scroll paused - click to resume',

    'config.title': 'Settings',
    'config.general': 'General',
    'config.title_label': 'Dashboard Title',
    'config.refresh': 'Refresh Interval (seconds)',
    'config.switches': 'Switches',
    'config.name': 'Name',
    'config.ip': 'IP Address',
    'config.password': 'Password',
    'config.model': 'Model',
    'config.saved': 'Configuration saved.',

    'map.title': 'Network Map',
    'map.switch': 'Switch',
    'map.client': 'Client',
    'map.internet': 'Internet',
    'map.router': 'Router',
    'map.repeater': 'Repeater',
    'map.unmanaged_switch': 'Unmanaged Switch',
    'map.search_placeholder': 'Search nodes...',
    'map.details': 'Details',
    'map.links': 'Links',
    'map.nickname': 'Nickname',
    'map.toggle_clients': 'Toggle Clients',
    'map.reset_layout': 'Reset Layout',
    'map.bulk_rename': 'Bulk Rename',
    'map.bulk_rename_title': 'Edit Client Names',
    'map.bulk_rename_saved': 'Names saved',
    'map.bulk_save': 'Save All',
    'map.bulk_close': 'Close',
    'map.failed_load': 'Failed to load topology',
    'map.loading': 'Loading topology...',
    'map.ip': 'IP',
    'map.vendor': 'Vendor',
    'map.device_type': 'Device Type',
    'map.forget': 'Forget Device',
    'map.forget_confirm': 'Are you sure you want to forget and remove this client device?',
    'map.last_location': 'Last Location',
    'map.last_active': 'Last Active',
    'map.import_csv': 'Import CSV',
    'map.import_csv_done': 'Imported {n} clients',
    'map.import_csv_failed': 'CSV import failed',
    'map.refresh_in': 'refresh in {n}s',
    'map.live': 'Live',
    'map.legend_port': 'PORT STATUS',
    'map.legend_down': 'Down',
    'map.legend_disabled': 'Disabled',

    'footer.back': 'Back to Dashboard',
    'footer.api': 'API Reference',
    'footer.version': 'v{ver}',

    'confirm.reset_title': 'Reset All Counters',
    'confirm.reset_msg': 'This will reset all cumulative port counters. This action cannot be undone.',
    'confirm.reboot_title': 'Reboot Switch',
    'confirm.reboot_msg': 'Are you sure you want to reboot {name}? This will cause network interruption.',

    'console.prompt': 'Enter CLI command (e.g. show, port 5 1g):',
    'console.response': 'Response: {msg}',
    'console.error': 'Error: {msg}',
    'console.sent': 'Command sent',

    'toast.saved': 'Note saved',
    'toast.failed': 'Failed',
    'toast.reset': 'Counters reset',
    'toast.reboot': 'Reboot command sent',
    'toast.backup': 'Backup saved to server',
  },

  ja: {
    'nav.dashboard': 'ダッシュボード',
    'nav.map': 'ネットワークマップ',
    'nav.config': '設定',
    'nav.backups': 'バックアップ',
    'nav.logs': 'ログ',
    'nav.api': 'API',

    'status.live': 'ライブ',
    'status.paused': '一時停止',
    'status.update_in': 'あと{n}秒',
    'status.online': 'オンライン',
    'status.offline': 'オフライン',
    'status.up': 'アップ',
    'status.down': 'ダウン',
    'status.disable': '無効',
    'status.link_up': 'リンクアップ',
    'status.link_down': 'リンクダウン',
    'status.disabled': '無効',

    'port.port': 'ポート',
    'port.status': '状態',
    'port.speed': '速度',
    'port.duplex': '双方向',
    'port.packets': 'パケット',
    'port.bytes': 'バイト',
    'port.info': '情報',
    'port.notes': 'メモ',

    'mac.vendor': 'ベンダー',
    'mac.mac': 'MAC',
    'mac.type': 'タイプ',
    'mac.port': 'ポート',
    'mac.vlan': 'VLAN',
    'mac.host': 'ホスト',
    'mac.entries': '{n}件',
    'mac.table': 'MACアドレステーブル',
    'mac.no_entries': 'MACエントリーがありません',
    'mac.search': 'MACテーブルを検索...',

    'btn.backup': 'バックアップ',
    'btn.console': 'コンソール',
    'btn.save': '保存',
    'btn.cancel': 'キャンセル',
    'btn.delete': '削除',
    'btn.download': 'ダウンロード',
    'btn.clear': 'クリア',
    'btn.yes': 'はい',
    'btn.config_save': '設定を保存',

    'sfp.title': 'ポート{p} SFP+ トランシーバー診断',
    'sfp.subtitle': 'DDMI テレメトリー ({name})',
    'sfp.vendor': 'ベンダー',
    'sfp.model': '型番',
    'sfp.serial': 'シリアル',
    'sfp.type': '種類',
    'sfp.temperature': '温度',
    'sfp.voltage': '電圧',
    'sfp.bias': 'バイアス電流',
    'sfp.tx_power': 'TX 光出力',
    'sfp.rx_power': 'RX 光入力',
    'sfp.loading': 'トランシーバーデータを読み込み中...',
    'sfp.no_module': 'SFP モジュールが検出されませんでした',
    'sfp.note': 'DDMI テレメトリーは SFP モジュールのレジスターから直接読み取っています。',

    'graph.title': '帯域 - {label} ({name})',
    'graph.live': 'ライブ',
    'graph.1h': '1時間',
    'graph.24h': '24時間',
    'graph.peak_tx': '最大TX',
    'graph.peak_rx': '最大RX',
    'graph.current_tx': '現在TX',
    'graph.current_rx': '現在RX',
    'graph.total_tx': '合計TX',
    'graph.total_rx': '合計RX',
    'graph.no_data': 'この範囲の履歴データはまだありません',
    'graph.load_failed': '履歴の読み込みに失敗しました',
    'graph.cumulative': '累計トラフィック - {n}秒ごとに更新',

    'backup.title': '設定バックアップ',
    'backup.no_backups': 'バックアップはまだありません',
    'backup.description': 'ダッシュボードのスイッチカードからバックアップを実行してください。',
    'backup.confirm_delete': 'バックアップ {name} を削除しますか？',
    'backup.saved': 'バックアップ完了！',
    'backup.failed': '失敗',

    'logs.title': 'システムログ',
    'logs.no_logs': 'ログはまだありません',
    'logs.waiting': 'サーバーの稼働とともにログが表示されます。',
    'logs.level': 'ログレベル',
    'logs.auto_scroll_paused': '自動スクロール停止 - クリックで再開',

    'config.title': '設定',
    'config.general': '一般',
    'config.title_label': 'ダッシュボードタイトル',
    'config.refresh': '更新間隔（秒）',
    'config.switches': 'スイッチ',
    'config.name': '名前',
    'config.ip': 'IPアドレス',
    'config.password': 'パスワード',
    'config.model': '機種',
    'config.saved': '設定を保存しました。',

    'map.title': 'ネットワークマップ',
    'map.switch': 'スイッチ',
    'map.client': 'クライアント',
    'map.internet': 'インターネット',
    'map.router': 'ルーター',
    'map.repeater': 'リピーター',
    'map.unmanaged_switch': 'アンマネージドスイッチ',
    'map.search_placeholder': 'ノードを検索...',
    'map.details': '詳細',
    'map.links': 'リンク',
    'map.nickname': 'ニックネーム',
    'map.toggle_clients': 'クライアント表示切替',
    'map.reset_layout': 'レイアウトリセット',
    'map.bulk_rename': '一括リネーム',
    'map.bulk_rename_title': 'クライアント名の編集',
    'map.bulk_rename_saved': '名前を保存しました',
    'map.bulk_save': 'すべて保存',
    'map.bulk_close': '閉じる',
    'map.failed_load': 'トポロジーの読み込みに失敗しました',
    'map.loading': 'トポロジーを読み込み中...',
    'map.ip': 'IP',
    'map.vendor': 'ベンダー',
    'map.device_type': 'デバイスタイプ',
    'map.forget': 'このデバイスを削除',
    'map.forget_confirm': 'このクライアントデバイスを削除してもよろしいですか？',
    'map.last_location': '最終接続位置',
    'map.last_active': '最終アクティブ',
    'map.import_csv': 'CSVインポート',
    'map.import_csv_done': '{n}件のクライアントをインポートしました',
    'map.import_csv_failed': 'CSVインポートに失敗しました',
    'map.refresh_in': 'あと{n}秒で更新',
    'map.live': 'ライブ',
    'map.legend_port': 'ポート状態',
    'map.legend_down': 'ダウン',
    'map.legend_disabled': '無効',

    'footer.back': 'ダッシュボードに戻る',
    'footer.api': 'APIリファレンス',
    'footer.version': 'v{ver}',

    'confirm.reset_title': '全カウンターリセット',
    'confirm.reset_msg': '全ての累積ポートカウンターをリセットします。この操作は元に戻せません。',
    'confirm.reboot_title': 'スイッチ再起動',
    'confirm.reboot_msg': '{name} を再起動してもよろしいですか？ネットワークが一時的に停止します。',

    'console.prompt': 'CLIコマンドを入力 (例: show, port 5 1g):',
    'console.response': '応答: {msg}',
    'console.error': 'エラー: {msg}',
    'console.sent': 'コマンドを送信しました',

    'toast.saved': 'メモを保存しました',
    'toast.failed': '失敗しました',
    'toast.reset': 'カウンターをリセットしました',
    'toast.reboot': '再起動コマンドを送信しました',
    'toast.backup': 'バックアップを保存しました',
  },
};

function getBrowserLang(): Lang {
  try {
    if (typeof navigator === 'undefined') return 'en';
    const lang = (navigator.language || '').slice(0, 2);
    if (lang === 'ja') return 'ja';
  } catch {}
  return 'en';
}

function loadLang(): Lang {
  try {
    if (typeof localStorage === 'undefined') return getBrowserLang();
    const stored = localStorage.getItem(STORAGE_KEY);
    if (stored === 'ja' || stored === 'en') return stored;
  } catch {}
  return getBrowserLang();
}

let currentLang: Lang = loadLang();

export function getLang(): Lang {
  return currentLang;
}

export function setLang(lang: Lang) {
  currentLang = lang;
  try { if (typeof localStorage !== 'undefined') localStorage.setItem(STORAGE_KEY, lang); } catch {}
  if (typeof document !== 'undefined') document.documentElement.lang = lang;
  updateUILang();
}

export function t(key: string, params?: Record<string, string | number>): string {
  const dict = translations[currentLang] || translations.en;
  let msg = dict[key] || translations.en[key] || key;
  if (params) {
    for (const [k, v] of Object.entries(params)) {
      msg = msg.replace(`{${k}}`, String(v));
    }
  }
  return msg;
}

function updateUILang() {
  if (typeof document === 'undefined') return;
  document.querySelectorAll('[data-i18n]').forEach(el => {
    const key = el.getAttribute('data-i18n');
    if (key) el.textContent = t(key);
  });
  document.querySelectorAll('[data-i18n-title]').forEach(el => {
    const key = el.getAttribute('data-i18n-title');
    if (key) el.setAttribute('title', t(key));
  });
  document.querySelectorAll('[data-i18n-placeholder]').forEach(el => {
    const key = el.getAttribute('data-i18n-placeholder');
    if (key) (el as HTMLInputElement).placeholder = t(key);
  });
}

// Run on load (skip in test environments)
if (typeof document !== 'undefined') {
  document.addEventListener('DOMContentLoaded', () => {
    setLang(currentLang);
  });
}
