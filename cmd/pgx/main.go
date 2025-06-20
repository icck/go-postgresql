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

func main() {
	log.Println("go-postgresql (PGX version) starting up")

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
	fmt.Println("\n--- Resetting database for a clean run ---")
	_, err = conn.Exec(ctx, "TRUNCATE TABLE users RESTART IDENTITY")
	if err != nil {
		log.Fatalf("Failed to truncate users table: %v", err)
	}
	fmt.Println("Table 'users' cleared.")

	// Seed initial data
	initialUsers := []User{
		{Name: "Alice", Email: "alice@example.com"},
		{Name: "Bob", Email: "bob@example.com"},
	}

	for _, user := range initialUsers {
		_, err = conn.Exec(ctx, "INSERT INTO users (name, email, created_at) VALUES ($1, $2, $3)", user.Name, user.Email, time.Now())
		if err != nil {
			log.Fatalf("Failed to seed initial user %s: %v", user.Name, err)
		}
	}
	fmt.Println("Initial data seeded.")

	// --- Read: Get all users ---
	fmt.Println("\n--- Reading all users after seeding ---")
	users, err := getAllUsers(ctx, conn)
	if err != nil {
		log.Fatalf("Failed to retrieve users: %v", err)
	}

	fmt.Printf("Found %d users after seeding:\n", len(users))
	for _, user := range users {
		fmt.Printf("- ID: %d, Name: %s, Email: %s, CreatedAt: %s\n", user.ID, user.Name, user.Email, user.CreatedAt.Format(time.RFC3339))
	}

	// --- Update: Change Bob's name to Bobby ---
	fmt.Println("\n--- Updating Bob's name to Bobby ---")
	bob, err := getUserByName(ctx, conn, "Bob")
	if err != nil {
		log.Printf("Failed to find user Bob: %v", err)
	} else if bob != nil {
		_, err = conn.Exec(ctx, "UPDATE users SET name = $1 WHERE id = $2", "Bobby", bob.ID)
		if err != nil {
			log.Fatalf("Failed to update user: %v", err)
		}
		fmt.Printf("Updated user: Bobby (ID: %d)\n", bob.ID)
	} else {
		fmt.Println("User Bob not found.")
	}

	// --- Delete: Remove Alice ---
	fmt.Println("\n--- Deleting Alice ---")
	alice, err := getUserByName(ctx, conn, "Alice")
	if err != nil {
		log.Printf("Failed to find user Alice: %v", err)
	} else if alice != nil {
		_, err = conn.Exec(ctx, "DELETE FROM users WHERE id = $1", alice.ID)
		if err != nil {
			log.Fatalf("Failed to delete user: %v", err)
		}
		fmt.Printf("Deleted user: Alice (ID: %d)\n", alice.ID)
	} else {
		fmt.Println("User Alice not found.")
	}

	// --- Create: Add a new user 'Charlie' ---
	fmt.Println("\n--- Creating a new user 'Charlie' ---")
	var newUserID int
	err = conn.QueryRow(ctx, "INSERT INTO users (name, email, created_at) VALUES ($1, $2, $3) RETURNING id", "Charlie", "charlie@example.com", time.Now()).Scan(&newUserID)
	if err != nil {
		log.Fatalf("Failed to create user: %v", err)
	}
	fmt.Printf("Created user: Charlie (ID: %d)\n", newUserID)

	// --- Read again: Get all users to see the final state ---
	fmt.Println("\n--- Reading all users for the final time ---")
	users, err = getAllUsers(ctx, conn)
	if err != nil {
		log.Fatalf("Failed to retrieve users: %v", err)
	}

	fmt.Printf("Found %d users:\n", len(users))
	for _, user := range users {
		fmt.Printf("- ID: %d, Name: %s, Email: %s, CreatedAt: %s\n", user.ID, user.Name, user.Email, user.CreatedAt.Format(time.RFC3339))
	}
}

// getAllUsers retrieves all users from the database
func getAllUsers(ctx context.Context, conn *pgx.Conn) ([]User, error) {
	rows, err := conn.Query(ctx, "SELECT id, name, email, created_at FROM users ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, rows.Err()
}

// getUserByName retrieves a user by name
func getUserByName(ctx context.Context, conn *pgx.Conn, name string) (*User, error) {
	var user User
	err := conn.QueryRow(ctx, "SELECT id, name, email, created_at FROM users WHERE name = $1", name).Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // User not found
		}
		return nil, err
	}
	return &user, nil
}