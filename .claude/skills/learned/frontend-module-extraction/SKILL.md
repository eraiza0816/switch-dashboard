---
name: frontend-module-extraction
description: 大きな TS モジュールを分割する際の安全手順。循環依存はコールバック/フックで解消、行範囲抽出の落とし穴
---

## Problem

1918 行の `dashboard.ts` を機能別モジュールに分割する際に以下の問題に遭遇した:
- 関数が他モジュールの関数を直接呼ぶと **循環 import**（`handleDrop` → `updateDashboard`）
- ES module の `import { let x }` は**読み取り専用**。import 先で代入すると実行時エラー
- `sed -n 'A,Bp'` で行範囲抽出すると、**ブロック末尾の閉じ括弧を1行欠く**ことがある（構文エラー）
- 複数回の `sed -i 'A,Bd'` 削除は**行番号がズレてファイルを破壊**する

## Solution

1. **循環参照はコールバック/フック登録方式**で解消:
   ```ts
   // 被依存側（table/inputs モジュール）は空実装のフックを export
   let refreshDashboard: () => void = () => {};
   export function setRefreshDashboard(fn: () => void) { refreshDashboard = fn; }
   // エントリ（dashboard.ts）で配線
   setRefreshDashboard(updateDashboard);
   ```
2. **状態は所有権を持つモジュールに置く**。`pollingPaused` を inputs モジュールが export し、dashboard.ts は `setPollingController({ setPaused, refresh })` で配線。import 先から `let` に書き込まない。
3. **抽出はブロック単位で正確に**:
   - 削除する前に抽出（`sed -n` は削除対象の**最終行を含める**ことを確認）
   - 削除は**行番号の大きい順**に実行
   - 抽出後は `tsc` と `awk` のブレースバランスで検証
   - 破損したら `git` の内容を基に Write で書き直す
4. **HTML の `onclick="fn(...)"` は文字列**なので、fn は window グローバル（エントリで再 export）か同じページのグローバルスコープで解決される。モジュール import は不要。

## Example

```ts
// dashboard-inputs.ts（アクション側）
export function setPollingController(c: { setPaused: (p: boolean) => void; refresh: () => void }) {
  setPaused = c.setPaused; refreshDashboard = c.refresh;
}
export function pausePolling() { setPaused(true); }

// dashboard.ts（エントリ）
setPollingController({
  setPaused: (p) => { pollingPaused = p; /* DOM更新 */ },
  refresh: updateDashboard,
});
```

## When to Apply

- 1 画面モジュール（dashboard/logs/map 等）が巨大化し機能別に分割したい
- 分割中にモジュール間の直接参照による循環が発生した
- 大量のコードを正確に別ファイルへ移す必要がある
