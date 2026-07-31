# Testing Rules

## テストファイル配置

テストはソースファイルと同じディレクトリに配置する:

```
internal/server/
├── server.go
└── server_test.go          # Go テスト

frontend/src/
├── dashboard-utils.ts
└── dashboard-utils.test.ts # vitest テスト

e2e/tests/
└── dashboard.spec.ts       # Playwright E2E
```

## Go テスト

標準 `testing` パッケージを使用（`go test ./...` で実行）。サードパーティ不要。

テストヘルパーは現在のパターンに従う（`server_test.go` の `newTestServer()` 等）:

```go
package server

import (
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestAPISwitches(t *testing.T) {
    s := newTestServer()

    req := httptest.NewRequest("GET", "/api/switches", nil)
    w := httptest.NewRecorder()
    s.handleAPISwitches(w, req)

    if w.Code != http.StatusOK {
        t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
    }
}
```

### テスト命名規則

```go
func TestHandlerName_Condition(t *testing.T) { ... }
// 例: TestAPISwitches / TestAPIHistory_UnknownRange / TestConfigSave_InvalidJSON
```

### AAA（Arrange-Act-Assert）

```go
func TestFormatBytes(t *testing.T) {
    // Arrange
    input := uint64(1000)

    // Act
    got := formatBytes(input)

    // Assert
    if got != "1.00 KB" {
        t.Errorf("got %q, want %q", got, "1.00 KB")
    }
}
```

## フロントエンド単体テスト（vitest）

`frontend/src/*.test.ts` に配置。jsdom 環境（`vitest.config.ts`）。

```typescript
import { describe, it, expect } from 'vitest';
import { formatBytes } from './dashboard-utils';

describe('formatBytes', () => {
  it('returns KB', () => {
    expect(formatBytes(1000)).toBe('1.00 KB');
  });
});
```

実行:

```sh
cd frontend && bun run test       # vitest run
cd frontend && bun run typecheck  # bun tsc --noEmit
```

## E2E（Playwright）

`e2e/tests/*.spec.ts` に配置。サーバー起動済みで実行:

```sh
cd e2e && npx playwright test
```

- `screenshots.spec.ts` はスクリーンショット更新用（`OUT_DIR=../images npx playwright test tests/screenshots.spec.ts`）
- 単体テストではカバーできない画面遷移・実スイッチとの疎通を確認する

## テスト駆動パターン

- 純粋関数（`formatBytes`、`calcSpeedBps`、履歴集約など）は単体テストで検証する
- ハンドラは `httptest` でレスポンスコード・JSONボディを検証する
- DOM を直接テストする場合は jsdom 環境（vitest）を使用する
