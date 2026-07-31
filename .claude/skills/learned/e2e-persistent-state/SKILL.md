---
name: e2e-persistent-state
description: E2E テストが再実行でフレークになる原因として、サーバーが永続化する状態（layout_positions 等）の累積を疑う診断手順
---

## Problem

Playwright テストで「前回までは全部通っていたのに、ある実行から1テストだけ連続失敗」する状況。今回の例:
- `map.spec.ts` の「drag-and-drop moves a node」が失敗
- ノードを +200px ドラッグしても `movedX > 50` にならない
- リファクタはそのページ（map.js）に一切触れていない

## Solution

1. **そのテストが依存する永続状態を疑う**。サーバーは `layout_positions.json` にノード位置を保存する。テストがドラッグを繰り返すたびに位置が累積し、ノードが画面端まで到達するとドラッグが効かなくなる。
2. **単独再実行とフル再実行で「連続失敗」を確認**（1回限りのフレークと区別）。
3. **状態ディレクトリをリセットして再実行**:
   ```sh
   rm -rf /tmp/e2e-data && mkdir -p /tmp/e2e-data   # 設定も再作成
   # サーバー再起動 → テスト再実行 → 通れば状態汚染が原因と確定
   ```
4. 失敗が**コード変更と無関係**であることを示す: 対象ページが読み込む bundle（map.js）が未変更であること、`git diff` で対象ファイルが無いことを確認。
5. 恒久対策: E2E は実行ごとにデータディレクトリを新規作成するか、テスト冒頭で永続状態（layout positions 等）をリセットする。

## Example

```sh
# 疑わしい永続ファイルを確認
cat /tmp/e2e-data/layout_positions.json   # {"192.168.10.247":{"x":770,"y":700}} ← 端到達
# リセットして再実行
pkill -f switch-dashboard; rm -rf /tmp/e2e-data; mkdir -p /tmp/e2e-data
# config.json を再作成 → サーバー起動 → npx playwright test
```

## When to Apply

- E2E が再実行でフレークし、対象コードを触っていないのに失敗する
- テストがドラッグ・位置・設定の保存・再読み込みを含む
- データディレクトリに `.json`（positions / clients / settings 等）が累積している
