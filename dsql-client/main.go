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
	clusterID = "guabumyfv3jxv2ymjmqtbjqmjq"
	region    = "ap-northeast-1"
	database  = "postgres"  // DSQLのデフォルトデータベース名
	username  = "admin"
)

func generateAuthToken(hostname string) (string, error) {
	// AWS CLIを使用してAdmin認証トークンを生成
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

func connectToDSQL() (*sql.DB, error) {
	hostname := fmt.Sprintf("%s.dsql.%s.on.aws", clusterID, region)

	// 認証トークンを生成
	fmt.Println("🔑 認証トークンを生成中...")
	token, err := generateAuthToken(hostname)
	if err != nil {
		return nil, err
	}

	fmt.Printf("✅ 認証トークン生成成功 (長さ: %d文字)\n", len(token))

	// PostgreSQL接続文字列を構築
	// DSQLはPostgreSQLプロトコルを使用
	connStr := fmt.Sprintf("host=%s port=5432 dbname=%s user=%s password=%s sslmode=require",
		hostname, database, username, token)

	// データベース接続
	fmt.Println("🔌 DSQLデータベースに接続中...")
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %v", err)
	}

	// 接続テスト
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	fmt.Println("✅ DSQL接続成功!")
	return db, nil
}

func createTable(db *sql.DB) error {
	fmt.Println("📋 テーブルを作成中...")

	createTableSQL := `
	CREATE TABLE IF NOT EXISTS button_clicks (
		id BIGINT PRIMARY KEY,
		timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		action VARCHAR(100),
		user_agent TEXT,
		ip_address VARCHAR(45),
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);`

	_, err := db.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create table: %v", err)
	}

	fmt.Println("✅ テーブル作成成功!")
	return nil
}

func insertSampleData(db *sql.DB) error {
	fmt.Println("💾 サンプルデータを確認・挿入中...")

	// 既存データの確認
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM button_clicks").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to count existing data: %v", err)
	}

	if count > 0 {
		fmt.Printf("✅ 既存データが%d件見つかりました。新しいデータは挿入しません。\n", count)
		return nil
	}

	// データが存在しない場合のみ挿入
	insertSQL := `
	INSERT INTO button_clicks (id, action, user_agent, ip_address) VALUES
	($1, $2, $3, $4),
	($5, $6, $7, $8),
	($9, $10, $11, $12);`

	_, err = db.Exec(insertSQL,
		1, "record", "Mozilla/5.0 (Test Browser)", "192.168.1.1",
		2, "record", "Mozilla/5.0 (Another Browser)", "192.168.1.2",
		3, "record", "Go DSQL Client", "127.0.0.1")

	if err != nil {
		return fmt.Errorf("failed to insert sample data: %v", err)
	}

	fmt.Println("✅ サンプルデータ挿入成功!")
	return nil
}

func executeQuery(db *sql.DB, query string) error {
	fmt.Printf("📊 SQL実行中: %s\n", query)

	// SQLがSELECTかどうかを判定
	queryLower := strings.ToLower(strings.TrimSpace(query))
	isSelect := strings.HasPrefix(queryLower, "select")

	if isSelect {
		return executeSelectQuery(db, query)
	} else {
		return executeNonSelectQuery(db, query)
	}
}

func executeSelectQuery(db *sql.DB, query string) error {
	rows, err := db.Query(query)
	if err != nil {
		return fmt.Errorf("failed to execute query: %v", err)
	}
	defer rows.Close()

	// カラム名を取得
	columns, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("failed to get columns: %v", err)
	}

	fmt.Printf("\n=== クエリ結果 ===\n")

	// ヘッダーを表示
	for i, col := range columns {
		if i > 0 {
			fmt.Print(" | ")
		}
		fmt.Printf("%-20s", col)
	}
	fmt.Println()
	fmt.Println(strings.Repeat("-", len(columns)*23))

	// データを表示
	for rows.Next() {
		// 動的にスキャン用のスライスを作成
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		err := rows.Scan(valuePtrs...)
		if err != nil {
			return fmt.Errorf("failed to scan row: %v", err)
		}

		// 値を表示
		for i, val := range values {
			if i > 0 {
				fmt.Print(" | ")
			}

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
		return fmt.Errorf("error iterating rows: %v", err)
	}

	fmt.Println("\n✅ クエリ実行完了!")
	return nil
}

func executeNonSelectQuery(db *sql.DB, query string) error {
	result, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to execute query: %v", err)
	}

	// 影響を受けた行数を取得
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		fmt.Println("✅ クエリ実行成功!")
	} else {
		fmt.Printf("✅ クエリ実行成功! (影響を受けた行数: %d)\n", rowsAffected)
	}

	return nil
}

func selectAllData(db *sql.DB) error {
	query := "SELECT id, timestamp, action, user_agent, ip_address, created_at FROM button_clicks ORDER BY created_at DESC;"
	return executeQuery(db, query)
}

func main() {
	fmt.Println("🚀 DSQL Go クライアント開始")
	fmt.Printf("📍 クラスターID: %s\n", clusterID)
	fmt.Printf("🌍 リージョン: %s\n", region)
	fmt.Printf("💾 データベース: %s\n", database)
	fmt.Println()

	// DSQLに接続
	db, err := connectToDSQL()
	if err != nil {
		log.Fatalf("❌ DSQL接続エラー: %v", err)
	}
	defer db.Close()

	// テーブル作成
	if err := createTable(db); err != nil {
		log.Fatalf("❌ テーブル作成エラー: %v", err)
	}

	// サンプルデータ挿入
	if err := insertSampleData(db); err != nil {
		log.Fatalf("❌ データ挿入エラー: %v", err)
	}

	// 全データを取得
	if err := selectAllData(db); err != nil {
		log.Fatalf("❌ データ取得エラー: %v", err)
	}

	fmt.Println("\n🎉 全ての操作が完了しました!")
}
