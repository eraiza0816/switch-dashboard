# Project Structure

## 構成哲学

switch-dashboard は **単一 Go バイナリ + 静的フロントエンド** 構成:
- `cmd/switch-dashboard/` — Go エントリポイント（main.go）
- `internal/` — Go パッケージ（chi ルーター、DuckDB 履歴、スクレイパー等）
- `frontend/` — TypeScript（bun build）→ `static/dist/` に出力
- `templates/` — サーバーサイドテンプレート（`window.__DATA__` 注入）
- `e2e/` — Playwright テスト

```
switch-dashboard/
├── AGENTS.md                    ← エージェント実行手順
├── CLAUDE.md                    ← AIオーケストレータ
├── README.md                    ← 機能・API・アーキテクチャの一次ソース
├── cmd/
│   └── switch-dashboard/
│       └── main.go              ← エントリポイント
├── internal/
│   ├── server/                  ← chi ルーター、HTTPハンドラ（handler_*.go）
│   │   ├── server.go            ← Server構造体、registerRoutes()
│   │   ├── handler_switches.go
│   │   ├── handler_history.go
│   │   └── handler_openapi.go   ← OpenAPI 3.0 定義
│   ├── rtlplayground/           ← スイッチ API スクレイパー
│   ├── poller/                  ← 定期ポーリング
│   ├── config/                  ← 設定ファイル読み込み・検証
│   ├── history/                 ← DuckDB 時系列ストア（live/1h/24h 集約）
│   ├── store/                   ← データ保存
│   ├── oui/                     ← IEEE OUI ベンダー解決
│   └── logbuf/                  ← サーバーログリングバッファ
├── frontend/
│   ├── package.json
│   ├── vitest.config.ts         ← vitest（jsdom, src/**/*.test.ts）
│   └── src/
│       ├── dashboard.ts         ← ダッシュボード画面（DOM + fetch + Chart.js）
│       ├── logs.ts              ← ログビューア
│       ├── backups.ts           ← バックアップ画面
│       ├── map.ts               ← ネットワークトポロジーマップ
│       ├── i18n.ts              ← 多言語対応
│       ├── types.ts             ← 型定義集約（window.__DATA__ 等）
│       └── *.test.ts            ← vitest テスト（ソースと同ディレクトリ）
├── templates/                   ← *.html（window.__DATA__ で初期データ注入）
├── static/
│   ├── style.css
│   └── dist/                    ← bun build 出力先
├── e2e/
│   ├── tests/                   ← Playwright テスト（*.spec.ts）
│   └── playwright.config.ts
├── images/
├── Dockerfile
├── Dockerfile.e2e
└── go.mod
```

## Backend レイヤー

### server/
HTTP層。`Server` 構造体が依存（Logger, Store, Config 等）を保持。
- ハンドラは `func (s *Server) handleAPISwitchXxx(w http.ResponseWriter, r *http.Request)` のシグネチャ
- 1ファイル = 1リソース（`handler_*.go`）
- ルート定義は `server.go` の `registerRoutes()` に集約（chi: `s.Router.Get/Post/Route`）
- エラー時は `http.Error(w, \`{"error":"..."}\`, http.StatusXxx)` でJSONを返す
- 正常時は `w.Header().Set("Content-Type", "application/json"); json.NewEncoder(w).Encode(resp)`

### rtlplayground/
RTLPlayground の JSON エンドポイントへ直接 HTTP 呼び出し（HTML 解析なし）。

### history/
DuckDB 時系列ストア。live / 1h / 24h の集約、1年保持。

### config/
設定ファイル（`config.json`）読み込み・検証。`-c` でパス変更可。

## Frontend レイヤー

### src/
- `dashboard.ts` / `logs.ts` / `backups.ts` / `map.ts` — 画面ごとのモジュール（`bun build` の入力）
- サーバー側テンプレートから `window.__DATA__` で初期データを注入
- グローバル関数は `window.xxx` にアタッチし、`types.ts` で型宣言
- 素の TypeScript + DOM。フレームワーク不使用

## 新しいファイルを追加するルール

### Backend
- 新しいハンドラファイルは `internal/server/` に `handler_*.go` として追加、`registerRoutes()` にルート登録
- 新しいAPIエンドポイントを追加した場合、`handler_openapi.go` と `README.md` の API セクションも更新
- 新しいパッケージは `internal/` 配下に責務単位で追加

### Frontend
- 新しい画面モジュールは `frontend/src/` に追加し、`package.json` の build script とサーバーテンプレートに登録
- テストはソースと同じディレクトリに `*.test.ts` で配置
- 新しい型は `types.ts` に追加
