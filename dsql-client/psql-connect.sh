#!/bin/bash

# DSQL用psql接続コマンド生成スクリプト

CLUSTER_ID="baabumxlegra2drhyb7t77y4cq"
REGION="ap-northeast-1"
DATABASE="postgres"
USERNAME="admin"
HOSTNAME="${CLUSTER_ID}.dsql.${REGION}.on.aws"

echo "🔐 DSQL認証トークンを生成中..."

# 認証トークンを生成
TOKEN=$(aws dsql generate-db-connect-admin-auth-token \
    --hostname "$HOSTNAME" \
    --region "$REGION" \
    --expires-in 3600 \
    --output text)

if [ $? -ne 0 ]; then
    echo "❌ 認証トークン生成に失敗しました"
    exit 1
fi

echo "✅ 認証トークン生成成功"
echo ""

echo "📋 psqlコマンド:"
echo "PGPASSWORD='$TOKEN' psql -h $HOSTNAME -p 5432 -d $DATABASE -U $USERNAME"

echo ""
echo "📋 SELECT文の実行例:"
echo "PGPASSWORD='$TOKEN' psql -h $HOSTNAME -p 5432 -d $DATABASE -U $USERNAME -c \"SELECT * FROM button_clicks;\""

echo ""
echo "📋 対話モードでpsqlに接続する場合:"
echo "export PGPASSWORD='$TOKEN'"
echo "psql -h $HOSTNAME -p 5432 -d $DATABASE -U $USERNAME"

echo ""
echo "⏰ このトークンは1時間有効です"