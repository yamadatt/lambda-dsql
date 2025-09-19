#!/bin/bash

# DSQL Go Client 実行スクリプト

echo "🚀 DSQL Go Client を実行します..."
echo "📍 作業ディレクトリ: $(pwd)"
echo

# 依存関係の確認とダウンロード
if [ ! -f "go.sum" ] || [ "go.mod" -nt "go.sum" ]; then
    echo "📦 依存関係をダウンロード中..."
    go mod tidy
    echo
fi

# プログラムの実行
echo "▶️ プログラムを実行中..."
go run main.go

echo
echo "✅ 実行完了"