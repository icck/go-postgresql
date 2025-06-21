package config

// DatabaseConfig holds database performance test configuration
type DatabaseConfig struct {
	InitialUsersCount int // 初期データ数
	BatchSize         int // バッチサイズ
	UpdateCount       int // 更新対象数
	DeleteCount       int // 削除対象数
	NewUsersCount     int // 新規作成数
}

// DefaultConfig returns the default configuration for performance tests
func DefaultConfig() *DatabaseConfig {
	return &DatabaseConfig{
		InitialUsersCount: 50000, // 初期データ数
		BatchSize:         5000,  // バッチサイズ
		UpdateCount:       5000,  // 更新対象数
		DeleteCount:       2500,  // 削除対象数
		NewUsersCount:     10000, // 新規作成数
	}
}

// GetConfig returns the current configuration
// 将来的には環境変数や設定ファイルから読み込む拡張も可能
func GetConfig() *DatabaseConfig {
	return DefaultConfig()
}
