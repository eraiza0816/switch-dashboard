# Switch Dashboard Go移植 検討資料

> **スコープ**: RTLPlayground ファームウェアが動作するスイッチのみサポート対象とする。
> RTLPlayground は 20 機種以上のデバイス（Horaco, KeepLink, Lianguo, TrendNet 等）に対応しており、
> すべて同一の JSON API を提供する。機種ごとの差分は LED 設定・ポート数・SFP スロット数のみ。

## 1. プロジェクト規模（移植対象のみ）

| 区分 | ファイル | 行数 | 備考 |
|---|---|---|---|
| Python バックエンド | app.py | 3,073 | Flask コア、REST API、スレッド管理 |
| | scraper.py | 2,544 | 全スクレイパー + YAMLテンプレートエンジン |
| | **小計** | **5,617** | |
| HTML テンプレート | index.html | 2,330 | メインダッシュボード（JS/CSS インライン） |
| | config.html | 1,800 | 設定ページ |
| | api_docs.html | 1,831 | API ドキュメント（静的） |
| | logs.html | 753 | ログビューアー |
| | backups.html | 662 | バックアップ管理 |
| | **小計** | **7,376** | |
| 合計 | | **~13,000** | 参考するコード量（Go での実装行数は 1,500-2,000行程度）|

## 2. RTLPlayground スイッチの API

全 22 エンドポイント。認証は `POST /login` → `Set-Cookie: session=<hex>`。全 JSON API は認証必須。

### JSON API（全 14 エンドポイント）

| エンドポイント | メソッド | パラメーター | 応答 | 用途 |
|---|---|---|---|---|
| `/information.json` | GET | なし | JSON | デバイス情報（MAC, IP, FW, HW, ホスト名, Telnet/Web有効, SFP vendor） |
| `/status.json` | GET | なし | JSON 配列 | 全ポートの状態・リンク速度・アドバタイズ・SFP 情報・TX/RX パケットカウンター |
| `/counters.json?port=N` | GET | port | JSON 配列 | 55 個の MIB カウンター値（Hex 文字列） |
| `/l2.json?idx=N` | GET | idx | JSON 配列 | MAC テーブル（最大 30 エントリー/ページ） |
| `/l2_del.json?idx=N` | GET | idx | JSON | L2 エントリー削除 |
| `/vlan.json?vid=N` | GET | vid | JSON | VLAN メンバーシップ・PVID・VLAN 名 |
| `/vlanlist` | GET | なし | JSON 配列 | 全 VLAN 一覧（ID + 名前） |
| `/sfp_diag.json` | GET | なし | JSON 配列 | SFP DDMI 診断（温度, VCC, TX Bias, TX Power, RX Power） |
| `/sfp_eeprom.json?slot=N` | GET | slot | JSON | SFP EEPROM 全 256 バイト（16進文字列） |
| `/eee.json` | GET | なし | JSON 配列 | ポートごとの EEE 設定・リンクパートナー能力・動作状態 |
| `/bandwidth.json` | GET | なし | JSON 配列 | ポートごとの Ingress/Egress 帯域制御設定 |
| `/mirror.json` | GET | なし | JSON | ポートミラーリング設定 |
| `/lag.json` | GET | なし | JSON 配列 | Link Aggregation Group 設定（全 4 グループ） |
| `/mtu.json` | GET | なし | JSON 配列 | ポートごとの MTU 設定 |

### Text API（2 エンドポイント）

| エンドポイント | メソッド | 応答 | 用途 |
|---|---|---|---|
| `/config` | GET | text/plain | 設定ファイルバックアップ（現在の flash 設定をテキストで取得） |
| `/cmd_log` | GET | text/plain | コマンド実行履歴 |

### Action（2 エンドポイント）

| エンドポイント | メソッド | 用途 |
|---|---|---|
| `/cmd_log_clear` | GET | コマンド履歴クリア |
| `/reset` | GET | スイッチ再起動（即座に接続断） |

### POST（4 エンドポイント）

