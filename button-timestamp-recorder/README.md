# ボタンクリックタイムスタンプ記録システム

API Gateway、Lambda、Aurora Serverless v2 (DSQL)を使用してボタンクリックの日時を記録するシステムです。

## 構成

- **Lambda関数**: Go言語で実装
- **API Gateway**: REST API
- **データベース**: Aurora Serverless v2 Data API (DSQL)
- **フロントエンド**: シンプルなHTML/JavaScript
- **デプロイ**: AWS SAM

## 必要な準備

SAMがインストールされていることを確認してください：

```bash
# SAM CLIのインストール（未インストールの場合）
pip install aws-sam-cli

# バージョン確認
sam --version
```

## デプロイ手順

### 1. 依存関係のダウンロード

```bash
cd button-timestamp-recorder/functions/record-timestamp
go mod download
cd ../../..
```

### 2. SAMアプリケーションのビルド

```bash
cd button-timestamp-recorder
sam build
```

### 3. デプロイ（初回）

```bash
sam deploy --guided
```

以下のパラメータを入力してください：

- **Stack Name**: button-timestamp-recorder
- **AWS Region**: ap-northeast-1（東京リージョン）
- **DatabaseName**: button_db（デフォルトのまま）
- **MasterUsername**: admin（デフォルトのまま）
- **MasterPassword**: ButtonTimestamp2024!（デフォルトのまま、または変更）

**注意**: パスワードはtemplate.yamlにデフォルト値が設定されていますが、本番環境では必ず変更してください。

### 4. フロントエンドのデプロイ

```bash
# S3バケット名を取得
BUCKET_NAME=$(aws cloudformation describe-stacks \
  --stack-name button-timestamp-recorder \
  --query "Stacks[0].Outputs[?OutputKey=='FrontendBucket'].OutputValue" \
  --output text)

# フロントエンドファイルをアップロード
aws s3 cp frontend/index.html s3://${BUCKET_NAME}/
```

## 使用方法

1. デプロイ完了後、CloudFormationスタックの出力から以下の情報を取得：
   - **ApiEndpoint**: API GatewayのURL
   - **WebsiteURL**: フロントエンドのURL

2. ブラウザでフロントエンドURLにアクセス

3. APIエンドポイントURLを設定欄に入力

4. ボタンをクリックすると、クリック日時がデータベースに記録されます

## データベーステーブル構造

```sql
CREATE TABLE button_clicks (
    id INT AUTO_INCREMENT PRIMARY KEY,
    clicked_at TIMESTAMP NOT NULL,
    action VARCHAR(255)
);
```

## ローカル開発

sam local invoke RecordTimestampFunction -e events/record.json --env-vars env.local.json | jq .

### Lambda関数のローカルテスト

```bash
make local
```

### ログの確認

```bash
make logs
```

## クリーンアップ

```bash
# スタックの削除（VPC、データベース、Lambda、S3バケットすべて削除されます）
make delete
```

**注意**: スタック削除時にAurora Serverless v2クラスターも一緒に削除されます。

## トラブルシューティング

### CORSエラーが発生する場合

フロントエンドとAPIが異なるドメインにある場合、Lambda関数のCORSヘッダー設定を確認してください。

### データベース接続エラー

1. Aurora Serverless v2のData APIが有効になっているか確認
2. Secrets ManagerのシークレットARNが正しいか確認
3. Lambda関数のIAMロールに必要な権限があるか確認

### ビルドエラー

Go 1.21以上がインストールされているか確認してください：

```bash
go version
```
