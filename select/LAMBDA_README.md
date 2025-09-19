# Aurora DSQL Lambda with API Gateway

このプロジェクトはAurora DSQLに接続して`button_clicks`テーブルからデータを取得するLambda関数と、API Gatewayを通じてアクセスできるようにするSAMアプリケーションです。

## プロジェクト構造

```
.
├── lambda/
│   ├── main.go          # Lambda関数のソースコード
│   ├── go.mod           # Go依存関係
│   └── bootstrap        # ビルド後のバイナリ（自動生成）
├── template.yaml        # SAMテンプレート
├── samconfig.toml       # SAM設定ファイル
├── Makefile            # ビルド・デプロイ自動化
├── test.sh             # curlテストスクリプト
└── events/             # テストイベント（自動生成）
```

## 前提条件

- AWS CLI v2がインストール・設定済み
- SAM CLI v1.100.0以上
- Go 1.23以上
- Docker（SAMのコンテナビルド用）
- jq（テストスクリプトでJSON整形用、オプション）

## セットアップ

### 1. 依存関係のインストール

```bash
make install-deps
```

### 2. ビルド

```bash
# Lambda関数をビルド
make build

# SAMでビルド（コンテナ使用）
make sam-build
```

## ローカルテスト

### 前提条件
- Docker がインストールされ、起動していること
- AWS認証情報が設定されていること（`aws configure`またはIAMロール）
- Aurora DSQLクラスタへのネットワークアクセスがあること

### 利用可能なテストスクリプト

| スクリプト | 目的 | 使用方法 |
|-----------|------|----------|
| `./test-local.sh` | ワンライナー統合テスト | 自動ビルド→実行→結果表示 |
| `./quick-test.sh` | 簡易テスト | Lambda関数の動作確認 |
| `./test.sh` | インタラクティブテスト | メニュー形式の総合テスト |

### テスト方法一覧

#### 方法1: Lambda関数を直接実行（推奨）

最も簡単で高速なテスト方法です。

```bash
# ワンライナーテスト（自動ビルド・実行・結果表示）
./test-local.sh

# 手動でテスト
sam build --use-container
sam local invoke DSQLVersionFunction --event events/test-event.json --region ap-northeast-1

# レスポンスを整形して表示
sam local invoke DSQLVersionFunction --event events/test-event.json --region ap-northeast-1 2>/dev/null | jq '.body | fromjson'

# 簡易テストスクリプトを使用
./quick-test.sh
```

**`./test-local.sh`の実行例:**
```
🚀 Aurora DSQL Lambda Local Test
=================================
✅ Prerequisites check passed
🔨 Building Lambda function with SAM...
✅ Build successful
🧪 Testing Lambda function...
✅ Lambda function executed successfully
📈 Summary: Retrieved 3 button click records
✨ Test completed successfully!
```

**期待される出力:**
```json
{
  "button_clicks": [
    {
      "action": "record",
      "created_at": "2025-09-18T22:56:03Z",
      "id": 1,
      "ip_address": "192.168.1.1",
      "timestamp": "2025-09-18T22:56:03Z",
      "user_agent": "Mozilla/5.0 (Test Browser)"
    }
  ],
  "count": 3,
  "message": "Successfully retrieved button clicks",
  "timestamp": "2025-09-18T23:00:00Z"
}
```

#### 方法2: ローカルAPI Gateway環境

実際のAPIエンドポイントと同様の環境でテストできます。

```bash
# ターミナル1: ローカルAPIを起動
sam local start-api --region ap-northeast-1 --port 3000

# ターミナル2: APIをテスト
curl -X GET http://localhost:3000/version | jq '.'

# または Makefileを使用
make test-local
```

**APIエンドポイント:**
- URL: `http://localhost:3000/version`
- Method: `GET`
- Response: JSON形式のbutton_clicksデータ

#### 方法3: インタラクティブテストスクリプト

メニュー形式で複数のテストオプションを選択できます。

```bash
./test.sh
```

**メニューオプション:**
1. ローカルAPIテスト
2. デプロイ済みAPIテスト
3. パフォーマンステスト（ローカル）
4. パフォーマンステスト（デプロイ済み）
5. エラーハンドリングテスト
6. 全テスト実行

#### 方法4: Makefileコマンド

```bash
# Lambda関数を直接実行
make invoke-local

# ローカルAPIを起動
make local-api

# ローカルAPIをテスト（別ターミナルで）
make test-local

# ビルド後にローカルテスト
make build && make invoke-local

# SAMビルド後にローカルテスト
make sam-build && sam local invoke DSQLVersionFunction --event events/test-event.json --region ap-northeast-1
```

### テスト設定のカスタマイズ

#### テストイベントの編集

`events/test-event.json`を編集してリクエストパラメータを変更できます：

```json
{
  "httpMethod": "GET",
  "path": "/version",
  "headers": {
    "Accept": "application/json",
    "User-Agent": "Test Client"
  },
  "body": null
}
```

#### 環境変数の設定

`template.yaml`の環境変数セクションで設定を変更できます：

```yaml
Environment:
  Variables:
    DSQL_ENDPOINT: your-dsql-endpoint
    DSQL_REGION: ap-northeast-1
    DSQL_DATABASE: postgres
    DSQL_USER: admin
```

### トラブルシューティング

