# Go PostgreSQL Example

This repository provides a `docker-compose.yml` to run PostgreSQL and includes performance comparison between GORM and PGX.

## Quick Start

Start the database container in detached mode:

```bash
docker-compose up -d
```

The database listens on `localhost:5432` and stores data in `./data`.

## Performance Benchmark

This project includes two implementations for performance comparison:

- **GORM Version** (`cmd/gorm/main.go`): Uses GORM ORM
- **PGX Version** (`cmd/pgx/main.go`): Uses native PGX driver

### Running the Benchmark

Execute the automated benchmark script:

```bash
./benchmark.sh
```

This script will:
1. Start PostgreSQL container
2. Run the GORM version with performance timing
3. Run the PGX version with performance timing
4. Display detailed performance summaries for comparison

### Manual Execution

You can also run each version individually:

**GORM Version:**
```bash
cd cmd/gorm
go run main.go
```

**PGX Version:**
```bash
cd cmd/pgx
go run main.go
```

### Benchmark Operations

Both versions perform identical operations with large datasets:

- **Reset**: Truncate table and restart ID sequence
- **Seed**: Insert 10,000 initial users in batches of 1,000
- **Read**: Count total users
- **Update**: Update 1,000 users with new names
- **Delete**: Delete 500 users
- **Create**: Insert 2,000 new users in batches
- **Final Read**: Count final users

### Performance Metrics

The benchmark measures:
- Individual operation timing
- Batch processing efficiency
- Total execution time
- Memory usage patterns

### Expected Results

Generally, you can expect:
- **PGX**: Lower latency, higher throughput, less memory usage
- **GORM**: Higher-level abstractions, more memory usage, additional overhead

### Configuration

Data volumes can be adjusted by modifying the constants in each main.go:

```go
const (
    INITIAL_USERS_COUNT = 10000  // 初期データ数
    BATCH_SIZE         = 1000    // バッチサイズ
    UPDATE_COUNT       = 1000    // 更新対象数
    DELETE_COUNT       = 500     // 削除対象数
    NEW_USERS_COUNT    = 2000    // 新規作成数
)
```

Available presets: `small`, `medium`, `large`, `xlarge`
