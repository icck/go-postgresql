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

func main() {
	log.Println("go-postgresql starting up")

	// DSN for connecting to the PostgreSQL database.
	dsn := "host=127.0.0.1 user=user password=password dbname=go_database port=5432 sslmode=disable TimeZone=Asia/Tokyo"

	// Open a connection to the database.
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("Database connection successful.")

	// --- Reset database for idempotent run ---
	fmt.Println("\n--- Resetting database for a clean run ---")
	// Using TRUNCATE is fast and resets the ID sequence.
	if err := db.Exec("TRUNCATE TABLE users RESTART IDENTITY").Error; err != nil {
		log.Fatalf("Failed to truncate users table: %v", err)
	}
	fmt.Println("Table 'users' cleared.")

	initialUsers := []User{
		{Name: "Alice", Email: "alice@example.com"},
		{Name: "Bob", Email: "bob@example.com"},
	}
	if err := db.Create(&initialUsers).Error; err != nil {
		log.Fatalf("Failed to seed initial users: %v", err)
	}
	fmt.Println("Initial data seeded.")

	// --- Read: Get all users ---
	var users []User
	result := db.Find(&users)
	if result.Error != nil {
		log.Fatalf("Failed to retrieve users: %v", result.Error)
	}

	fmt.Printf("\nFound %d users after seeding:\n", result.RowsAffected)
	for _, user := range users {
		fmt.Printf("- ID: %d, Name: %s, Email: %s, CreatedAt: %s\n", user.ID, user.Name, user.Email, user.CreatedAt.Format(time.RFC3339))
	}

	// --- Update: Change Bob's name to Bobby ---
	fmt.Println("\n--- Updating Bob's name to Bobby ---")
	var bob User
	db.First(&bob, "name = ?", "Bob")
	if bob.ID != 0 {
		db.Model(&bob).Update("Name", "Bobby")
		fmt.Printf("Updated user: %s (ID: %d)\n", bob.Name, bob.ID)
	} else {
		fmt.Println("User Bob not found.")
	}

	// --- Delete: Remove Alice ---
	fmt.Println("\n--- Deleting Alice ---")
	var alice User
	db.First(&alice, "name = ?", "Alice")
	if alice.ID != 0 {
		deleteResult := db.Delete(&alice)
		if deleteResult.Error != nil {
			log.Fatalf("Failed to delete user: %v", deleteResult.Error)
		}
		fmt.Printf("Deleted user: Alice (ID: %d)\n", alice.ID)
	} else {
		fmt.Println("User Alice not found.")
	}

	// --- Create: Add a new user 'Charlie' ---
	fmt.Println("\n--- Creating a new user 'Charlie' ---")
	newUser := User{Name: "Charlie", Email: "charlie@example.com"}
	createResult := db.Create(&newUser)
	if createResult.Error != nil {
		log.Fatalf("Failed to create user: %v", createResult.Error)
	}
	fmt.Printf("Created user: %s (ID: %d)\n", newUser.Name, newUser.ID)

	// --- Read again: Get all users to see the final state ---
	fmt.Println("\n--- Reading all users for the final time ---")
	result = db.Find(&users)
	if result.Error != nil {
		log.Fatalf("Failed to retrieve users: %v", result.Error)
	}

	fmt.Printf("Found %d users:\n", result.RowsAffected)
	for _, user := range users {
		fmt.Printf("- ID: %d, Name: %s, Email: %s, CreatedAt: %s\n", user.ID, user.Name, user.Email, user.CreatedAt.Format(time.RFC3339))
	}
}
