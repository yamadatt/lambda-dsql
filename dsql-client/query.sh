#!/bin/bash

# DSQL SELECT クエリ実行スクリプト
# 使用方法: ./query.sh "SELECT * FROM button_clicks;"

if [ $# -eq 0 ]; then
    echo "使用方法: $0 \"SQL文\""
    echo "例: $0 \"SELECT * FROM button_clicks;\""
    echo "例: $0 \"SELECT COUNT(*) FROM button_clicks;\""
    echo "例: $0 \"SELECT action, COUNT(*) FROM button_clicks GROUP BY action;\""
    exit 1
fi

SQL_QUERY="$1"

echo "🔍 SQL実行中: $SQL_QUERY"
echo

# Goプログラムを使ってクエリを実行
cd "$(dirname "$0")"
go run -ldflags="-s -w" - <<EOF
package main

import (
	"database/sql"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"
	_ "github.com/lib/pq"
)

const (
	clusterID = "baabumxlegra2drhyb7t77y4cq"
	region    = "ap-northeast-1"
	database  = "postgres"
	username  = "admin"
)

func generateAuthToken(hostname string) (string, error) {
	cmd := exec.Command("aws", "dsql", "generate-db-connect-admin-auth-token",
		"--hostname", hostname,
		"--region", region,
		"--expires-in", "3600",
		"--output", "text")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to generate auth token: %v", err)
	}
	return strings.TrimSpace(string(output)), nil
}

func main() {
	hostname := fmt.Sprintf("%s.dsql.%s.on.aws", clusterID, region)
	token, err := generateAuthToken(hostname)
	if err != nil {
		log.Fatalf("❌ 認証エラー: %v", err)
	}

	connStr := fmt.Sprintf("host=%s port=5432 dbname=%s user=%s password=%s sslmode=require",
		hostname, database, username, token)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("❌ 接続エラー: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("❌ 接続テストエラー: %v", err)
	}

	query := "$SQL_QUERY"
	rows, err := db.Query(query)
	if err != nil {
		log.Fatalf("❌ クエリエラー: %v", err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		log.Fatalf("❌ カラム取得エラー: %v", err)
	}

	for i, col := range columns {
		if i > 0 { fmt.Print(" | ") }
		fmt.Printf("%-20s", col)
	}
	fmt.Println()
	fmt.Println(strings.Repeat("-", len(columns)*23))

	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		err := rows.Scan(valuePtrs...)
		if err != nil {
			log.Fatalf("❌ スキャンエラー: %v", err)
		}

		for i, val := range values {
			if i > 0 { fmt.Print(" | ") }
			var displayValue string
			if val == nil {
				displayValue = "NULL"
			} else {
				switch v := val.(type) {
				case time.Time:
					displayValue = v.Format("2006-01-02 15:04:05")
				case []byte:
					displayValue = string(v)
				default:
					displayValue = fmt.Sprintf("%v", v)
				}
			}
			fmt.Printf("%-20s", displayValue)
		}
		fmt.Println()
	}

	if err = rows.Err(); err != nil {
		log.Fatalf("❌ 行処理エラー: %v", err)
	}
}
EOF