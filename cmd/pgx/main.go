package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
)

// User represents a user in the database
type User struct {
	ID        int
	Name      string
	Email     string
	CreatedAt time.Time
}

// Configuration for data volume (same as GORM version)
const (
	INITIAL_USERS_COUNT = 50000 // 初期データ数
	BATCH_SIZE          = 5000  // バッチサイズ
	UPDATE_COUNT        = 5000  // 更新対象数
	DELETE_COUNT        = 2500  // 削除対象数
	NEW_USERS_COUNT     = 10000 // 新規作成数
)

func main() {
	log.Println("go-postgresql (PGX version) starting up - Performance Test Mode")

	totalStart := time.Now()

	// Database connection string
	connString := "host=127.0.0.1 user=user password=password dbname=go_database port=5432 sslmode=disable"

	// Connect to the database
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, connString)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer conn.Close(ctx)

	log.Println("Database connection successful.")

	// --- Reset database for idempotent run ---
	fmt.Println("\n=== Resetting database for a clean run ===")
	resetStart := time.Now()
	_, err = conn.Exec(ctx, "TRUNCATE TABLE users RESTART IDENTITY")
	if err != nil {
		log.Fatalf("Failed to truncate users table: %v", err)
	}
	resetDuration := time.Since(resetStart)
	fmt.Printf("Table 'users' cleared in %v\n", resetDuration)

	// --- Seed large amount of initial data ---
	fmt.Printf("\n=== Seeding %d initial users ===\n", INITIAL_USERS_COUNT)
	seedStart := time.Now()

	// Generate initial users in batches
	for i := 0; i < INITIAL_USERS_COUNT; i += BATCH_SIZE {
		batchStart := time.Now()
		end := i + BATCH_SIZE
		if end > INITIAL_USERS_COUNT {
			end = INITIAL_USERS_COUNT
		}

		// Prepare batch insert
		batch := &pgx.Batch{}
		for j := i; j < end; j++ {
			name := fmt.Sprintf("User_%06d", j+1)
			email := fmt.Sprintf("user%06d@example.com", j+1)
			batch.Queue("INSERT INTO users (name, email, created_at) VALUES ($1, $2, $3)", name, email, time.Now())
		}

		// Execute batch
		batchResults := conn.SendBatch(ctx, batch)
		for k := 0; k < end-i; k++ {
			_, err := batchResults.Exec()
			if err != nil {
				log.Fatalf("Failed to execute batch insert %d: %v", k, err)
			}
		}
		batchResults.Close()

		batchDuration := time.Since(batchStart)
		fmt.Printf("Batch %d-%d inserted in %v\n", i+1, end, batchDuration)
	}

	seedDuration := time.Since(seedStart)
	fmt.Printf("Initial data seeding completed in %v\n", seedDuration)

	// --- Read: Get user count ---
	fmt.Println("\n=== Reading user count after seeding ===")
	readStart := time.Now()
	var userCount int
	err = conn.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&userCount)
	if err != nil {
		log.Fatalf("Failed to count users: %v", err)
	}
	readDuration := time.Since(readStart)
	fmt.Printf("Found %d users in %v\n", userCount, readDuration)

	// --- Update: Change multiple users' names ---
	fmt.Printf("\n=== Updating %d users ===\n", UPDATE_COUNT)
	updateStart := time.Now()

	// Get users to update
	rows, err := conn.Query(ctx, "SELECT id FROM users LIMIT $1", UPDATE_COUNT)
	if err != nil {
		log.Fatalf("Failed to get users for update: %v", err)
	}

	var userIDs []int
	for rows.Next() {
		var id int
		err := rows.Scan(&id)
		if err != nil {
			log.Printf("Failed to scan user ID: %v", err)
			continue
		}
		userIDs = append(userIDs, id)
	}
	rows.Close()

	// Update users
	for i, userID := range userIDs {
		newName := fmt.Sprintf("Updated_User_%06d", userID)
		_, err := conn.Exec(ctx, "UPDATE users SET name = $1 WHERE id = $2", newName, userID)
		if err != nil {
			log.Printf("Failed to update user ID %d: %v", userID, err)
		}

		if (i+1)%100 == 0 {
			fmt.Printf("Updated %d users...\n", i+1)
		}
	}

	updateDuration := time.Since(updateStart)
	fmt.Printf("Updated %d users in %v\n", len(userIDs), updateDuration)

	// --- Delete: Remove multiple users ---
	fmt.Printf("\n=== Deleting %d users ===\n", DELETE_COUNT)
	deleteStart := time.Now()

	// Get users to delete
	rows, err = conn.Query(ctx, "SELECT id FROM users OFFSET 1000 LIMIT $1", DELETE_COUNT)
	if err != nil {
		log.Fatalf("Failed to get users for deletion: %v", err)
	}

	var deleteIDs []int
	for rows.Next() {
		var id int
		err := rows.Scan(&id)
		if err != nil {
			log.Printf("Failed to scan user ID for deletion: %v", err)
			continue
		}
		deleteIDs = append(deleteIDs, id)
	}
	rows.Close()

	// Delete users
	for i, userID := range deleteIDs {
		_, err := conn.Exec(ctx, "DELETE FROM users WHERE id = $1", userID)
		if err != nil {
			log.Printf("Failed to delete user ID %d: %v", userID, err)
		}

		if (i+1)%100 == 0 {
			fmt.Printf("Deleted %d users...\n", i+1)
		}
	}

	deleteDuration := time.Since(deleteStart)
	fmt.Printf("Deleted %d users in %v\n", len(deleteIDs), deleteDuration)

	// --- Create: Add new users ---
	fmt.Printf("\n=== Creating %d new users ===\n", NEW_USERS_COUNT)
	createStart := time.Now()

	// Generate new users in batches
	for i := 0; i < NEW_USERS_COUNT; i += BATCH_SIZE {
		batchStart := time.Now()
		end := i + BATCH_SIZE
		if end > NEW_USERS_COUNT {
			end = NEW_USERS_COUNT
		}

		// Prepare batch insert for new users
		batch := &pgx.Batch{}
		for j := i; j < end; j++ {
			name := fmt.Sprintf("New_User_%06d", j+1)
			email := fmt.Sprintf("newuser%06d@example.com", j+1)
			batch.Queue("INSERT INTO users (name, email, created_at) VALUES ($1, $2, $3)", name, email, time.Now())
		}

		// Execute batch
		batchResults := conn.SendBatch(ctx, batch)
		for k := 0; k < end-i; k++ {
			_, err := batchResults.Exec()
			if err != nil {
				log.Printf("Failed to execute new user batch insert %d: %v", k, err)
			}
		}
		batchResults.Close()

		batchDuration := time.Since(batchStart)
		fmt.Printf("New batch %d-%d created in %v\n", i+1, end, batchDuration)
	}

	createDuration := time.Since(createStart)
	fmt.Printf("Created %d new users in %v\n", NEW_USERS_COUNT, createDuration)

	// --- Final Read: Get final user count ---
	fmt.Println("\n=== Final user count ===")
	finalReadStart := time.Now()
	err = conn.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&userCount)
	if err != nil {
		log.Fatalf("Failed to count final users: %v", err)
	}
	finalReadDuration := time.Since(finalReadStart)
	fmt.Printf("Final user count: %d (retrieved in %v)\n", userCount, finalReadDuration)

	// --- Performance Summary ---
	totalDuration := time.Since(totalStart)
	fmt.Println("\n==================================================")
	fmt.Println("PGX PERFORMANCE SUMMARY")
	fmt.Println("==================================================")
	fmt.Printf("Reset:          %v\n", resetDuration)
	fmt.Printf("Seed (%d):      %v\n", INITIAL_USERS_COUNT, seedDuration)
	fmt.Printf("Read Count:     %v\n", readDuration)
	fmt.Printf("Update (%d):    %v\n", UPDATE_COUNT, updateDuration)
	fmt.Printf("Delete (%d):    %v\n", DELETE_COUNT, deleteDuration)
	fmt.Printf("Create (%d):    %v\n", NEW_USERS_COUNT, createDuration)
	fmt.Printf("Final Read:     %v\n", finalReadDuration)
	fmt.Println("--------------------------------------------------")
	fmt.Printf("TOTAL TIME:     %v\n", totalDuration)
	fmt.Println("==================================================")
}