| エンドポイント | メソッド | Content-Type | 用途 |
|---|---|---|---|
| `/login` | POST | x-www-form-urlencoded | `pwd=<password>` でログイン、Cookie 発行 |
| `/cmd` | POST | plain text | CLI コマンド実行（ボディにコマンド文字列） |
| `/upload` | POST | multipart/form-data | ファームウェアアップロード |
| `/config` | POST | multipart/form-data | 設定ファイルアップロード |

**すべて JSON 応答。HTML パースは一切不要。**

**RTLPlayground の API は全機種共通。** 機種によりポート数・SFP スロット数が異なるのみ。

## 3. Go プロジェクト構成

```
switch-dashboard/
├── cmd/
│   └── switch-dashboard/main.go
├── internal/
│   ├── server/                  # HTTP サーバー + ルーティング（chi）
│   │   ├── server.go            # ルーター設定
│   │   ├── middleware.go        # ロギング
│   │   ├── routes.go            # ルート定義
│   │   ├── handler_switches.go  # GET /api/switches（全データ + SFP + IGMP）
│   │   ├── handler_history.go   # GET /api/speeds, GET /api/history
│   │   ├── handler_topology.go  # GET /api/topology（MAC テーブルからグラフ生成）
│   │   ├── handler_sfp.go       # GET /api/switches/<ip>/sfp_diag（DDMI）
│   │   ├── handler_config.go    # /config, GET/POST /api/settings
│   │   ├── handler_backup.go    # POST /api/backup, GET/DELETE /api/backups
│   │   └── handler_logs.go      # GET/POST /api/logs
│   ├── rtlplayground/           # RTLPlayground スイッチ操作用クライアント
│   │   ├── client.go            # HTTP クライアント（認証 + Cookie 管理）
│   │   ├── types.go             # 全 JSON 応答の型定義（status, info, sfp, eee, vlan, lag, mirror等）
│   │   ├── scrape.go            # /information.json + /status.json 統合取得
│   │   ├── status.go            # /status.json 各種パース
│   │   ├── port_counters.go     # /counters.json MIB カウンター
│   │   ├── mac_table.go         # /l2.json ページング対応取得 + /l2_del.json
│   │   ├── sfp.go               # /sfp_diag.json + /sfp_eeprom.json
│   │   ├── vlan.go              # /vlan.json + /vlanlist
│   │   ├── eee.go               # /eee.json
│   │   ├── bandwidth.go         # /bandwidth.json
│   │   ├── mirror.go            # /mirror.json
│   │   ├── lag.go               # /lag.json
│   │   ├── mtu.go               # /mtu.json
│   │   ├── config.go            # GET/POST /config（バックアップ・リストア）
│   │   ├── cmd.go               # POST /cmd（CLI コマンド実行, リブート）
│   │   └── upload.go            # POST /upload（ファームウェアアップデート）
│   ├── store/                   # データ永続化
│   │   ├── store.go             # インターフェース
│   │   └── json.go              # JSON ファイル読み書き
│   ├── poller/                  # バックグラウンドポーリング
│   │   ├── poller.go            # ポーリングループ（goroutine + ticker + context）
│   │   └── history.go           # 3層ヒストリー管理
│   ├── vendor/                  # MAC OUI ベンダー解決
│   │   ├── oui.go               # IEEE OUI ファイルパース
│   │   └── lookup.go            # MAC → ベンダー名 解決
│   └── config/
│       └── config.go            # config.json 読み書き
├── templates/
│   ├── index.html               # Jinja2 → html/template に修正
│   ├── config.html              # CodeMirror/YAML 編集タブ削除、SFP FW update 追加
│   ├── map.html                 # トポロジーマップ（RTLPlayground 版に修正）
│   ├── logs.html
│   ├── backups.html
│   └── api_docs.html
├── static/
├── go.mod / go.sum
└── Makefile
```

## 4. 推奨 Go ライブラリ

| 目的 | ライブラリ | 備考 |
|---|---|---|
| HTTP ルーター | `github.com/go-chi/chi/v5` | 軽量、標準 `net/http` 準拠 |
| HTML テンプレート | `html/template` | 標準ライブラリ |
| ロギング | `log/slog` | Go 1.21+ 標準構造化ロギング |
| テスト | `testing` + `github.com/stretchr/testify` | アサーション補助 |
| HTTP テスト | `net/http/httptest` | 標準ライブラリ |

