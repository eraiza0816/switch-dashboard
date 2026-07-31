---
name: learn
description: /learn — 開発セッションから再利用可能パターンを抽出し、.claude/skills/learned/ に保存
---

# /learn — Pattern Extraction

## 目的

開発中に解決した非自明な問題を再利用可能なスキルとして保存する。

## 抽出対象

1. **エラー解決** — エラーの原因、修正方法、類似問題への適用
2. **デバッグ手法** — 非自明な診断手順
3. **ワークアラウンド** — ライブラリの制限事項、バージョン固有の解決策
4. **プロジェクトパターン** — コードベースの慣習、設計判断（RTLPlayground API、DuckDB 集約、Chart.js レンダリングなど）

## 保存先

`.claude/skills/learned/{topic-name}/SKILL.md`

## テンプレート

```yaml
---
name: descriptive-name
description: 簡潔な説明（関連コード作業時に自動呼び出し）
---

## Problem
[問題]

## Solution
[解決方法]

## Example
[コード例]

## When to Apply
[適用条件]
```
