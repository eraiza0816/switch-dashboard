# E2E テスト

Playwright を使用した E2E (End-to-End) テストです。

## テスト構成

```
e2e/
├── playwright.config.ts    # Playwright 設定
├── tests/
│   ├── api.spec.ts         # API エンドポイント 9 項目
│   ├── dashboard.spec.ts   # ダッシュボード画面 8 項目
│   ├── features.spec.ts    # Config/Backups/Logs/API Docs/静的アセット/ナビ 19 項目
│   ├── map.spec.ts         # ネットワークマップ画面 10 項目
│   ├── screenshots.spec.ts # スクリーンショット生成 4 項目
│   └── pages.spec.ts       # ページ遷移 5 項目（テストアーティファクト）
├── screenshots.ts          # スクリーンショット生成スクリプト（単体実行用）
└── README.md
```

全 **51 テスト**。対象ページ:

- Dashboard (`/`) — カード表示、ポートテーブル、MACテーブル展開、グラフモーダル、フォントサイズ切替
- Config (`/config`) — フォーム表示、スイッチ設定、言語セレクタ、保存ボタン
- Backups (`/backups`) — ページ表示、API応答
- Logs (`/logs`) — ビューワ表示、ログレベル変更、クリア、ダウンロード
- Map (`/map`) — SVGノード表示、ノード選択、ドラッグ移動、検索、サイドバー
- API Docs (`/api-docs`) — エンドポイント一覧、詳細表示
- API — 9 エンドポイントのスモークテスト
- Static assets — CSS/JS/画像の配信確認
- Navigation — 全ナビリンクの遷移確認

## 実行方法

### ローカル

```bash
# Go サーバーを起動（config.json 必須）
cd /path/to/switch-dashboard
go build -o switch-dashboard ./cmd/switch-dashboard/
echo '{"title":"E2E","refresh_interval":30,"switches":[{"name":"Test","ip":"192.168.10.247","password":"1234","model":"RTLPlayground","port_count":8,"enabled":true}]}' > config.json
./switch-dashboard &

# テスト実行
cd frontend
npx playwright test --config=e2e/playwright.config.ts

# サーバー停止
kill %1
```

### Docker

```bash
# ビルド
docker build -f Dockerfile.e2e -t switch-dashboard-e2e .

# 実行（テスト後コンテナ自動削除）
docker run --rm switch-dashboard-e2e
```

### スクリーンショット更新

```bash
# テスト経由
cd frontend
OUT_DIR=../images npx playwright test --config=e2e/playwright.config.ts tests/screenshots.spec.ts
```

## 注意事項

- Go サーバーはモックデータを自動生成するため、実際のスイッチは不要
- DuckDB (`history.duckdb`) が自動作成され、モック履歴データが投入される
- Playwright の最新版は `npm install @playwright/test@latest` で更新可能
