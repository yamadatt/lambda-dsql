#!/bin/bash

# デプロイスクリプト
# Aurora Serverless v2クラスターとすべての必要なリソースを自動作成します

set -e

echo "🚀 ボタンタイムスタンプ記録システムのデプロイを開始します"

# Go依存関係のダウンロード
echo "📦 Go依存関係をダウンロード中..."
cd functions/record-timestamp
go mod download
cd ../..

# SAMアプリケーションのビルド
echo "🔨 SAMアプリケーションをビルド中..."
sam build

# デプロイパラメータの確認
echo ""
echo "📝 このデプロイでは以下のリソースが自動作成されます："
echo "  - VPC とサブネット"
echo "  - Aurora Serverless v2 クラスター"
echo "  - Secrets Manager シークレット"
echo "  - Lambda 関数"
echo "  - API Gateway"
echo "  - S3 バケット（フロントエンド用）"
echo ""
echo "デフォルトパスワード: ButtonTimestamp2024!"
echo "（本番環境では変更することを推奨します）"
echo ""

# デプロイ実行
echo "🚢 デプロイを実行中..."
sam deploy --guided

# デプロイ結果の取得
echo ""
echo "✅ デプロイが完了しました！"
echo ""

# 出力値の取得と表示
API_ENDPOINT=$(aws cloudformation describe-stacks \
    --stack-name button-timestamp-recorder \
    --query "Stacks[0].Outputs[?OutputKey=='ApiEndpoint'].OutputValue" \
    --output text 2>/dev/null)

WEBSITE_URL=$(aws cloudformation describe-stacks \
    --stack-name button-timestamp-recorder \
    --query "Stacks[0].Outputs[?OutputKey=='WebsiteURL'].OutputValue" \
    --output text 2>/dev/null)

BUCKET_NAME=$(aws cloudformation describe-stacks \
    --stack-name button-timestamp-recorder \
    --query "Stacks[0].Outputs[?OutputKey=='FrontendBucket'].OutputValue" \
    --output text 2>/dev/null)

# フロントエンドのアップロード
if [ ! -z "$BUCKET_NAME" ]; then
    echo "📤 フロントエンドをS3にアップロード中..."
    aws s3 cp frontend/index.html s3://${BUCKET_NAME}/ --content-type "text/html"
    echo "✅ フロントエンドのアップロードが完了しました"
fi

echo ""
echo "========================================="
echo "📋 デプロイ情報"
echo "========================================="
if [ ! -z "$API_ENDPOINT" ]; then
    echo "API エンドポイント: $API_ENDPOINT"
fi
if [ ! -z "$WEBSITE_URL" ]; then
    echo "Webサイト URL: $WEBSITE_URL"
fi
echo ""
echo "🎉 すべての設定が完了しました！"
echo "Webサイトにアクセスして、APIエンドポイントURLを設定してください。"