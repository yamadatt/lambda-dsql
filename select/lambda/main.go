package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dsql/auth"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// グローバル変数でプールを保持（Lambda実行間で再利用）
var pool *pgxpool.Pool

// Response はAPI Gatewayへのレスポンス構造体
type Response struct {
	ButtonClicks []map[string]interface{} `json:"button_clicks"`
	Count        int                      `json:"count"`
	Message      string                   `json:"message"`
	Timestamp    string                   `json:"timestamp"`
}

// ErrorResponse はエラーレスポンス構造体
type ErrorResponse struct {
	Error     string `json:"error"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

func init() {
	// Lambda初期化時にプールを作成
	ctx := context.Background()
	var err error
	pool, err = createPool(ctx)
	if err != nil {
		log.Printf("Failed to create connection pool during init: %v", err)
	}
}

func createPool(ctx context.Context) (*pgxpool.Pool, error) {
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

	// プール設定をパース
	poolConfig, err := pgxpool.ParseConfig(connectionURL)
	if err != nil {
		return nil, fmt.Errorf("unable to parse pool config: %v", err)
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
		log.Printf("Auth token generated (length: %d)", len(token))

		return nil
	}

	// Lambda用の接続プール設定（控えめに設定）
	poolConfig.MaxConns = 5
	poolConfig.MinConns = 1
	poolConfig.MaxConnLifetime = 5 * time.Minute
	poolConfig.MaxConnIdleTime = 1 * time.Minute
	poolConfig.HealthCheckPeriod = 30 * time.Second

	// 接続プールを作成
	newPool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %v", err)
	}

	// 接続をテスト
	err = newPool.Ping(ctx)
	if err != nil {
		newPool.Close()
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	log.Println("Successfully connected to Aurora DSQL!")
	return newPool, nil
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Printf("Received request: Method=%s, Path=%s", request.HTTPMethod, request.Path)

	// プールが初期化されていない場合は作成
	if pool == nil {
		var err error
		pool, err = createPool(ctx)
		if err != nil {
			errorResponse := ErrorResponse{
				Error:     "Database Connection Error",
				Message:   fmt.Sprintf("Failed to connect to database: %v", err),
				Timestamp: time.Now().UTC().Format(time.RFC3339),
			}
			body, _ := json.Marshal(errorResponse)
			return events.APIGatewayProxyResponse{
				StatusCode: 500,
				Headers: map[string]string{
					"Content-Type":                "application/json",
					"Access-Control-Allow-Origin": "*",
				},
				Body: string(body),
			}, nil
		}
	}

	// SELECT * FROM button_clicks を実行
	rows, err := pool.Query(ctx, "SELECT * FROM button_clicks ORDER BY id")
	if err != nil {
		// エラーの場合、プールをリセット
		if pool != nil {
			pool.Close()
			pool = nil
		}

		errorResponse := ErrorResponse{
			Error:     "Query Execution Error",
			Message:   fmt.Sprintf("Failed to execute query: %v", err),
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}
		body, _ := json.Marshal(errorResponse)
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Headers: map[string]string{
				"Content-Type":                "application/json",
				"Access-Control-Allow-Origin": "*",
			},
			Body: string(body),
		}, nil
	}
	defer rows.Close()

	// 列の情報を取得
	fieldDescriptions := rows.FieldDescriptions()
	columnCount := len(fieldDescriptions)

	// 結果を格納するスライス
	var buttonClicks []map[string]interface{}

	// 各行を処理
	for rows.Next() {
		// 動的に値を格納するスライスを作成
		values := make([]interface{}, columnCount)
		valuePtrs := make([]interface{}, columnCount)
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		// 行をスキャン
		err := rows.Scan(valuePtrs...)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		// 結果をマップに変換
		rowMap := make(map[string]interface{})
		for i, fieldDesc := range fieldDescriptions {
			columnName := fieldDesc.Name
			value := values[i]

			// null値の処理
			if value == nil {
				rowMap[columnName] = nil
			} else {
				// 値の型に応じて適切に設定
				switch v := value.(type) {
				case []byte:
					// bytea型やtext型の場合
					rowMap[columnName] = string(v)
				case time.Time:
					// timestamp型の場合
					rowMap[columnName] = v.Format(time.RFC3339)
				default:
					rowMap[columnName] = value
				}
			}
		}

		buttonClicks = append(buttonClicks, rowMap)
	}

	// 読み取りエラーをチェック
	if err = rows.Err(); err != nil {
		errorResponse := ErrorResponse{
			Error:     "Row Reading Error",
			Message:   fmt.Sprintf("Failed to read rows: %v", err),
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}
		body, _ := json.Marshal(errorResponse)
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Headers: map[string]string{
				"Content-Type":                "application/json",
				"Access-Control-Allow-Origin": "*",
			},
			Body: string(body),
		}, nil
	}

	// 成功レスポンスを作成
	response := Response{
		ButtonClicks: buttonClicks,
		Count:        len(buttonClicks),
		Message:      "Successfully retrieved button clicks",
		Timestamp:    time.Now().UTC().Format(time.RFC3339),
	}

	body, err := json.Marshal(response)
	if err != nil {
		errorResponse := ErrorResponse{
			Error:     "Response Serialization Error",
			Message:   fmt.Sprintf("Failed to serialize response: %v", err),
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}
		errorBody, _ := json.Marshal(errorResponse)
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Headers: map[string]string{
				"Content-Type":                "application/json",
				"Access-Control-Allow-Origin": "*",
			},
			Body: string(errorBody),
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type":                "application/json",
			"Access-Control-Allow-Origin": "*",
		},
		Body: string(body),
	}, nil
}

func main() {
	lambda.Start(handler)
}