package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dsql/auth"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// Aurora DSQLクラスタ情報
	hostname := "guabumyfv3jxv2ymjmqtbjqmjq.dsql.ap-northeast-1.on.aws"
	database := "postgres"
	username := "admin"
	region := "ap-northeast-1"
	port := "5432"

	// 接続URLを構築（パスワードなし）
	connectionURL := fmt.Sprintf("postgres://%s@%s:%s/%s?sslmode=verify-full&sslnegotiation=direct",
		username,
		hostname,
		port,
		database,
	)

	fmt.Printf("Connecting to Aurora DSQL at %s\n", hostname)

	// プール設定をパース
	poolConfig, err := pgxpool.ParseConfig(connectionURL)
	if err != nil {
		log.Fatalf("Unable to parse pool config: %v", err)
	}

	// 接続前のフックを設定（トークン生成）
	poolConfig.BeforeConnect = func(ctx context.Context, cfg *pgx.ConnConfig) error {
		// AWS設定をロード
		awsCfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
		if err != nil {
			return fmt.Errorf("failed to load AWS config: %w", err)
		}

		// 認証トークンを生成（有効期限30秒）
		tokenOptions := func(options *auth.TokenOptions) {
			options.ExpiresIn = 30 * time.Second
		}

		// adminユーザーの場合はAdmin用トークンを生成
		var token string
		if username == "admin" {
			token, err = auth.GenerateDBConnectAdminAuthToken(
				ctx,
				hostname,
				region,
				awsCfg.Credentials,
				tokenOptions,
			)
		} else {
			token, err = auth.GenerateDbConnectAuthToken(
				ctx,
				hostname,
				region,
				awsCfg.Credentials,
				tokenOptions,
			)
		}

		if err != nil {
			return fmt.Errorf("failed to generate auth token: %w", err)
		}

		// トークンをパスワードとして設定
		cfg.Password = token
		fmt.Printf("Auth token generated (length: %d)\n", len(token))

		return nil
	}

	// 接続プールの設定
	poolConfig.MaxConns = 10
	poolConfig.MinConns = 2
	poolConfig.MaxConnLifetime = 1 * time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute
	poolConfig.HealthCheckPeriod = 1 * time.Minute

	// コンテキストを作成
	ctx := context.Background()

	// 接続プールを作成
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v", err)
	}
	defer pool.Close()

	// 接続をテスト
	err = pool.Ping(ctx)
	if err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	fmt.Println("Successfully connected to Aurora DSQL!")

	// SELECT version()を実行
	var version string
	err = pool.QueryRow(ctx, "SELECT version()").Scan(&version)
	if err != nil {
		log.Fatalf("Failed to execute query: %v", err)
	}

	// 結果を表示
	fmt.Printf("\nDatabase version:\n%s\n", version)

	// バージョン情報の簡潔な表示
	if strings.Contains(version, "PostgreSQL") {
		parts := strings.Fields(version)
		if len(parts) >= 2 {
			fmt.Printf("\nVersion summary: %s %s\n", parts[0], parts[1])
		}
	}
}