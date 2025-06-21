# Go PostgreSQL ベンチマーク

このリポジトリは、PostgreSQLを実行するための`docker-compose.yml`を提供し、GORM、PGX、PQのパフォーマンス比較を含んでいます。

## プロジェクト構造

```
go-postgresql/
├── config/
│   └── env.go          # 設定管理（全ベンチマーク共通）
├── cmd/
│   ├── gorm/main.go    # GORM実装
│   ├── pgx/main.go     # PGX実装
│   └── pq/main.go      # PQ実装
├── docker-compose.yml  # PostgreSQLコンテナ設定
└── benchmark.sh        # 自動ベンチマークスクリプト
```

## クイックスタート

データベースコンテナをデタッチモードで起動します：

```bash
docker-compose up -d
```

データベースは`localhost:5432`でリッスンし、データは`./data`に保存されます。

## パフォーマンスベンチマーク

このプロジェクトには、パフォーマンス比較のための3つの実装が含まれています：

- **GORMバージョン** (`cmd/gorm/main.go`): GORM ORMを使用
- **PGXバージョン** (`cmd/pgx/main.go`): ネイティブPGXドライバーを使用
- **PQバージョン** (`cmd/pq/main.go`): `database/sql`とlib/pqドライバーを使用

### ベンチマークの実行

自動ベンチマークスクリプトを実行します：

```bash
./benchmark.sh
```

このスクリプトは以下を実行します：
1. PostgreSQLコンテナを起動
2. GORMバージョンをパフォーマンス計測付きで実行
3. PGXバージョンをパフォーマンス計測付きで実行
4. PQバージョンをパフォーマンス計測付きで実行
5. 比較のための詳細なパフォーマンス要約を表示

### 手動実行

各バージョンを個別に実行することもできます：

**GORMバージョン:**
```bash
cd cmd/gorm
go run main.go
```

**PGXバージョン:**
```bash
cd cmd/pgx
go run main.go
```

**PQバージョン:**
```bash
cd cmd/pq
go run main.go
```

### ベンチマーク操作

各バージョンとも大規模データセットで同一の操作を実行します：

- **Reset**: テーブルを切り詰め、IDシーケンスを再開
- **Seed**: 初期ユーザー50,000件を5,000件のバッチで挿入
- **Read**: 総ユーザー数をカウント
- **Update**: 5,000ユーザーの名前を新しい名前で更新
- **Delete**: 2,500ユーザーを削除
- **Create**: 新しいユーザー10,000件をバッチで挿入
- **Final Read**: 最終ユーザー数をカウント

### パフォーマンス指標

ベンチマークでは以下を測定します：
- 個別操作のタイミング
- バッチ処理の効率性
- 総実行時間
- メモリ使用パターン

### 期待される結果

一般的に、以下のような結果が期待できます：
- **PGX**: 低レイテンシ、高スループット、少ないメモリ使用量
- **GORM**: 高レベルな抽象化、多いメモリ使用量、追加のオーバーヘッド
- **PQ**: 標準的な`database/sql`ドライバー。PGXよりやや遅いがシンプル

### 設定

データ量は`config/env.go`で一元管理されており、以下の値を変更することで調整できます：

```go
// config/env.go
func DefaultConfig() *DatabaseConfig {
    return &DatabaseConfig{
        InitialUsersCount: 50000, // 初期データ数
        BatchSize:         5000,  // バッチサイズ
        UpdateCount:       5000,  // 更新対象数
        DeleteCount:       2500,  // 削除対象数
        NewUsersCount:     10000, // 新規作成数
    }
}
```

この設定は全てのベンチマーク（GORM、PGX、PQ）で共通して使用されるため、一箇所の変更で全ての実装に反映されます。将来的には環境変数や設定ファイルからの読み込みも可能な拡張性のある設計になっています。