#### エラー: "Docker daemon is not running"
```bash
# Dockerを起動
sudo systemctl start docker
# または Docker Desktop を起動
```

#### エラー: "Auth token generation failed"
```bash
# AWS認証情報を確認
aws sts get-caller-identity

# AWS認証情報を設定
aws configure
```

#### エラー: "Failed to connect to database"
```bash
# ネットワーク接続を確認
ping guabumyfv3jxv2ymjmqtbjqmjq.dsql.ap-northeast-1.on.aws

# IAM権限を確認
aws iam get-user
```

#### エラー: "Template file not found"
```bash
# 正しいディレクトリにいることを確認
pwd
ls template.yaml

# 必要に応じてディレクトリを移動
cd /path/to/project/root
```

#### パフォーマンス最適化

```bash
# キャッシュを使用した高速ビルド
sam build --cached --use-container

# 並列ビルド
sam build --parallel --use-container

# ビルドログの確認
sam build --use-container --debug
```

### ログとデバッグ

#### 詳細ログの表示
```bash
# Lambda関数のログを表示
sam local invoke DSQLVersionFunction --event events/test-event.json --region ap-northeast-1 --debug

# APIゲートウェイのログを表示
sam local start-api --region ap-northeast-1 --port 3000 --debug
```

#### ログの保存
```bash
# ログをファイルに保存
sam local invoke DSQLVersionFunction --event events/test-event.json --region ap-northeast-1 > test-output.log 2>&1

# ログを監視
tail -f test-output.log
```

## AWSへのデプロイ

### 1. デプロイ

```bash
make deploy
```

初回デプロイ時はS3バケット名の入力を求められる場合があります。

### 2. デプロイされたAPIをテスト

```bash
# Makefileを使用
make test-deployed

# テストスクリプトを使用
./test.sh
# メニューから「2」を選択
```

### 3. スタック情報の確認

```bash
make show-outputs
```

## 使用方法

### API エンドポイント

デプロイ後、以下のようなエンドポイントが作成されます：

```
https://{api-id}.execute-api.ap-northeast-1.amazonaws.com/prod/version
```

### curl でのテスト

```bash
# ローカル
curl -X GET http://localhost:3000/version

# デプロイ済み（実際のURLに置き換え）
curl -X GET https://xxxxx.execute-api.ap-northeast-1.amazonaws.com/prod/version
```

### レスポンス例

```json
{
  "button_clicks": [
    {
      "id": 1,
      "action": "record",
      "timestamp": "2025-09-18T22:56:03Z",
      "created_at": "2025-09-18T22:56:03Z",
      "ip_address": "192.168.1.1",
      "user_agent": "Mozilla/5.0 (Test Browser)"
    },
    {
      "id": 2,
      "action": "record",
      "timestamp": "2025-09-18T22:56:03Z",
      "created_at": "2025-09-18T22:56:03Z",
      "ip_address": "192.168.1.2",
      "user_agent": "Mozilla/5.0 (Another Browser)"
    }
  ],
  "count": 2,
  "message": "Successfully retrieved button clicks",
  "timestamp": "2025-01-19T10:00:00Z"
}
```

### エラーレスポンス例

```json
{
  "error": "Database Connection Error",
  "message": "Failed to connect to database: ...",
  "timestamp": "2025-01-19T10:00:00Z"
}
```

## Makefile コマンド一覧

| コマンド | 説明 |
|---------|------|
| `make build` | Lambda関数をビルド |
| `make install-deps` | Go依存関係をインストール |
| `make clean` | ビルド成果物をクリーン |
| `make sam-build` | SAMでコンテナビルド |
| `make deploy` | AWSへデプロイ |
| `make local-api` | ローカルAPIを起動 |
| `make test-local` | ローカルAPIをテスト |
| `make test-deployed` | デプロイ済みAPIをテスト |
| `make invoke-local` | Lambda関数をローカル実行 |
| `make show-outputs` | スタック出力を表示 |
| `make delete` | スタックを削除 |
| `make help` | ヘルプを表示 |

## トラブルシューティング

### ローカルテストでエラーが出る場合

1. Dockerが起動していることを確認
2. AWS認証情報が設定されていることを確認
3. Aurora DSQLへのネットワークアクセスを確認

### デプロイエラーの場合

1. IAMロールの権限を確認
2. リージョンが正しいか確認（ap-northeast-1）
3. SAM CLIが最新バージョンか確認

### 接続エラーの場合

1. Aurora DSQLクラスタが起動していることを確認
2. IAM認証権限があることを確認
3. セキュリティグループの設定を確認

## 環境変数

Lambda関数は以下の環境変数を使用します（template.yamlで設定済み）：

- `DSQL_ENDPOINT`: Aurora DSQLエンドポイント
- `DSQL_REGION`: AWSリージョン
- `DSQL_DATABASE`: データベース名
- `DSQL_USER`: ユーザー名

## クリーンアップ

```bash
# スタックを削除
make delete

# ローカルのビルド成果物を削除
make clean
```

## セキュリティ注意事項

- Lambda関数のIAMロールには最小限の権限のみ付与
- API Gatewayは本番環境ではAPI Keyや認証を追加することを推奨
- Aurora DSQLへの接続は常にSSL/TLS必須

## パフォーマンス最適化

- Lambda関数は接続プールを保持して再利用
- コールドスタート時でも30秒以内に応答
- プール設定は控えめ（最大5接続）に設定済み