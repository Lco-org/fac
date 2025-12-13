package main

import (
	"cso-erp/internal/config"
	"cso-erp/internal/security"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	cfg := config.Load()

	fmt.Println("===========================================")
	fmt.Println("فحص قاعدة بيانات erp_master")
	fmt.Println("===========================================")

	// Connect to erp_master
	masterConnStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.MultiTenant.MasterDBName, // erp_master
		cfg.Database.SSLMode,
	)

	fmt.Printf("الاتصال بـ: %s\n", cfg.MultiTenant.MasterDBName)

	db, err := sql.Open("postgres", masterConnStr)
	if err != nil {
		log.Fatalf("❌ فشل الاتصال: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("❌ فشل Ping: %v", err)
	}
	fmt.Println("✅ الاتصال ناجح!")

	// Check if super_system_admins table exists
	fmt.Println("\n===========================================")
	fmt.Println("فحص جدول super_system_admins")
	fmt.Println("===========================================")

	var tableExists bool
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_name = 'super_system_admins'
		)
	`).Scan(&tableExists)

	if err != nil {
		log.Fatalf("❌ خطأ في فحص الجدول: %v", err)
	}

	if !tableExists {
		fmt.Println("❌ الجدول super_system_admins غير موجود!")
		fmt.Println("يجب تشغيل: database/migrations/master_schema.sql")
		return
	}

	fmt.Println("✅ الجدول موجود!")

	// List users
	fmt.Println("\n===========================================")
	fmt.Println("المستخدمين في الجدول")
	fmt.Println("===========================================")

	rows, err := db.Query(`
		SELECT id, username, email, is_active, LEFT(password_hash, 50) as hash_preview
		FROM super_system_admins
	`)
	if err != nil {
		log.Fatalf("❌ خطأ في قراءة المستخدمين: %v", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var id, username, email, hashPreview string
		var isActive bool
		rows.Scan(&id, &username, &email, &isActive, &hashPreview)
		count++
		fmt.Printf("%d. username: %s | email: %s | active: %v\n", count, username, email, isActive)
		fmt.Printf("   hash preview: %s...\n", hashPreview)
	}

	if count == 0 {
		fmt.Println("❌ لا يوجد مستخدمين! يجب إضافة مستخدم superadmin")
		return
	}

	// Test password verification
	fmt.Println("\n===========================================")
	fmt.Println("اختبار كلمة المرور لمستخدم superadmin")
	fmt.Println("===========================================")

	var passwordHash string
	err = db.QueryRow(`
		SELECT password_hash FROM super_system_admins WHERE username = 'superadmin'
	`).Scan(&passwordHash)

	if err == sql.ErrNoRows {
		fmt.Println("❌ المستخدم superadmin غير موجود!")
		return
	}
	if err != nil {
		log.Fatalf("❌ خطأ: %v", err)
	}

	// Check if it's a placeholder
	if passwordHash == "$argon2id$v=19$m=131072,t=4,p=4$PLACEHOLDER$PLACEHOLDER" {
		fmt.Println("❌ كلمة المرور لا تزال PLACEHOLDER!")
		fmt.Println("يجب تحديثها بالأمر:")
		fmt.Println("UPDATE super_system_admins SET password_hash = '<valid_hash>' WHERE username = 'superadmin';")
		return
	}

	// Test password
	testPassword := "Admin@123456"
	valid, err := security.VerifyPassword(testPassword, passwordHash)
	if err != nil {
		fmt.Printf("❌ خطأ في التحقق من كلمة المرور: %v\n", err)
		return
	}

	if valid {
		fmt.Printf("✅ كلمة المرور '%s' صحيحة!\n", testPassword)
	} else {
		fmt.Printf("❌ كلمة المرور '%s' غير صحيحة!\n", testPassword)
		fmt.Println("يجب تحديث كلمة المرور في قاعدة البيانات")
	}

	fmt.Println("\n===========================================")
	fmt.Println("انتهى الفحص")
	fmt.Println("===========================================")
}
