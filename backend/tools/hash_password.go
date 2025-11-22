package main

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	password := "123456"

	// Generate bcrypt hash with cost 10
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("Password:", password)
	fmt.Println("Bcrypt Hash:", string(hash))
	fmt.Println("\nSQL Command:")
	fmt.Printf("UPDATE users SET password_hash = '%s' WHERE username = 'Ezy';\n", string(hash))
}
