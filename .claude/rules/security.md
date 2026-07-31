# Security Guidelines

## マンデートリチェック（コミット前）

- [ ] ハードコードされた秘密情報がない（APIキー、パスワード、トークン）
- [ ] 全ユーザー入力のバリデーションとサニタイズ
- [ ] コマンド注入対策（CLIコマンド実行・外部プロセスへのユーザー入力受け渡し）
- [ ] パストラバーサル対策（ファイル名・パス指定の検証、`path.Clean` 使用）
- [ ] XSS防止（`innerHTML` への未検証入力を防ぐ、`textContent` 使用）
- [ ] アップロードの検証（マジックバイト、サイズ制限）
- [ ] エラーメッセージで機密情報を漏洩しない

## シークレット管理

```go
// NEVER: ハードコードされた秘密
const password = "1234"

// ALWAYS: 環境変数 or config から読み込む
password := os.Getenv("SWITCH_PASSWORD")
```

**認識必須のシークレット箇所**:

| 箇所 | 内容 |
|---|---|
| `config.json` | 各スイッチの管理パスワード（`switches[].password`） |

- `config.json` はスイッチの管理パスワードを含むため、リポジトリにコミットしない（`.gitignore` 済み）こと
- バイナリに秘密を焼き込まない。実行時データ（`~/.local/share/switch-dashboard/`）に保存する

## switch-dashboard 固有のセキュリティルール

### コマンド実行（`POST /api/switches/:ip/cmd`）

CLI コマンド実行は RTLPlayground の制約内で行う。シェルを経由せず、許可されたコマンドのみを送信する。

```go
// ❌ WRONG: ユーザー入力をシェルで実行
out, _ := exec.Command("sh", "-c", "cli "+req.Command).CombinedOutput()

// ✅ CORRECT: シェルを経由しない（exec.Command の引数で直接渡す）
out, err := exec.Command(cliBin, "cmd", req.Command).CombinedOutput()
if err != nil {
    http.Error(w, `{"error":"command failed"}`, http.StatusInternalServerError)
    return
}
```

- ユーザー入力をそのまま OS コマンドに連結しない
- コマンド名・引数は引数リストで渡す（`sh -c` 不使用）

### パストラバーサル（`/api/backups/:filename/download` 等）

ファイル名・パスを受け取るエンドポイントでは、ファイル名を正規化して安全なディレクトリ内に閉じ込める。

```go
// ✅ CORRECT: path.Base + 基準ディレクトリに制限
name := path.Base(r.PathValue("filename"))
full := filepath.Join(backupDir, name)
```

- `path.Base` でディレクトリ部分を除去し、基準ディレクトリ配下であることを確認する
- `..` や絶対パス、ヌルバイトを拒否する

### ファームウェア・設定アップロード

- ファイルサイズ制限を設ける（巨大ファイル拒否）
- 内容のマジックバイト検証（画像等）を実施
- 保存先は固定ディレクトリとし、元のファイル名をそのままパスに使わない

### 認証・ネットワーク

- 本アプリは認証なしのローカルサービス（:8081 待受）であることを認識する
- インターネット公開時はリバースプロキシ等でアクセス制限を行うこと（本番要件として扱う）

### XSS

- フロントエンドで HTML を動的生成する際、`innerHTML` にユーザー入力を直接埋め込まない
- 表示値は `textContent` / `escapeHTML` を使用する

## セキュリティ対応プロトコル

問題発見時:
1. 直ちに作業を停止
2. 重大な問題は先に修正
3. 露出した秘密情報はローテーション

## インシデント報告

セキュリティ問題はコードコメントに残さず、ユーザーに報告すること。
