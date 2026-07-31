---
name: go-import-cycle-refactor
description: Go パッケージ間の重複コードを統合する際、import cycle を避けるレイヤー設計（leaf パッケージへの移動・依存方向の確認）
---

## Problem

`internal/poller` が `internal/server` を import している（`server.Cache` 等を参照）。このため `internal/server` から `poller` の関数を使おうとすると **import cycle** になる。

実際の重複例（switch-dashboard）:
- `poller.parseHex` == `handler_config_save.parseHexVal`（完全同一）
- `poller.linkSpeedToString` == `handler_config_save.linkSpeedFromInt`（完全同一）
- ステータスマッピング（up/down/disable）が poll 側と config_save 側で二重実装

## Solution

1. **依存方向を先に確認**する。`go list -deps` や既存 import から「誰が誰を参照しているか」を把握。
2. **低レベル変換は leaf パッケージへ**。`rtlplayground`（誰も import しない leaf）に `ParseHex` / `LinkSpeedString` を export し、poller と server の両方から参照。
3. **両側が参照可能な場所に共通ヘルパーを置く**:
   - 変換先の型を所有するパッケージに置く（`server.PortStatus(link, enabled)` は `server.PortState` を作る側なので server に置き、poller が使う）
   - server は poller を import できず、poller は server を import 済み → **server に置いたヘルパーは poller から使える**が、逆は不可
4. テストは移動先パッケージへ一緒に移す（`convert_test.go`）。

## Example

```go
// internal/rtlplayground/convert.go（leaf パッケージ、両者から参照可）
func ParseHex(s string) int64      { ... }
func LinkSpeedString(link int) string { ... }

// internal/server/cache.go（server.PortState の所有者）
func PortStatus(link, enabled int) (status, linkStr, duplex string) { ... }

// poller.go と handler_config_save.go の両方で使用
speedStr := rtlplayground.LinkSpeedString(entry.Link)
statusStr, linkStr, duplex := server.PortStatus(entry.Link, entry.Enabled)
```

## When to Apply

- 2 パッケージ以上で同一の変換/マッピング関数が重複している
- 一方のパッケージが他方を import していて循環の恐れがある
- レイヤー整理（server が raw scrape 処理を持ちすぎ）をしたい
