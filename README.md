# Switch Dashboard

**RTLPlayground** ファームウェア（RTL8372/RTL8373 ベースの 2.5GbE スイッチ）を搭載したスイッチのためのリアルタイム監視ダッシュボード。Go + TypeScript 製、シングルバイナリで動作します。

![Dashboard](https://raw.githubusercontent.com/eraiza0816/switch-dashboard/refs/heads/main/images/dashboard.png)

## 機能

- **リアルタイムポート状態**: リンク状態、速度、 duplex、TX/RX カウンター・パケット数
- **帯域チャート**: ライブ / 1時間 / 24時間 のロールング履歴（Chart.js + DuckDB）
- **SFP+ DDMI**: 温度、電圧、バイアス電流、TX/RX パワーテレメトリー
- **MAC フォワーディングテーブル**: 検索・フィルタリング・ベンダー解決対応
- **EEE / VLAN / LAG / MTU / ミラー / 帯域制御**: 状態表示
- **設定バックアップ & リストア**: Web UI から設定のダウンロード・アップロード
- **ファームウェア更新**: Web インターフェースからファームウェアをアップロード
- **リモート再起動 & コマンド実行**: 確認付き再起動、および CLI コマンド実行
- **ネットワークトポロジーマップ**: ポート単位の接続クライアント表示、ホスト名上書き対応、ドラッグ＆ドロップレイアウト
- **ログビューア**: ブラウザ上でサーバーログ表示、レベル制御、ダウンロード
- **設定エディタ**: Web ベースのスイッチ・ダッシュボード設定編集
- **ダークガラスモーフィック UI**: カスタムタイポグラフィ、すりガラス風コンポーネント

## 対応ハードウェア

RTLPlayground ファームウェアは Ampcom、Davuaz、FOXNEO、Hisource、Horaco、KeepLink、LIANGUO、Mokerlink、Sodola、Steamemo、TrendNet、XikeStor など 20 以上のデバイスモデルで動作します。詳細は [RTLPlayground supported devices](https://github.com/logicog/RTLPlayground/blob/main/doc/supported_devices.md) をご覧ください。

## クイックスタート

```bash
# フロントエンドをビルド
cd frontend && bun install && bun run build && cd ..

# バイナリをビルド
go build -o switch-dashboard ./cmd/switch-dashboard/

# config.json を用意
cat > config.json << 'EOF'
{
  "title": "My Dashboard",
  "refresh_interval": 30,
  "switches": [
    {
      "name": "Core Switch",
      "ip": "192.168.10.247",
      "password": "1234",
      "model": "RTLPlayground",
      "port_count": 8,
      "enabled": true
    }
  ]
}
EOF

# 実行
./switch-dashboard
```

http://localhost:8081 を開く

## Docker

```bash
docker build -t switch-dashboard .
docker run -d --name switch-dashboard -p 8081:8081 \
  -v $(pwd)/config.json:/config.json \
  -v $(pwd)/history.duckdb:/history.duckdb \
  switch-dashboard
```

## API

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/switches` | 全スイッチのライブ状態（ポート、MAC、SFP、EEE、VLAN、LAG） |
| GET | `/api/switches/:ip/sfp` | SFP+ DDMI 診断情報 |
| GET | `/api/switches/:ip/transceiver` | SFP EEPROM 情報 |
| POST | `/api/switches/:ip/refresh_mac` | MAC フォワーディングテーブルを更新 |
| POST | `/api/switches/:ip/backup` | 設定バックアップをダウンロード |
| POST | `/api/switches/:ip/reboot` | スイッチを再起動 |
| POST | `/api/switches/:ip/upload` | ファームウェアをアップロード |
| POST | `/api/switches/:ip/cmd` | CLI コマンドを実行 |
| GET | `/api/speeds` | ポートごとのリアルタイム帯域（bps） |
| GET | `/api/history?ip=...&port=...&range=live\|1h\|24h` | 帯域履歴 |
| POST | `/api/notes` | ポート注釈を保存 |
| POST | `/api/reset` | 累積カウンターをリセット |
| GET | `/api/topology` | MAC フォワーディングテーブルからネットワークグラフを生成 |
| GET/POST | `/api/settings` | UI 設定 |
| GET/POST | `/api/config/settings` | ダッシュボード設定 |
| GET/POST | `/api/vendors` | MAC ベンダーカスタムマッピング |
| POST | `/api/vendors/update_oui` | IEEE OUI データベースをダウンロード |
| POST | `/api/clients/update_host` | トポロジー上のクライアントホスト名を上書き |
| GET/POST | `/api/layout_positions` | トポロジーノードのレイアウト位置 |
| GET | `/api/logs` | サーバーログ行を取得 |
| POST | `/api/logs/level` | ログレベルを変更 |
| POST | `/api/logs/clear` | サーバーログをクリア |
| GET | `/api/logs/download` | サーバーログをダウンロード |
| GET | `/api/backups` | 設定バックアップ一覧 |
| GET | `/api/backups/:filename/download` | バックアップファイルをダウンロード |
| DELETE | `/api/backups/:filename` | バックアップファイルを削除 |
| GET | `/api/openapi.json` | OpenAPI 3.0 仕様 |

## アーキテクチャ

```
User Browser (TypeScript + Chart.js)
        ↕  HTTP JSON API
Go Server (chi router, single binary, DuckDB history)
        ↕  HTTP JSON API
RTLPlayground Switch (uIP embedded webserver)
```

- **バックエンド**: Go 1.26+、chi router、DuckDB による帯域履歴保存
- **フロントエンド**: TypeScript、Chart.js（CDN）、bun ビルド、インライン CSS ダークテーマ
- **スクレイパー**: RTLPlayground の JSON エンドポイントに直接 HTTP 呼び出し（HTML 解析なし）
- **履歴**: DuckDB 時系列ストア、ライブ / 1h / 24h の集約に対応

## プロジェクト構成

```
├── cmd/switch-dashboard/      # エントリーポイント
├── frontend/
│   ├── src/                   # TypeScript ソース（dashboard, logs, backups, map）
│   └── package.json           # bun ビルド設定
├── internal/
│   ├── server/                # HTTP ハンドラー、キャッシュ、テンプレート、OpenAPI 仕様
│   ├── rtlplayground/         # スイッチ HTTP クライアント + JSON 型
│   ├── poller/                # バックグラウンドポーリング、カウンター、履歴取込
│   ├── config/                # config.json 管理
│   ├── history/               # DuckDB による帯域履歴保存
│   └── store/                 # 汎用 JSON 永続化インターフェース
├── templates/                 # Go html/templates（6 ページ）
├── static/
│   ├── style.css              # ダークガラスモーフィックテーマ
│   ├── dist/                  # コンパイル済みフロントエンド
│   └── logo.png
├── config.json                # スイッチ設定
├── mac_vendors.txt            # MAC OUI データベース
├── Dockerfile                 # マルチステージビルド（bun → Go）
└── history.duckdb             # 時系列データベース（自動生成）
```

## ライセンス

MIT
