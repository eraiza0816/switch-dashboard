# switch-dashboard — AGENTS.md

## Stack

- **Backend**: Go 1.26+, chi router, single binary, DuckDB（帯域履歴保存）
- **Frontend**: TypeScript（`bun build`）、Chart.js（CDN）、vitest、インライン CSS ダークテーマ
- **E2E**: Playwright（`e2e/`）
- **対象**: RTLPlayground ファームウェア搭載スイッチ（RTL8372/RTL8373）

## Knowledge Base

### Tiers

| Directory | Scope | Loaded |
|---|---|---|
| `/.claude/rules/` | Repo全体 | 常時ロード |
| `/.claude/commands/` | `/` コマンド | オンデマンド |
| `/.claude/skills/` | スキル | Skill Tool 経由 |
| `/.claude/skills/learned/` | 学習済みパターン | 自動呼び出し |

### 常時ロードルール

| Rule | Description |
|---|---|
| **coding-style** | Go + TypeScript コーディング規約、不変性、命名規則 |
| **project-structure** | ディレクトリ構成、レイヤー責務、ファイル追加ルール |
| **security** | セキュリティチェックリスト、コマンド注入・パストラバーサル対策 |
| **testing** | テスト配置、Go / vitest / Playwright テストパターン |

## Run

```sh
# フロントエンドをビルド
cd frontend && bun install && bun run build && cd ..

# バイナリをビルド
go build -o switch-dashboard ./cmd/switch-dashboard/

# 実行（データは ~/.local/share/switch-dashboard/ に保存、:8081 で待受）
./switch-dashboard

# カレントディレクトリにデータ保存
./switch-dashboard -d .
```

- 設定ファイル: `~/.local/share/switch-dashboard/config.json`（`-c` で変更可）
- フロントエンドのビルド出力先は `static/dist/`

## Backend

### アーキテクチャ

```
User Browser (TypeScript + Chart.js)
        ↕  HTTP JSON API
Go Server (chi router, single binary, DuckDB history)
        ↕  HTTP JSON API
RTLPlayground Switch (uIP embedded webserver)
```

- **ルーター**: chi（`s.Router.Get/Post/Route`）。エントリポイントは `cmd/switch-dashboard/main.go`、ルート定義は `internal/server/server.go` の `registerRoutes()`
- **スクレイパー**: RTLPlayground の JSON エンドポイントに直接 HTTP 呼び出し（HTML 解析なし）
- **履歴**: DuckDB 時系列ストア（`internal/history/`）、live / 1h / 24h 集約、1年保持

### ハンドラ規約

- ハンドラは `Server` 構造体のメソッド（`internal/server/`）
  ```go
  func (s *Server) handleAPISwitchCmd(w http.ResponseWriter, r *http.Request)
  ```
- 1ファイル = 1リソース（`handler_*.go`）
- 正常時は `w.Header().Set("Content-Type", "application/json"); json.NewEncoder(w).Encode(resp)`
- エラー時は `http.Error(w, \`{"error":"..."}\`, http.StatusXxx)` でJSONを返す

### API

- 主要エンドポイントは `README.md` の API セクションを参照
- OpenAPI 仕様: `GET /api/openapi.json`（定義は `internal/server/handler_openapi.go`）

## Frontend

### Run

```sh
cd frontend
bun run build      # ../static/dist/ へ出力（dashboard.ts / logs.ts / backups.ts / map.ts）
bun run watch      # --watch ビルド
bun run typecheck  # bun tsc --noEmit
bun run test       # vitest run
```

### Structure

- `src/dashboard.ts` — ダッシュボード画面（DOM + fetch + Chart.js）
- `src/logs.ts` — ログビューア
- `src/backups.ts` — バックアップ画面
- `src/map.ts` — ネットワークトポロジーマップ
- `src/i18n.ts` — 多言語対応
- `src/types.ts` — 型定義集約（`Window` インターフェースに `window.__DATA__` とグローバル関数を宣言）
- `src/*.test.ts` — vitest テスト（ソースと同じディレクトリに配置）

### フロントエンド規約

- サーバー側テンプレート（`templates/*.html`）から `window.__DATA__` で初期データを注入
- グローバル関数は `window.xxx` にアタッチし、HTML の `onclick` 等から呼ぶ（`types.ts` で型宣言）
- 実装は素の TypeScript + DOM。フレームワーク（React/Vue 等）は不使用

## Tests

### Run all

```sh
# Go ユニットテスト（internal/ 配下、ソースと同ディレクトリに *_test.go）
go test ./...

# フロントエンド単体テスト（vitest, jsdom）
cd frontend && bun run test

# 型チェック
cd frontend && bun run typecheck

# E2E（Playwright、サーバー起動済みであること）
cd e2e && npx playwright test

# E2E（Docker 単体）
docker build -f Dockerfile.e2e -t switch-dashboard-e2e . && docker run --rm switch-dashboard-e2e
```

### Test structure

- `internal/**/*_test.go` — Go ユニットテスト（`server_test.go` など）
- `frontend/src/*.test.ts` — vitest（jsdom 環境）
- `e2e/tests/*.spec.ts` — Playwright E2E（`screenshots.spec.ts` は画像更新用）

## Before Committing

- [ ] `gofmt` が通っている（PostToolUse hook で自動実行）
- [ ] `.claude/rules/security.md` のチェックリストを満たしている
- [ ] `go build -o switch-dashboard ./cmd/switch-dashboard/` がエラーなく通る
- [ ] `go test ./...` が通っている
- [ ] `cd frontend && bun run typecheck && bun run test` が通っている
- [ ] 新しいAPIエンドポイントを追加した場合、`README.md` の API セクションと OpenAPI 定義（`handler_openapi.go`）を更新した
- [ ] フロントエンドに新しい機能を追加した場合、既存のパターン（`window.__DATA__`、素のTS + DOM）に従っている