**外部依存は chi のみ。** 他はすべて標準ライブラリで完結。

## 5. 技術的課題と対応方針

### 5.1 RTLPlayground HTTP クライアント

**現状**: Python `urllib` + CookieJar。`POST /login` でパスワード送信、Cookie 保持。

**Go での対応**: `net/http` + `cookiejar.Jar`。~50行。

```go
type Client struct {
    client *http.Client
    baseURL string
}

func New(ip, password string) *Client {
    jar, _ := cookiejar.New(nil)
    c := &http.Client{Jar: jar}
    c.PostForm(fmt.Sprintf("http://%s/login", ip), url.Values{"pwd": {password}})
    return &Client{client: c, baseURL: fmt.Sprintf("http://%s", ip)}
}
```

### 5.2 JSON 応答の型定義

**現状**: Python の辞書アクセス（動的型付け）。

**Go での対応**: 構造体定義 + `encoding/json`。RTLPlayground のスキーマは固定なので一度書けば終わり。

```go
type StatusJSON struct {
    PortNum  int   `json:"portNum"`
    Enabled  int   `json:"enabled"`
    Link     int   `json:"link"`
    TxG      int64 `json:"txG"`
    RxG      int64 `json:"rxG"`
    TxB      int64 `json:"txB"`
    RxB      int64 `json:"rxB"`
}
```

### 5.3 バックグラウンドポーリング

**現状**: Python `threading.Thread` + `threading.Lock`。

**Go での対応**: goroutine + `time.Ticker` + `context.Context`。

```go
func (p *Poller) Run(ctx context.Context) {
    ticker := time.NewTicker(p.interval)
    defer ticker.Stop()
    for {
        select {
        case <-ticker.C:
            p.poll()
        case <-ctx.Done():
            return
        }
    }
}
```

### 5.4 データ永続化

**現状**: JSON ファイル 4種（config.json, counters.json, history_hourly.json, history_daily.json）。

**Go での対応**: `encoding/json` + `os.ReadFile/WriteFile`。`sync.RWMutex` で排他制御。

### 5.5 累積カウンターとスパイク検出

**現状**: 32bit カウンターラップ検出、リンク速度ベースのスパイクフィルター、スイッチ再起動検出。

**Go での対応**: 純粋な数値演算のため逐語移植可能。TDD で正確性を担保すべき最重要箇所。

## 6. TDD ベースの開発順序

### Phase 1: コアモデルとデータ層（1-2日）
1. モデル構造体定義（Config, SwitchInfo, Port, HistoryPoint, Counters）
2. JSON ストアの Load/Save + 排他制御のテスト
3. 累積カウンターエンジンのテスト（ラップアラウンド、スパイクフィルター）
4. 3層ヒストリー管理のテスト（Live/1h/24h、速度計算）

*テスト容易性: 高（外部依存なし）*

### Phase 2: RTLPlayground クライアント（2-3日）
5. HTTP クライアント + Cookie 管理のラッパー（`net/http` + `cookiejar`）
6. 全 JSON 応答の型定義と Unmarshal テスト（`information.json`, `status.json`, `sfp_diag.json`, `eee.json`, `vlan.json`, `lag.json`, `mtu.json`, `bandwidth.json`, `mirror.json`, `l2.json`, `counters.json`）
7. `ScrapeState()`: `/information.json` + `/status.json` 統合（デバイス情報 + ポート状態 + カウンター）
8. `ScrapeSFP()`: `/sfp_diag.json` で DDMI 診断データ取得
9. `ScrapeMACTable()`: `/l2.json` ページング対応読み取り
10. `ScrapeVLANs()`, `ScrapeEEE()`, `ScrapeLAG()`, `ScrapeMTU()`, `ScrapeBandwidth()`, `ScrapeMirror()`
11. `DownloadBackup()`: `GET /config`
12. `Reboot()`: `POST /cmd` でリブート
13. `UploadFirmware()`: `POST /upload`（multipart）
14. ログイン失敗・応答なし・404 のエラーハンドリングテスト

*テスト容易性: 中（httptest で全エンドポイントのモックサーバーを構築）*

