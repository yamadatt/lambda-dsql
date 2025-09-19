
## 背景

以下のエンドポイントをもつAurora DSQLがある。

```
guabumyfv3jxv2ymjmqtbjqmjq.dsql.ap-northeast-1.on.aws
```

マネージメントコンソールのクラウドシェルで以下のコマンドで接続できる。

```bash
dsql-connect --hostname guabumyfv3jxv2ymjmqtbjqmjq.dsql.ap-northeast-1.on.aws --region ap-northeast-1 --database postgres --username admin
```

接続後、以下のSQLを投入すると

```
SELECT version();
```

以下の結果が返却される。

```
    version    
---------------
 PostgreSQL 16
(1 row)
```

## 要求

この上記のAurora DSQLに接続して、```SELECT version();```するGoのアプリケーションを書いてほしい。

以下のAWSのサンプルを参考に作ってほしい。

https://github.com/aws-samples/aurora-dsql-samples/tree/main/go/pgx
