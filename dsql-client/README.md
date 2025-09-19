# DSQL Go Client

AWS Aurora DSQLに接続してデータベース操作を行うGoクライアントです。

## 機能

- DSQL クラスターへの接続
- テーブル作成（button_clicks）
- サンプルデータの挿入
- SELECT * でのデータ取得

## 前提条件

- Go 1.21以上
- AWS CLI設定済み
- DSQL クラスターへのアクセス権限

## 実行方法

```bash
# 依存関係のダウンロード
go mod tidy

# プログラムの実行
go run main.go

# バイナリの作成
go build -o dsql-client main.go

# バイナリの実行
./dsql-client
```

## 設定

プログラム内の定数を変更して、異なるクラスターに接続できます：

```go
const (
    clusterID = "baabumxlegra2drhyb7t77y4cq"  // DSQLクラスターID
    region    = "ap-northeast-1"              // AWSリージョン
    database  = "postgres"                    // データベース名
    username  = "admin"                       // ユーザー名
)
```

## 注意事項

- Admin認証トークンを使用しているため、管理者権限が必要です
- DSQLはSERIALやIDENTITY制約をサポートしていません
- 既存データがある場合、重複挿入は行いません

## トラブルシューティング

### 接続エラーの場合
1. AWS CLIの設定を確認
2. DSQLクラスターが稼働中であることを確認
3. IAM権限を確認

### 認証エラーの場合
1. `aws dsql generate-db-connect-admin-auth-token`コマンドを手動実行して確認
2. IAMユーザー/ロールにDSQL権限があることを確認