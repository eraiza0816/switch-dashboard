# E2E テスト

Playwright を使用した E2E (End-to-End) テストです。`e2e/` ディレクトリは `frontend/` から独立しており、`@playwright/test` のみに依存します。

## テスト構成

```
e2e/
├── package.json             # 依存: @playwright/test のみ
├── playwright.config.ts     # Playwright 設定（baseURL: localhost:8081）
├── tests/
│   ├── api.spec.ts          # API エンドポイント 9 項目
│   ├── dashboard.spec.ts    # ダッシュボード画面 8 項目
│   ├── features.spec.ts     # Config/Backups/Logs/API Docs/静的アセット/ナビ 19 項目
│   ├── map.spec.ts          # ネットワークマップ画面 10 項目
│   ├── screenshots.spec.ts  # スクリーンショット生成 4 項目
│   └── pages.spec.ts        # ページ遷移 5 項目（テストアーティファクト）
├── screenshots.ts           # スクリーンショット生成スクリプト（単体実行用）
└── README.md
```

全 **55 テスト**。対象ページ:

| ページ | テスト数 | 内容 |
|--------|---------|------|
| Dashboard (`/`) | 8 | カード表示、ポートテーブル、MACテーブル展開、グラフモーダル、フォントサイズ、デバイス情報、更新時刻 |
| Config (`/config`) | 4 | フォーム表示、スイッチ設定値、言語セレクタ、保存ボタン |
| Backups (`/backups`) | 2 | ページ表示、API応答 |
| Logs (`/logs`) | 7 | ビューワ表示、ログレベル変更・クリア・ダウンロード、API検証 |
| Map (`/map`) | 10 | SVGノード表示、ノード選択（mousedown）、ドラッグ移動、検索、サイドバー、トグル/リセット/一括リネームリンク、言語セレクタ |
| API Docs (`/api-docs`) | 2 | エンドポイント一覧、詳細表示（OpenAPIから動的生成） |
| API | 9 | `/api/switches`, `/api/speeds`, `/api/topology`, `/api/settings`, `/api/notes`, `/api/reset`, `/api/vendors`, `/api/openapi.json` |
| Static assets | 3 | CSS/JS/画像の配信 |
| Navigation | 1 | 全ナビリンク（Dashboard/Map/Config/Backups/Logs）の遷移確認 |

## 実行方法

### ローカル

```bash
# 依存インストール（初回のみ）
cd e2e && npm install && npx playwright install chromium

# Go サーバーを起動（-d でデータディレクトリ指定）
cd /path/to/switch-dashboard
go build -o switch-dashboard ./cmd/switch-dashboard/
mkdir -p /tmp/e2e-data
cat > /tmp/e2e-data/config.json << 'EOF'
{"title":"E2E","refresh_interval":30,"switches":[{"name":"Test","ip":"192.168.10.247","password":"1234","model":"RTLPlayground","port_count":8,"enabled":true}]}
EOF
# -demo: デモモード（モックデータを投入、実スイッチ不要）
./switch-dashboard -d /tmp/e2e-data -demo &

# テスト実行
cd e2e && npx playwright test

# サーバー停止
kill %1; rm -rf /tmp/e2e-data
```

### Docker（推奨）

```bash
# プロジェクトルートで実行
docker build -f Dockerfile.e2e -t switch-dashboard-e2e .
docker run --rm switch-dashboard-e2e
```

1 コンテナでサーバー起動 → テスト実行 → 終了 まで完了。Go のインストール不要。

### スクリーンショット更新

```bash
cd e2e && OUT_DIR=../images npx playwright test tests/screenshots.spec.ts
```

## 注意事項

- Go サーバーは `-demo` フラグでデモモードになり、モックデータを自動生成するため、実際のスイッチは不要
- デモモードで表示されるデータは「デモデータ」として UI 上に明示される（DEMO バッジ）
- 本番モード（`-demo` なし）ではモックデータは生成されず、到達不能なスイッチは offline 表示になる
- DuckDB (`history.duckdb`) が自動作成され、デモモード時はモック履歴データ（過去2分間の帯域）が投入される。本番モードでは推定値（パケット数×800）のサンプルは履歴に記録されない
- スクリーンショット（`images/*.png`）はテスト時のサーバー状態を撮影したものである。`-demo` なしで実スイッチに接続できている場合は実データ、それ以外はデモデータ（UI に DEMO バッジ表示）になる
- Playwright の最新版は `npm install @playwright/test@latest` で更新可能
- テストデータは `-d` で指定したディレクトリに保存され、プロジェクトルートを汚染しない
