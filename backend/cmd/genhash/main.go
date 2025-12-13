package main

import (
	"cso-erp/internal/security"
	"fmt"
)

func main() {
	// كلمة المرور الجديدة لحساب superadmin
	password := "Admin@123456"

	hash, err := security.HashPassword(password, nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("===========================================")
	fmt.Println("كلمة المرور:", password)
	fmt.Println("===========================================")
	fmt.Println("الهاش:")
	fmt.Println(hash)
	fmt.Println("===========================================")
	fmt.Println("")
	fmt.Println("SQL لتحديث كلمة المرور:")
	fmt.Printf("UPDATE super_system_admins SET password_hash = '%s' WHERE username = 'superadmin';\n", hash)
}
