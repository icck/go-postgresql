package main

// BenchmarkConfig holds the configuration for performance testing
type BenchmarkConfig struct {
	InitialUsersCount int
	BatchSize         int
	UpdateCount       int
	DeleteCount       int
	NewUsersCount     int
}

// GetBenchmarkConfig returns different preset configurations
func GetBenchmarkConfig(preset string) BenchmarkConfig {
	switch preset {
	case "small":
		return BenchmarkConfig{
			InitialUsersCount: 1000,
			BatchSize:         100,
			UpdateCount:       100,
			DeleteCount:       50,
			NewUsersCount:     200,
		}
	case "medium":
		return BenchmarkConfig{
			InitialUsersCount: 10000,
			BatchSize:         1000,
			UpdateCount:       1000,
			DeleteCount:       500,
			NewUsersCount:     2000,
		}
	case "large":
		return BenchmarkConfig{
			InitialUsersCount: 100000,
			BatchSize:         5000,
			UpdateCount:       10000,
			DeleteCount:       5000,
			NewUsersCount:     20000,
		}
	case "xlarge":
		return BenchmarkConfig{
			InitialUsersCount: 1000000,
			BatchSize:         10000,
			UpdateCount:       50000,
			DeleteCount:       25000,
			NewUsersCount:     100000,
		}
	default: // medium
		return BenchmarkConfig{
			InitialUsersCount: 10000,
			BatchSize:         1000,
			UpdateCount:       1000,
			DeleteCount:       500,
			NewUsersCount:     2000,
		}
	}
}
