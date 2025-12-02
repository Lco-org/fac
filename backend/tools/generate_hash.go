package main

import (
	"fmt"
	"os"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run generate_hash.go <password>")
		os.Exit(1)
	}

	password := os.Args[1]

	// Generate bcrypt hash with cost 10
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("Password:", password)
	fmt.Println("Bcrypt Hash:", string(hash))
	fmt.Println("\nSQL Command:")
	fmt.Printf("UPDATE users SET password_hash = '%s' WHERE username = 'admin';\n", string(hash))
}