### Phase 3: HTTP サーバー + ハンドラー（3-4日）
12. chi ルーター設定 + ミドルウェア（ロギング・リカバリー）
13. 全 API ハンドラーの実装と httptest テスト
    - `GET /api/switches` → ポーリングデータ（状態 + SFP + VLAN + EEE + LAG + 帯域制御）
    - `GET /api/switches/<ip>/sfp` → SFP DDMI 診断
    - `GET /api/switches/<ip>/transceiver` → SFP EEPROM 情報
    - `GET /api/topology` → MAC テーブルからトポロジーグラフ生成
    - `GET /api/speeds`, `GET /api/history` → 帯域履歴
    - `POST /api/notes` → ポート注釈
    - `POST /api/reset` → カウンターリセット
    - `POST /api/switches/<ip>/backup` → 設定バックアップ
    - `POST /api/switches/<ip>/reboot` → リブート
    - `POST /api/switches/<ip>/upload` → ファームウェアアップデート
    - `GET/POST /api/settings` → UI 設定
    - `GET /api/logs`, `POST /api/logs/level`, `POST /api/logs/clear`
    - `GET /api/vendors`, `POST /api/vendors/update_oui` → MAC ベンダー管理

*テスト容易性: 高（httptest）*

### Phase 4: バックグラウンドポーリング（1日）
14. ポーリングループ（goroutine + ticker + context）
15. キャッシュ更新と排他制御
16. 3層ヒストリー自動記録

### Phase 5: フロントエンド統合（0.5-1日）
17. Jinja2 → html/template 構文修正
18. config.html 簡略化（CodeMirror/YAML 編集タブ削除）
19. api_docs.html のエンドポイント一覧更新

### Phase 6: CI/CD・デプロイ（0.5日）
20. GitHub Actions でテスト + ビルド
21. マルチステージ Dockerfile
22. systemd ユニットファイル / Makefile

**合計: 6-10日（フルタイム換算）**

## 7. 推奨テスト戦略

```
switch-dashboard/
├── internal/
│   ├── store/
│   │   └── json_test.go          # 一時ファイル Load/Save テスト
│   ├── rtlplayground/
│   │   ├── client_test.go        # httptest モックサーバーテスト
│   │   ├── types_test.go         # JSON Unmarshal テスト
│   │   └── mac_table_test.go     # ページングロジックテスト
│   ├── poller/
│   │   ├── poller_test.go        # goroutine 制御テスト
│   │   └── history_test.go       # カウンター演算テスト
│   └── server/
│       └── handler_test.go       # 全 API 統合テスト
├── testdata/
│   ├── information.json
│   ├── status.json
│   ├── sfp_diag.json
│   ├── sfp_eeprom.json
│   ├── l2_page0.json
│   ├── l2_page1.json
│   ├── eee.json
│   ├── vlan.json
│   ├── vlanlist.json
│   ├── lag.json
│   ├── bandwidth.json
│   ├── mirror.json
│   └── mtu.json
└── integration_test.go           # 実機結合テスト（-tags=integration）
```

## 8. Go 移植のメリット・デメリット

### メリット
- **単一バイナリ**: Python 環境不要、`scp` で配置して実行するだけ
- **Docker イメージ ~8MB**: `scratch` ベース。Python の 120MB から大幅削減
- **外部依存ゼロ**: `chi` のみ。CGO 不要
- **型安全性**: RTLPlayground の JSON スキーマが固定 → Go の静的型付けが最大限活きる
- **テスト容易**: `httptest` でスイッチモックを立てて全動作をオフラインテスト可能
- **クロスコンパイル**: `GOARCH=arm64` で Raspberry Pi 用バイナリも即生成

### デメリット
- **Go 習熟コスト**: Python からの移行に学習曲線が必要
- **JSON ファイル永続化**: `encoding/json` は Python の `json.dump` より冗長

## 9. 結論

**RTLPlayground 専用に絞った Go 移植は極めて合理的。**

```
                Python (現状)          Go (移植後)
依存ライブラリ    Flask, requests        chi のみ
CGO              あり (Scapy)           不要
Docker サイズ     ~120MB                ~8MB
バックエンド行数   ~5,600行              ~1,500-2,000行
プロセスモデル    スレッド + Lock         goroutine + channel
```

