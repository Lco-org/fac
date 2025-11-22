package main

import (
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Generate hash for Admin@123
	password1 := "Admin@123"
	hash1, err := bcrypt.GenerateFromPassword([]byte(password1), 10)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("=== Admin User ===")
	fmt.Println("Password:", password1)
	fmt.Println("Hash:", string(hash1))
	fmt.Println()

	// Generate hash for 123456
	password2 := "123456"
	hash2, err := bcrypt.GenerateFromPassword([]byte(password2), 10)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("=== Ezy User ===")
	fmt.Println("Password:", password2)
	fmt.Println("Hash:", string(hash2))
	fmt.Println()

	// SQL Commands
	fmt.Println("=== SQL Commands ===")
	fmt.Printf("-- Update admin password to: %s\n", password1)
	fmt.Printf("UPDATE users SET password_hash = '%s' WHERE username = 'admin';\n\n", string(hash1))

	fmt.Printf("-- Add Ezy user with password: %s\n", password2)
	fmt.Printf("INSERT INTO users (username, email, password_hash, full_name, full_name_ar, role, is_active)\n")
	fmt.Printf("VALUES ('Ezy', 'ezy@cso-erp.local', '%s', 'Ezy User', 'المستخدم إيزي', 'admin', TRUE)\n", string(hash2))
	fmt.Printf("ON CONFLICT (username) DO UPDATE SET password_hash = EXCLUDED.password_hash;\n")
}
