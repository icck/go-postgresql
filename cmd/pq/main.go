package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

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

const (
	INITIAL_USERS_COUNT = 50000 // 初期データ数
	BATCH_SIZE          = 5000  // バッチサイズ
	UPDATE_COUNT        = 5000  // 更新対象数
	DELETE_COUNT        = 2500  // 削除対象数
	NEW_USERS_COUNT     = 10000 // 新規作成数
)

func main() {
	log.Println("go-postgresql (PQ version) starting up - Performance Test Mode")

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
	fmt.Printf("\n=== Seeding %d initial users ===\n", INITIAL_USERS_COUNT)
	seedStart := time.Now()
	for i := 0; i < INITIAL_USERS_COUNT; i += BATCH_SIZE {
		batchStart := time.Now()
		end := i + BATCH_SIZE
		if end > INITIAL_USERS_COUNT {
			end = INITIAL_USERS_COUNT
		}

		tx, err := db.Begin()
		if err != nil {
			log.Fatalf("Failed to begin transaction: %v", err)
		}

		stmt, err := tx.Prepare("INSERT INTO users (name, email, created_at) VALUES ($1, $2, $3)")
		if err != nil {
			log.Fatalf("Failed to prepare statement: %v", err)
		}

		for j := i; j < end; j++ {
			name := fmt.Sprintf("User_%06d", j+1)
			email := fmt.Sprintf("user%06d@example.com", j+1)
			if _, err := stmt.Exec(name, email, time.Now()); err != nil {
				log.Fatalf("Failed to insert user %d: %v", j+1, err)
			}
		}
		stmt.Close()
		if err := tx.Commit(); err != nil {
			log.Fatalf("Failed to commit batch: %v", err)
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
	fmt.Printf("\n=== Updating %d users ===\n", UPDATE_COUNT)
	updateStart := time.Now()

	rows, err := db.Query("SELECT id FROM users LIMIT $1", UPDATE_COUNT)
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

	for i, id := range ids {
		newName := fmt.Sprintf("Updated_User_%06d", id)
		if _, err := db.Exec("UPDATE users SET name = $1 WHERE id = $2", newName, id); err != nil {
			log.Printf("Failed to update user ID %d: %v", id, err)
		}
		if (i+1)%100 == 0 {
			fmt.Printf("Updated %d users...\n", i+1)
		}
	}
	updateDuration := time.Since(updateStart)
	fmt.Printf("Updated %d users in %v\n", len(ids), updateDuration)

	// --- Delete: Remove multiple users ---
	fmt.Printf("\n=== Deleting %d users ===\n", DELETE_COUNT)
	deleteStart := time.Now()
	rows, err = db.Query("SELECT id FROM users OFFSET 1000 LIMIT $1", DELETE_COUNT)
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

	for i, id := range deleteIDs {
		if _, err := db.Exec("DELETE FROM users WHERE id = $1", id); err != nil {
			log.Printf("Failed to delete user ID %d: %v", id, err)
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
	for i := 0; i < NEW_USERS_COUNT; i += BATCH_SIZE {
		batchStart := time.Now()
		end := i + BATCH_SIZE
		if end > NEW_USERS_COUNT {
			end = NEW_USERS_COUNT
		}

		tx, err := db.Begin()
		if err != nil {
			log.Fatalf("Failed to begin transaction: %v", err)
		}

		stmt, err := tx.Prepare("INSERT INTO users (name, email, created_at) VALUES ($1, $2, $3)")
		if err != nil {
			log.Fatalf("Failed to prepare statement: %v", err)
		}

		for j := i; j < end; j++ {
			name := fmt.Sprintf("New_User_%06d", j+1)
			email := fmt.Sprintf("newuser%06d@example.com", j+1)
			if _, err := stmt.Exec(name, email, time.Now()); err != nil {
				log.Printf("Failed to insert new user %d: %v", j+1, err)
			}
		}
		stmt.Close()
		if err := tx.Commit(); err != nil {
			log.Fatalf("Failed to commit batch: %v", err)
		}

		batchDuration := time.Since(batchStart)
		fmt.Printf("New batch %d-%d created in %v\n", i+1, end, batchDuration)
	}
	createDuration := time.Since(createStart)
	fmt.Printf("Created %d new users in %v\n", NEW_USERS_COUNT, createDuration)

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
