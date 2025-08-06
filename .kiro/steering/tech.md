# 技術スタック

## 言語・ランタイム
- **Go 1.23.8** - メインプログラミング言語
- **PostgreSQL** - データベースシステム（Docker経由で最新版）

## データベースライブラリ
- **GORM v1.30.0** - postgres driver v1.6.0付きORM
- **PGX v5.7.5** - ネイティブPostgreSQLドライバー
- **lib/pq v1.10.9** - database/sql用PostgreSQLドライバー

## インフラストラクチャ
- **Docker Compose** - PostgreSQLコンテナ化
- **Bash** - 自動化スクリプト

## ビルドシステム・コマンド

### 前提条件
```bash
# Dockerが動作していることを確認
docker info

# Go依存関係をインストール
go mod tidy
```

### データベース設定
```bash
# PostgreSQLコンテナを起動
docker-compose up -d

# データベースの準備状況を確認
docker-compose exec postgres pg_isready -U user -d go_database
```

### ベンチマーク実行
```bash
# 完全なベンチマークスイートを実行
./benchmark.sh

# 個別実装を実行
cd cmd/gorm && go run main.go
cd cmd/pgx && go run main.go  
cd cmd/pq && go run main.go
```

### 開発コマンド
```bash
# すべての実装をビルド
go build ./cmd/gorm
go build ./cmd/pgx
go build ./cmd/pq

# 新しい実行のためにデータベースをクリーン
docker-compose exec postgres psql -U user -d go_database -c "TRUNCATE TABLE users RESTART IDENTITY"

# コンテナを停止
docker-compose down
```

## データベース設定
- **ホスト**: localhost:5432
- **データベース**: go_database
- **ユーザー**: user
- **パスワード**: password
- **SSLモード**: 無効
- **タイムゾーン**: Asia/Tokyo（GORMのみ）