# Aurora DSQL接続アプリケーション

Aurora DSQLに接続して`SELECT version();`を実行するGoアプリケーションです。

## 機能

- AWS IAM認証を使用してAurora DSQLに接続
- 接続プール機能で効率的な接続管理
- 自動トークン更新による継続的な接続維持
- `SELECT version()`クエリを実行してPostgreSQLバージョンを表示

## 前提条件

- Go 1.23以上
- AWS認証情報が設定されていること（環境変数またはIAMロール）
- Aurora DSQLクラスタへのアクセス権限

## 依存パッケージ

- `github.com/aws/aws-sdk-go-v2/config` - AWS SDK設定
- `github.com/aws/aws-sdk-go-v2/feature/dsql/auth` - DSQL認証機能
- `github.com/jackc/pgx/v5` - PostgreSQLドライバー

## ビルド方法

```bash
# 依存関係のインストール
go mod tidy

# ビルド
go build -o dsql-select main.go
```

## 実行方法

```bash
./dsql-select
```

## 接続情報

- **ホスト名**: guabumyfv3jxv2ymjmqtbjqmjq.dsql.ap-northeast-1.on.aws
- **データベース**: postgres
- **ユーザー名**: admin
- **リージョン**: ap-northeast-1
- **ポート**: 5432

## 実行結果の例

```
Connecting to Aurora DSQL at guabumyfv3jxv2ymjmqtbjqmjq.dsql.ap-northeast-1.on.aws
Auth token generated (length: 347)
Successfully connected to Aurora DSQL!

Database version:
PostgreSQL 16

Version summary: PostgreSQL 16
```

## 実装の特徴

- **pgxpool**を使用した接続プール管理
- **BeforeConnect**フックでトークン生成を自動化
- SSL/TLS必須接続（`sslmode=verify-full`、`sslnegotiation=direct`）
- adminユーザー専用の認証トークン生成メソッド使用

## 注意事項

- AWS認証情報が正しく設定されている必要があります
- Aurora DSQLクラスタへのネットワークアクセスが必要です
- SSL/TLS接続が必須です
- トークンの有効期限は30秒に設定されています