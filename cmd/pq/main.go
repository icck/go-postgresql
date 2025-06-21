package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"go-postgresql/config"

	"strings"

	_ "github.com/lib/pq"
)

// User represents a user in the database
// (structure mirrors other implementations but isn't used directly)
type User struct {
	ID        int
	Name      string
	Email     string
	CreatedAt time.Time
}

// buildPlaceholders generates a string of placeholders for SQL IN clauses.
// Example: buildPlaceholders(3, 1) -> "$1, $2, $3"
func buildPlaceholders(count, start int) string {
	placeholders := make([]string, count)
	for i := 0; i < count; i++ {
		placeholders[i] = fmt.Sprintf("$%d", start+i)
	}
	return strings.Join(placeholders, ",")
}

func main() {
	log.Println("go-postgresql (PQ version) starting up - Performance Test Mode")

	// Load configuration
	cfg := config.GetConfig()

	totalStart := time.Now()

	connStr := "host=127.0.0.1 user=user password=password dbname=go_database port=5432 sslmode=disable"

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Println("Database connection successful.")

	// --- Reset database for idempotent run ---
	fmt.Println("\n=== Resetting database for a clean run ===")
	resetStart := time.Now()
	if _, err := db.Exec("TRUNCATE TABLE users RESTART IDENTITY"); err != nil {
		log.Fatalf("Failed to truncate users table: %v", err)
	}
	resetDuration := time.Since(resetStart)
	fmt.Printf("Table 'users' cleared in %v\n", resetDuration)

	// --- Seed large amount of initial data ---
	fmt.Printf("\n=== Seeding %d initial users ===\n", cfg.InitialUsersCount)
	seedStart := time.Now()
	// Use multi-row INSERT for bulk seeding
	for i := 0; i < cfg.InitialUsersCount; i += cfg.BatchSize {
		batchStart := time.Now()
		end := i + cfg.BatchSize
		if end > cfg.InitialUsersCount {
			end = cfg.InitialUsersCount
		}

		valueStrings := make([]string, 0, end-i)
		args := make([]interface{}, 0, (end-i)*3)
		argIndex := 1
		for j := i; j < end; j++ {
			valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d)", argIndex, argIndex+1, argIndex+2))
			argIndex += 3
			args = append(args, fmt.Sprintf("User_%06d", j+1), fmt.Sprintf("user%06d@example.com", j+1), time.Now())
		}

		query := fmt.Sprintf("INSERT INTO users (name, email, created_at) VALUES %s", strings.Join(valueStrings, ","))
		if _, err := db.Exec(query, args...); err != nil {
			log.Fatalf("Failed to bulk insert seed data: %v", err)
		}

		batchDuration := time.Since(batchStart)
		fmt.Printf("Batch %d-%d inserted in %v\n", i+1, end, batchDuration)
	}
	seedDuration := time.Since(seedStart)
	fmt.Printf("Initial data seeding completed in %v\n", seedDuration)

	// --- Read: Get user count ---
	fmt.Println("\n=== Reading user count after seeding ===")
	readStart := time.Now()
	var userCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount); err != nil {
		log.Fatalf("Failed to count users: %v", err)
	}
	readDuration := time.Since(readStart)
	fmt.Printf("Found %d users in %v\n", userCount, readDuration)

	// --- Update: Change multiple users' names ---
	fmt.Printf("\n=== Updating %d users ===\n", cfg.UpdateCount)
	updateStart := time.Now()

	rows, err := db.Query("SELECT id FROM users LIMIT $1", cfg.UpdateCount)
	if err != nil {
		log.Fatalf("Failed to get users for update: %v", err)
	}
	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			log.Fatalf("Failed to scan user ID: %v", err)
		}
		ids = append(ids, id)
	}
	rows.Close()

	if len(ids) > 0 {
		newName := "Updated_User_Bulk_PQ"
		query := fmt.Sprintf("UPDATE users SET name = $1 WHERE id IN (%s)", buildPlaceholders(len(ids), 2))

		args := make([]interface{}, len(ids)+1)
		args[0] = newName
		for i, id := range ids {
			args[i+1] = id
		}

		if _, err := db.Exec(query, args...); err != nil {
			log.Fatalf("Failed to bulk update users: %v", err)
		}
	}
	updateDuration := time.Since(updateStart)
	fmt.Printf("Updated %d users in %v\n", len(ids), updateDuration)

	// --- Delete: Remove multiple users ---
	fmt.Printf("\n=== Deleting %d users ===\n", cfg.DeleteCount)
	deleteStart := time.Now()
	rows, err = db.Query("SELECT id FROM users OFFSET 1000 LIMIT $1", cfg.DeleteCount)
	if err != nil {
		log.Fatalf("Failed to get users for deletion: %v", err)
	}
	var deleteIDs []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			log.Printf("Failed to scan user ID for deletion: %v", err)
			continue
		}
		deleteIDs = append(deleteIDs, id)
	}
	rows.Close()

	if len(deleteIDs) > 0 {
		query := fmt.Sprintf("DELETE FROM users WHERE id IN (%s)", buildPlaceholders(len(deleteIDs), 1))

		args := make([]interface{}, len(deleteIDs))
		for i, id := range deleteIDs {
			args[i] = id
		}

		if _, err := db.Exec(query, args...); err != nil {
			log.Fatalf("Failed to bulk delete users: %v", err)
		}
	}
	deleteDuration := time.Since(deleteStart)
	fmt.Printf("Deleted %d users in %v\n", len(deleteIDs), deleteDuration)

	// --- Create: Add new users ---
	fmt.Printf("\n=== Creating %d new users ===\n", cfg.NewUsersCount)
	createStart := time.Now()
	// Use multi-row INSERT for bulk creation
	for i := 0; i < cfg.NewUsersCount; i += cfg.BatchSize {
		batchStart := time.Now()
		end := i + cfg.BatchSize
		if end > cfg.NewUsersCount {
			end = cfg.NewUsersCount
		}

		valueStrings := make([]string, 0, end-i)
		args := make([]interface{}, 0, (end-i)*3)
		argIndex := 1
		for j := i; j < end; j++ {
			valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d)", argIndex, argIndex+1, argIndex+2))
			argIndex += 3
			args = append(args, fmt.Sprintf("New_User_%06d", j+1), fmt.Sprintf("newuser%06d@example.com", j+1), time.Now())
		}

		query := fmt.Sprintf("INSERT INTO users (name, email, created_at) VALUES %s", strings.Join(valueStrings, ","))
		if _, err := db.Exec(query, args...); err != nil {
			log.Fatalf("Failed to bulk insert new data: %v", err)
		}

		batchDuration := time.Since(batchStart)
		fmt.Printf("New batch %d-%d created in %v\n", i+1, end, batchDuration)
	}
	createDuration := time.Since(createStart)
	fmt.Printf("Created %d new users in %v\n", cfg.NewUsersCount, createDuration)

	// --- Final Read: Get final user count ---
	fmt.Println("\n=== Final user count ===")
	finalReadStart := time.Now()
	if err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount); err != nil {
		log.Fatalf("Failed to count final users: %v", err)
	}
	finalReadDuration := time.Since(finalReadStart)
	fmt.Printf("Final user count: %d (retrieved in %v)\n", userCount, finalReadDuration)

	// --- Performance Summary ---
	totalDuration := time.Since(totalStart)
	fmt.Println("\n==================================================")
	fmt.Println("PQ PERFORMANCE SUMMARY")
	fmt.Println("==================================================")
	fmt.Printf("Reset:          %v\n", resetDuration)
	fmt.Printf("Seed (%d):      %v\n", cfg.InitialUsersCount, seedDuration)
	fmt.Printf("Read Count:     %v\n", readDuration)
	fmt.Printf("Update (%d):    %v\n", cfg.UpdateCount, updateDuration)
	fmt.Printf("Delete (%d):    %v\n", cfg.DeleteCount, deleteDuration)
	fmt.Printf("Create (%d):    %v\n", cfg.NewUsersCount, createDuration)
	fmt.Printf("Final Read:     %v\n", finalReadDuration)
	fmt.Println("--------------------------------------------------")
	fmt.Printf("TOTAL TIME:     %v\n", totalDuration)
	fmt.Println("==================================================")
}
