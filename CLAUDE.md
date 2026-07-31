@AGENTS.md

# switch-dashboard — Agentic Development Workflow

## Source of Truth

- `README.md` — 機能一覧・API定義・アーキテクチャ・プロジェクト構成
- `AGENTS.md` — エージェント向け実行手順・スタック情報
- `.claude/rules/` — 常時ロードされるコーディング規約・セキュリティルール

## Workflow

1. **要件確認**: 機能追加/変更時はまず `README.md`（API・アーキテクチャ）と `AGENTS.md` を確認
2. **実装**: 既存コードパターンに従う。`AGENTS.md` のBackend/Frontendセクションを参照
3. **レビュー前チェック**: `.claude/rules/security.md` のチェックリストを満たしていることを確認

## 設計ドキュメントの更新

実装によって API（エンドポイント・レスポンス形式）が変わった場合は、`README.md` の API セクションと OpenAPI 定義（`internal/server/handler_openapi.go`）を同時に更新すること。
