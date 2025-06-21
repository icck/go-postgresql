package main

import (
	"fmt"
	"log"
	"time"

	"go-postgresql/config"

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

func main() {
	log.Println("go-postgresql (GORM version) starting up - Performance Test Mode")

	// Load configuration
	cfg := config.GetConfig()

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
	fmt.Printf("\n=== Seeding %d initial users ===\n", cfg.InitialUsersCount)
	seedStart := time.Now()

	// Generate initial users in batches
	for i := 0; i < cfg.InitialUsersCount; i += cfg.BatchSize {
		batchStart := time.Now()
		end := i + cfg.BatchSize
		if end > cfg.InitialUsersCount {
			end = cfg.InitialUsersCount
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
	fmt.Printf("\n=== Updating %d users ===\n", cfg.UpdateCount)
	updateStart := time.Now()

	// Get random users to update
	var usersToUpdate []User
	db.Limit(cfg.UpdateCount).Find(&usersToUpdate)

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
	fmt.Printf("\n=== Deleting %d users ===\n", cfg.DeleteCount)
	deleteStart := time.Now()

	// Get random users to delete
	var usersToDelete []User
	db.Offset(1000).Limit(cfg.DeleteCount).Find(&usersToDelete)

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
	fmt.Printf("\n=== Creating %d new users ===\n", cfg.NewUsersCount)
	createStart := time.Now()

	// Generate new users in batches
	for i := 0; i < cfg.NewUsersCount; i += cfg.BatchSize {
		batchStart := time.Now()
		end := i + cfg.BatchSize
		if end > cfg.NewUsersCount {
			end = cfg.NewUsersCount
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
	fmt.Printf("Created %d new users in %v\n", cfg.NewUsersCount, createDuration)

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
