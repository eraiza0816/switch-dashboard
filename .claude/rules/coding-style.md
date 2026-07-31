# Coding Style

## 不変性 (CRITICAL)

常に新しいオブジェクトを生成し、決して変更しない:

```go
// ❌ WRONG: Mutation
port.Status = "down"

// ✅ CORRECT: Make a copy
updated := port
updated.Status = "down"
```

```typescript
// ❌ WRONG: Mutation
switchData.ports[index].status = newStatus;

// ✅ CORRECT: Immutable update
switchData.ports = switchData.ports.map((p, i) => i === index ? { ...p, status: newStatus } : p);
```

## ファイル編成

**少数の大きなファイルより多数の小さなファイル**:

- 単一責任 — ファイル内の全シンボルが一つの責務を持つ
- 関数は50行未満
- ユーティリティは3ファイルに分割
- 型/機能/ドメインで整理、レイヤー種別では整理しない

## Go 命名規則

- `camelCase` — パッケージ内部変数、非公開関数
- `PascalCase` — エクスポートされた型、関数、メソッド
- `snake_case` — JSONフィールドタグ
- `UPPER_SNAKE_CASE` — パッケージレベルの定数
- ファイル名: `snake_case.go`

### Go のエラーハンドリング

```go
// エラーは必ずチェックして上位に伝播。無視しない
data, err := scrapeSwitch(ctx, ip)
if err != nil {
    return fmt.Errorf("scrape %s: %w", ip, err)
}

// HTTPハンドラではユーザー向けメッセージにラップ
if err != nil {
    http.Error(w, `{"error":"スイッチの取得に失敗しました"}`, http.StatusBadGateway)
    return
}
```

### Go のJSONレスポンスパターン

```go
w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(resp)
```

### Go のルーティング

chi router を使用（`internal/server/server.go` の `registerRoutes()` に集約）:

```go
// ✅ CORRECT: chi メソッドレシーバパターン
s.Router.Route("/api", func(r chi.Router) {
    r.Get("/switches", s.handleAPISwitches)
    r.Get("/switches/{ip}/sfp", s.handleAPISwitchSFP)
    r.Post("/switches/{ip}/cmd", s.handleAPISwitchCmd)
})
```

### Go のハンドラシグネチャ

ハンドラは `Server` 構造体のメソッド。1ファイル = 1リソース（`handler_*.go`）:

```go
func (s *Server) handleAPISwitchCmd(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Command string `json:"cmd"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, `{"error":"bad request"}`, http.StatusBadRequest)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
```

### Go の依存性注入

`Server` 構造体が依存（`Logger`, `Store`, `Config`, `Scraper` 等）をフィールドで保持し、ハンドラは `s.xxx` でアクセスする（現在のパターンに従う）。

## TypeScript 命名規則

- `camelCase` — 変数、関数、ファイル（ユーティリティ）
- `PascalCase` — 型、インターフェース
- `kebab-case` — null
- 命名付きエクスポートを優先

```typescript
// ✅ 命名付きエクスポート
export function formatBytes(n: number): string { ... }
```

## フロントエンド実装パターン

素の TypeScript + DOM を使用。フレームワーク（React/Vue 等）は不使用。

### 型定義の集約

`src/types.ts` に全型を集約。`Window` インターフェースに `window.__DATA__` とグローバル関数を宣言:

```typescript
interface Window {
  __DATA__: {
    refresh: number;
    columns: string[];
    portWrap: number;
  };
  openGraph: (ip: string, port: string, name: string, label: string, cumTX: number, cumRX: number) => void;
}

interface PortState {
  port: string;
  status: string;
  speed: string;
  tx_bytes: number;
  rx_bytes: number;
}
```

### サーバー初期データ

サーバー側テンプレート（`templates/*.html`）から `window.__DATA__` で初期データを注入し、DOM 構築前に読み取る:

```typescript
export const REFRESH_SECONDS = (window.__DATA__ && window.__DATA__.refresh) || 30;
```

### グローバル関数

HTML の `onclick` 等から呼ぶ関数は `window.xxx` にアタッチし、`types.ts` に型宣言:

```typescript
// ✅ 現在のパターン
window.backupConfig = (ip: string, el: HTMLElement) => { ... }
```

### 非同期処理

`async/await` で `.then()` 不使用。`fetch` は素の TypeScript で:

```typescript
const loadSwitches = async () => {
  try {
    const res = await fetch('/api/switches')
    const data = await res.json()
    renderTable(data)
  } catch (e) {
    console.error(e)
  }
}
```

### レンダリング

Chart.js（CDN）と DOM 生成。インライン CSS ダークテーマ。HTML の書き込みは要素の `textContent` を使用し、`innerHTML` にユーザー入力を直接埋め込まない（XSS 対策）。

## 型安全性

```typescript
// 可能な限り any を避ける。新規コードでは適切な型を定義する
interface SwitchData {
  name: string;
  ip: string;
  model: string;
  ports: PortState[];
  mac_table: MACEntry[];
  status: string;
  timestamp: number;
}
```

```go
// 構造体タグは JSON で必須
type PortState struct {
    Port     string `json:"port"`
    Status   string `json:"status"`
    Speed    string `json:"speed"`
    SpeedTX  uint64 `json:"speed_tx_bps"`
}
```

## 責務分割

フレームワーク固有ラッパー（HTTPハンドラ、DOM操作、Chart.js 生成）にビジネスロジックを書かない。ロジックを純粋関数として抽出する。

```go
// ❌ WRONG: Handler 内部にロジック
func (s *Server) handleAPISpeeds(w http.ResponseWriter, r *http.Request) {
    // 帯域計算の処理がハンドラにべったり
}

// ✅ CORRECT: 純粋関数＋ハンドラは薄いアダプタ
func calcSpeedBps(txDelta, rxDelta uint64, elapsed float64) (float64, float64) {
    return float64(txDelta) / elapsed, float64(rxDelta) / elapsed
}

func (s *Server) handleAPISpeeds(w http.ResponseWriter, r *http.Request) {
    tx, rx := calcSpeedBps(...)
    // ...
}
```

```typescript
// ✅ CORRECT: 純粋関数を抽出（テスト可能に）
export function formatBytes(n: number): string {
  if (n >= 1e12) return (n / 1e12).toFixed(2) + ' TB'
  // ...
}
```

## テストファイル配置

テストはソースファイルと同じディレクトリに配置:

```
internal/server/
├── handler_switches.go
└── server_test.go

frontend/src/
├── dashboard-utils.ts
└── dashboard-utils.test.ts
```

## コード品質チェックリスト

完了前に確認:

- [ ] 関数は50行未満
- [ ] ファイルは単一責務
- [ ] 不変更新を使っている
- [ ] エラーハンドリングが適切
- [ ] `console.log` がない（Go: `fmt.Print`/`log.Print` のデバッグ出力がない）
- [ ] Go: JSONタグが全フィールドにある
- [ ] TypeScript: 新規コードで `any` を使っていない（既存コードへの変更は最小限）
- [ ] `innerHTML` にユーザー入力を直接埋め込んでいない
