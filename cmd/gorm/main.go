package main

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// User corresponds to the users table in the database.
type User struct {
	ID        uint `gorm:"primaryKey"`
	Name      string
	Email     string `gorm:"unique"`
	CreatedAt time.Time
}

// Configuration for data volume
const (
	INITIAL_USERS_COUNT = 50000 // 初期データ数
	BATCH_SIZE          = 5000  // バッチサイズ
	UPDATE_COUNT        = 5000  // 更新対象数
	DELETE_COUNT        = 2500  // 削除対象数
	NEW_USERS_COUNT     = 10000 // 新規作成数
)

func main() {
	log.Println("go-postgresql (GORM version) starting up - Performance Test Mode")

	totalStart := time.Now()

	// DSN for connecting to the PostgreSQL database.
	dsn := "host=127.0.0.1 user=user password=password dbname=go_database port=5432 sslmode=disable TimeZone=Asia/Tokyo"

	// Open a connection to the database.
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("Database connection successful.")

	// --- Reset database for idempotent run ---
	fmt.Println("\n=== Resetting database for a clean run ===")
	resetStart := time.Now()
	if err := db.Exec("TRUNCATE TABLE users RESTART IDENTITY").Error; err != nil {
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

		var batchUsers []User
		for j := i; j < end; j++ {
			user := User{
				Name:  fmt.Sprintf("User_%06d", j+1),
				Email: fmt.Sprintf("user%06d@example.com", j+1),
			}
			batchUsers = append(batchUsers, user)
		}

		if err := db.Create(&batchUsers).Error; err != nil {
			log.Fatalf("Failed to seed batch users %d-%d: %v", i+1, end, err)
		}

		batchDuration := time.Since(batchStart)
		fmt.Printf("Batch %d-%d inserted in %v\n", i+1, end, batchDuration)
	}

	seedDuration := time.Since(seedStart)
	fmt.Printf("Initial data seeding completed in %v\n", seedDuration)

	// --- Read: Get user count ---
	fmt.Println("\n=== Reading user count after seeding ===")
	readStart := time.Now()
	var userCount int64
	db.Model(&User{}).Count(&userCount)
	readDuration := time.Since(readStart)
	fmt.Printf("Found %d users in %v\n", userCount, readDuration)

	// --- Update: Change multiple users' names ---
	fmt.Printf("\n=== Updating %d users ===\n", UPDATE_COUNT)
	updateStart := time.Now()

	// Get random users to update
	var usersToUpdate []User
	db.Limit(UPDATE_COUNT).Find(&usersToUpdate)

	for i, user := range usersToUpdate {
		newName := fmt.Sprintf("Updated_User_%06d", user.ID)
		if err := db.Model(&user).Update("Name", newName).Error; err != nil {
			log.Printf("Failed to update user ID %d: %v", user.ID, err)
		}

		if (i+1)%100 == 0 {
			fmt.Printf("Updated %d users...\n", i+1)
		}
	}

	updateDuration := time.Since(updateStart)
	fmt.Printf("Updated %d users in %v\n", len(usersToUpdate), updateDuration)

	// --- Delete: Remove multiple users ---
	fmt.Printf("\n=== Deleting %d users ===\n", DELETE_COUNT)
	deleteStart := time.Now()

	// Get random users to delete
	var usersToDelete []User
	db.Offset(1000).Limit(DELETE_COUNT).Find(&usersToDelete)

	for i, user := range usersToDelete {
		if err := db.Delete(&user).Error; err != nil {
			log.Printf("Failed to delete user ID %d: %v", user.ID, err)
		}

		if (i+1)%100 == 0 {
			fmt.Printf("Deleted %d users...\n", i+1)
		}
	}

	deleteDuration := time.Since(deleteStart)
	fmt.Printf("Deleted %d users in %v\n", len(usersToDelete), deleteDuration)

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

		var newUsers []User
		for j := i; j < end; j++ {
			user := User{
				Name:  fmt.Sprintf("New_User_%06d", j+1),
				Email: fmt.Sprintf("newuser%06d@example.com", j+1),
			}
			newUsers = append(newUsers, user)
		}

		if err := db.Create(&newUsers).Error; err != nil {
			log.Printf("Failed to create batch new users %d-%d: %v", i+1, end, err)
		}

		batchDuration := time.Since(batchStart)
		fmt.Printf("New batch %d-%d created in %v\n", i+1, end, batchDuration)
	}

	createDuration := time.Since(createStart)
	fmt.Printf("Created %d new users in %v\n", NEW_USERS_COUNT, createDuration)

	// --- Final Read: Get final user count ---
	fmt.Println("\n=== Final user count ===")
	finalReadStart := time.Now()
	db.Model(&User{}).Count(&userCount)
	finalReadDuration := time.Since(finalReadStart)
	fmt.Printf("Final user count: %d (retrieved in %v)\n", userCount, finalReadDuration)

	// --- Performance Summary ---
	totalDuration := time.Since(totalStart)
	fmt.Println("\n==================================================")
	fmt.Println("GORM PERFORMANCE SUMMARY")
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