リスクは以下の点のみ:
1. **カウンター演算の移植正確性** → TDD + フィクスチャテストでカバー
2. **Go 習熟コスト** → プロジェクト規模が小さいため許容範囲

**推奨アプローチ**: Go で完全新規実装。フロントエンドは既存 HTML/JS を流用（Jinja2 → html/template のみ修正）。

---

## 10. 計画・実装不要

以下の機能・ファイル・ライブラリは、RTLPlayground 専用化に伴い移植対象外とする。

### 10.1 移植不要な Python ファイル

| ファイル | 行数 | 理由 |
|---|---|---|
| `network_scanner.py` | 144 | Scapy ARP スキャン + TCP ポートスキャン。サブネットスキャン機能はスコープ外 |
| `scanner_db.py` | 304 | スキャン履歴の SQLite 管理。スコープ外 |
| `migrate_scanner_data.py` | 247 | MariaDB → SQLite のデータ移行。一度きりのスクリプト |

### 10.2 移植不要な HTML テンプレート

| ファイル | 行数 | 理由 |
|---|---|---|
| `templates/scanner.html` | 1,756 | サブネットスキャナー一覧画面。スコープ外 |
| `templates/scanner_history.html` | 1,056 | ホスト接続履歴タイムライン。スコープ外 |

### 10.3 移植不要なデータファイル

| ファイル | 理由 |
|---|---|
| `device-templates/*.yaml` (3ファイル) | RTLPlayground は全機種同一 API のためハードコードで十分 |
| `device_types.yaml` | デバイスタイプアイコン定義（元トポロジーマップ用） |

### 10.4 移植不要な Python ライブラリ

| ライブラリ | 用途 | 理由 |
|---|---|---|
| `BeautifulSoup4` | HTML パース | RTLPlayground はすべて JSON 応答のため不要 |
| `Scapy` | ARP ネットワークスキャン | サブネットスキャンはスコープ外 |
| `Paramiko` | SSH クライアント | OVS スクレイパーはスコープ外 |
| `PyYAML` | YAML パース | テンプレートエンジン不要、ハードコードのため不要 |
| `sqlite3` | データベース | スキャン履歴管理はスコープ外 |

### 10.5 移植不要な機能（app.py / scraper.py 由来）

| 機能 | 該当コード量（目安） | 理由 |
|---|---|---|
| HCSwitchScraper（CGI/HTML 汎用） | scraper.py ~1,400行 | 汎用中国製スイッチ（Horaco/KeepLink OEM等）はスコープ外 |
| OVSScraper（Open vSwitch SSH） | scraper.py ~350行 | Proxmox OVS はスコープ外 |
| FritzBoxScraper（SOAP/UPnP） | scraper.py ~350行 | AVM FRITZ!Box はスコープ外 |
| YAML テンプレートエンジン | scraper.py ~800行 | RTLPlayground は全機種同一 API のため不要。ただし 20+ 機種が存在することに注意 |
| サブネットスキャナー制御 | app.py ~250行 | ARP/TCP ポートスキャンはスコープ外 |
| クライアント DB 管理 | app.py ~400行 | スキャン結果の永続化はスコープ外 |
| DHCP Snooping 取得 | scraper.py + app.py ~100行 | RTLPlayground は DHCP 非対応（README より） |
| テレメトリー送信 | app.py ~80行 | 匿名使用状況報告は不要 |

### 10.6 「やらない」ことによる影響

| やらないこと | 影響 |
|---|---|
| サブネット ARP スキャン | ネットワーク上の未知のホスト自動発見は不可。ただしスイッチの L2 MAC テーブルで接続機器は確認可能 |
| DHCP Snooping | RTLPlayground が DHCP 非対応のため影響なし |
| FritzBox / OVS 対応 | スコープ外。RTLPlayground の対象外 |
| YAML テンプレートエンジン | RTLPlayground は全機種で API が統一されているため不要。機種差分は `internal/rtlplayground/types.go` の定数（ポート数・SFP 数）で吸収可能 |
