package main

import (
	"cso-erp/internal/config"
	"cso-erp/internal/database"
	"cso-erp/internal/handlers"
	"cso-erp/internal/middleware"
	"cso-erp/internal/models"
	"cso-erp/internal/services"
	"cso-erp/internal/swagger"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	cfg := config.Load()

	db, err := database.Connect(&cfg.Database)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Connect to Master Database for SSA operations
	masterDBConfig := &config.DatabaseConfig{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		DBName:   cfg.MultiTenant.MasterDBName, // erp_master
		SSLMode:  cfg.Database.SSLMode,
	}

	masterDB, err := database.Connect(masterDBConfig)
	if err != nil {
		log.Printf("⚠️ Failed to connect to master database (%s): %v", cfg.MultiTenant.MasterDBName, err)
		log.Println("SSA features will be disabled")
		masterDB = db // Fallback to main DB for development
	} else {
		log.Printf("✅ Master database connected: %s", cfg.MultiTenant.MasterDBName)
		defer masterDB.Close()
	}

	var emailConfig *services.EmailConfig
	if cfg.Email.SMTPHost != "" {
		emailConfig = &services.EmailConfig{
			SMTPHost:    cfg.Email.SMTPHost,
			SMTPPort:    cfg.Email.SMTPPort,
			Username:    cfg.Email.Username,
			Password:    cfg.Email.Password,
			FromAddress: cfg.Email.FromAddress,
			FromName:    cfg.Email.FromName,
			UseTLS:      cfg.Email.UseTLS,
			SkipVerify:  cfg.Email.SkipVerify,
		}
	}
	emailService := services.NewEmailService(emailConfig)

	// Initialize OTP cleanup scheduler for expired codes
	otpCleanupService := services.NewOTPCleanupService(db.DB)
	otpCleanupService.StartCleanupScheduler()
	defer otpCleanupService.StopCleanupScheduler()

	// authService is now created per-request in AuthHandler
	// authService := services.NewAuthService(db, cfg, emailService)

	// Keep userService for EmailHandler (until EmailHandler is refactored)
	userService := services.NewUserService(db, cfg)

	containerService := services.NewContainerService(db)
	transactionService := services.NewTransactionService(db)
	reportService := services.NewReportService(db)
	projectService := services.NewProjectService(db)
	employeeService := services.NewEmployeeService(db)
	memberService := services.NewMemberService(db)
	dashboardService := services.NewDashboardService(db)
	transferService := services.NewTransferService(db)
	receiptService := services.NewReceiptService(db)
	hrService := services.NewHRService(db)

	// Phase 3: Features Services
	// File Service
	fileService, err := services.NewFileService(db, "./uploads")
	if err != nil {
		log.Printf("⚠️ Failed to initialize FileService: %v", err)
	}

	activityService := services.NewActivityService(db)
	backupService := services.NewBackupService(db, "./backups")

	// Notification Service with background expiry checker
	notificationService := services.NewNotificationService(db, emailService)
	notificationService.StartExpiryChecker(24 * time.Hour) // Check daily
	defer notificationService.StopExpiryChecker()

	authHandler := handlers.NewAuthHandler(masterDB, cfg, emailService)
	userHandler := handlers.NewUserHandler(cfg, emailService)
	emailTemplateHandler := handlers.NewEmailTemplateHandler(cfg)

	containerHandler := handlers.NewContainerHandler(containerService)
	transactionHandler := handlers.NewTransactionHandler(transactionService)
	reportHandler := handlers.NewReportHandler(reportService)
	projectHandler := handlers.NewProjectHandler(projectService)
	employeeHandler := handlers.NewEmployeeHandler(employeeService)
	memberHandler := handlers.NewMemberHandler(memberService)
	dashboardHandler := handlers.NewDashboardHandler(dashboardService)
	transferHandler := handlers.NewTransferHandler(transferService)
	receiptHandler := handlers.NewReceiptHandler(receiptService)
	hrHandler := handlers.NewHRHandler(hrService)
	emailHandler := handlers.NewEmailHandler(emailService, userService, employeeService, memberService)

	// Feature Handlers
	attachmentHandler := handlers.NewAttachmentHandler(fileService)
	activityHandler := handlers.NewActivityHandler(activityService)
	backupHandler := handlers.NewBackupHandler(backupService)
	notificationHandler := handlers.NewNotificationHandler(notificationService)

	// Commercial Handlers
	invoiceService := services.NewInvoiceService(db)
	invoiceHandler := handlers.NewInvoiceHandler(invoiceService)
	legalHandler := handlers.NewLegalHandler()

	app := fiber.New(fiber.Config{
		AppName:      "CSO ERP System",
		ErrorHandler: customErrorHandler,
	})

	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${method} ${path} (${latency})\n",
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORS.Origin,
		AllowCredentials: cfg.CORS.Credentials,
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS, PATCH",
	}))

	// Tenant Resolver - Identifies tenant and injects DB
	// Uses masterDB because organizations table is in erp_master
	tenantResolver := middleware.NewTenantResolver(cfg, masterDB.DB)
	app.Use(tenantResolver.ResolveTenant())

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "healthy",
			"service": "CSO ERP API",
		})
	})

	// Setup Swagger API documentation
	swagger.SetupSwaggerRoutes(app, swagger.Config{
		Title:       "CSO-ERP API",
		Version:     "1.0.0",
		Description: "Multi-Tenant SaaS ERP Platform API Documentation",
		BasePath:    "/api/v1",
		Host:        cfg.App.Host + ":" + cfg.App.Port,
	})

	api := app.Group("/api/v1")

	auth := api.Group("/auth")
	auth.Post("/login", middleware.RateLimitMiddleware(&cfg.Security), authHandler.Login)
	auth.Post("/logout", authHandler.Logout)
	auth.Post("/register", middleware.RateLimitMiddleware(&cfg.Security), authHandler.Register)
	auth.Post("/verify-2fa", middleware.RateLimitMiddleware(&cfg.Security), authHandler.Verify2FA)
	auth.Post("/resend-otp", middleware.RateLimitMiddleware(&cfg.Security), authHandler.ResendOTP)

	protected := api.Group("", middleware.AuthMiddleware(&cfg.JWT))

	// Attachments Routes
	attachments := protected.Group("/attachments")
	attachments.Post("/upload", attachmentHandler.UploadAttachment)
	attachments.Get("", attachmentHandler.GetAttachments)
	attachments.Get("/:id/download", attachmentHandler.DownloadAttachment)
	attachments.Delete("/:id", attachmentHandler.DeleteAttachment)

	// Activity Logs Routes
	activities := protected.Group("/activities", middleware.RequireRole(models.RoleAdmin))
	activities.Get("/entity", activityHandler.GetEntityActivities)
	activities.Get("/user", activityHandler.GetUserActivities)

	// Backup Routes
	backups := protected.Group("/backups", middleware.RequireRole(models.RoleAdmin))
	backups.Post("", backupHandler.CreateBackup)
	backups.Get("", backupHandler.ListBackups)

	// Notification Routes
	notifications := protected.Group("/notifications")
	notifications.Get("", notificationHandler.GetNotifications)
	notifications.Get("/unread-count", notificationHandler.GetUnreadCount)
	notifications.Post("/:id/read", notificationHandler.MarkAsRead)
	notifications.Post("/read-all", notificationHandler.MarkAllAsRead)

	// Invoice Routes
	invoices := protected.Group("/invoices")
	invoices.Get("", invoiceHandler.ListInvoices)
	invoices.Get("/:id", invoiceHandler.GetInvoice)
	invoices.Get("/:id/print", invoiceHandler.PrintInvoice)
	invoices.Post("/from-transaction/:transaction_id", invoiceHandler.GenerateFromTransaction)

	// Legal Pages (Public)
	legal := api.Group("/legal")
	legal.Get("/privacy-policy", legalHandler.GetPrivacyPolicy)
	legal.Get("/terms-of-service", legalHandler.GetTermsOfService)

	protectedAuth := protected.Group("/auth")
	protectedAuth.Post("/revoke-all-sessions", authHandler.RevokeAllSessions)
	protectedAuth.Post("/change-password", authHandler.ChangePassword)
	protectedAuth.Post("/toggle-2fa", authHandler.Toggle2FA)

	users := protected.Group("/users")
	users.Get("/me", authHandler.GetCurrentUser)

	adminUsers := protected.Group("/users", middleware.RequireRole(models.RoleAdmin))
	adminUsers.Get("", userHandler.ListUsers)
	adminUsers.Post("", userHandler.CreateUser)
	adminUsers.Get("/:id", userHandler.GetUser)
	adminUsers.Put("/:id", userHandler.UpdateUser)
	adminUsers.Delete("/:id", userHandler.DeleteUser)
	adminUsers.Patch("/:id/role", userHandler.UpdateUserRole)

	// Organization Email Templates & SMTP Settings
	org := protected.Group("/organization", middleware.RequireRole(models.RoleAdmin))
	org.Get("/email-templates", emailTemplateHandler.ListTemplates)
	org.Get("/email-templates/:id", emailTemplateHandler.GetTemplate)
	org.Put("/email-templates/:id", emailTemplateHandler.UpdateTemplate)
	org.Post("/email-templates/preview", emailTemplateHandler.PreviewTemplate)
	org.Post("/email-templates/:key/reset", emailTemplateHandler.ResetTemplate)

	org.Get("/smtp-settings", emailTemplateHandler.ListSMTPSettings)
	org.Post("/smtp-settings", emailTemplateHandler.CreateSMTPSettings)
	org.Get("/smtp-settings/:id", emailTemplateHandler.GetSMTPSettings)
	org.Put("/smtp-settings/:id", emailTemplateHandler.UpdateSMTPSettings)
	org.Delete("/smtp-settings/:id", emailTemplateHandler.DeleteSMTPSettings)
	org.Post("/smtp-settings/:id/test", emailTemplateHandler.TestSMTPSettings)

	dashboardWidgets := protected.Group("/admin/dashboard-widgets", middleware.RequireRole(models.RoleAdmin))
	dashboardWidgets.Get("", dashboardHandler.GetWidgetConfigs)
	dashboardWidgets.Put("/:id", dashboardHandler.UpdateWidgetConfig)
	dashboardWidgets.Post("/bulk-order", dashboardHandler.BulkUpdateWidgetOrder)

	containers := protected.Group("/containers")
	containers.Get("", middleware.RequireFinanceOrAdmin(), containerHandler.ListContainers)
	containers.Post("", middleware.RequireRole(models.RoleAdmin), containerHandler.CreateContainer)
	containers.Get("/:id", middleware.RequireRole(models.RoleAdmin), containerHandler.GetContainer)
	containers.Put("/:id", middleware.RequireRole(models.RoleAdmin), containerHandler.UpdateContainer)
	containers.Delete("/:id", middleware.RequireRole(models.RoleAdmin), containerHandler.DeleteContainer)

	transactions := protected.Group("/transactions")
	transactions.Post("/", middleware.RequireFinanceOrAdmin(), transactionHandler.CreateTransaction)
	transactions.Get("/:id", middleware.RequireAnyAuthenticated(), transactionHandler.GetTransaction)
	transactions.Post("/:id/reverse", middleware.RequireFinanceOrAdmin(), transactionHandler.ReverseTransaction)

	transfers := protected.Group("/transfers")
	transfers.Post("/", middleware.RequireFinanceOrAdmin(), transferHandler.CreateTransfer)

	receipts := protected.Group("/receipts")
	receipts.Get("/:id", middleware.RequireAnyAuthenticated(), receiptHandler.GetReceiptData)

	reports := protected.Group("/reports")
	reports.Get("/dashboard", middleware.RequireAnyAuthenticated(), reportHandler.GetDashboardSummary)
	reports.Get("/containers/:id/statement", middleware.RequireAnyAuthenticated(), reportHandler.GetContainerStatement)
	reports.Get("/projects/:id/statement", middleware.RequireAnyAuthenticated(), reportHandler.GetProjectStatement)
	reports.Get("/employees/:id/statement", middleware.RequireFinanceOrAdmin(), reportHandler.GetEmployeeStatement)
	reports.Get("/members/:id/statement", middleware.RequireFinanceOrAdmin(), reportHandler.GetMemberStatement)
	reports.Get("/daily-position", middleware.RequireAnyAuthenticated(), reportHandler.GetDailyPosition)
	reports.Get("/project-summary", middleware.RequireAnyAuthenticated(), reportHandler.GetProjectSummary)
	reports.Get("/invoices", middleware.RequireAnyAuthenticated(), reportHandler.GetInvoices)

	projects := protected.Group("/projects", middleware.RequireRole(models.RoleAdmin))
	projects.Get("", projectHandler.ListProjects)
	projects.Post("", projectHandler.CreateProject)
	projects.Get("/:id", projectHandler.GetProject)
	projects.Put("/:id", projectHandler.UpdateProject)
	projects.Delete("/:id", projectHandler.DeleteProject)

	employees := protected.Group("/employees", middleware.RequireRole(models.RoleAdmin, models.RoleHR, models.RoleFinance))
	employees.Get("", employeeHandler.ListEmployees)
	employees.Post("", employeeHandler.CreateEmployee)
	employees.Get("/:id", employeeHandler.GetEmployee)
	employees.Put("/:id", employeeHandler.UpdateEmployee)
	employees.Delete("/:id", employeeHandler.DeleteEmployee)

	members := protected.Group("/members", middleware.RequireRole(models.RoleAdmin))
	members.Get("", memberHandler.ListMembers)
	members.Post("", memberHandler.CreateMember)
	members.Get("/expiring", memberHandler.GetExpiringMembers)
	members.Get("/:id", memberHandler.GetMember)
	members.Put("/:id", memberHandler.UpdateMember)
	members.Delete("/:id", memberHandler.DeleteMember)
	members.Post("/:id/renew", memberHandler.RenewMembership)
	members.Get("/:id/subscriptions", memberHandler.GetSubscriptions)

	hr := protected.Group("/hr", middleware.RequireRole(models.RoleHR, models.RoleAdmin, models.RoleFinance))
	hr.Get("/salaries", hrHandler.GetMonthlySalaries)
	hr.Post("/deductions", hrHandler.CreateDeduction)
	hr.Get("/deductions", hrHandler.ListDeductions)
	hr.Post("/bonuses", hrHandler.CreateBonus)
	hr.Get("/bonuses", hrHandler.ListBonuses)
	hr.Get("/employees/profile", hrHandler.GetEmployeeFinancialProfile)

	// Email routes (Admin only)
	email := protected.Group("/email", middleware.RequireRole(models.RoleAdmin))
	email.Post("/bulk", emailHandler.SendBulkEmail)

	// ==========================================================================
	// SUPER SYSTEM ADMIN (SSA) ROUTES - Multi-Tenant Management
	// ==========================================================================
	// Note: These routes require RoleSystemAdmin and additional SSA middleware
	// The SSA middleware includes: IP restriction, rate limiting, and 2FA check

	// Initialize SSA Auth Service first (needed by TenantHandler)
	ssaAuthService := services.NewSSAAuthService(masterDB, cfg, emailService)
	ssaAuthHandler := handlers.NewSSAAuthHandler(ssaAuthService, cfg)

	// Initialize Tenant Service and Handler for SSA operations (using MASTER DB)
	tenantService := services.NewTenantService(masterDB, cfg)
	tenantHandler := handlers.NewTenantHandler(tenantService, emailService, ssaAuthService, cfg)

	// SSA Auth Routes - No auth middleware (these are for initial login)
	// Registered directly on app to avoid ANY middleware from api group
	ssaAuth := app.Group("/api/v1/system-auth")
	ssaAuth.Post("/login", ssaAuthHandler.Login)
	ssaAuth.Post("/verify-2fa", ssaAuthHandler.Verify2FA)
	ssaAuth.Post("/logout", ssaAuthHandler.Logout)

	// SSA API Routes - Protected by Super Admin middleware
	ssa := api.Group("/ssa",
		middleware.AuthMiddleware(&cfg.JWT),
		middleware.RequireSuperAdmin(cfg),
	)

	// SSA Profile
	ssa.Get("/me", ssaAuthHandler.GetMe)

	// SSA Audit Logs
	ssa.Get("/audit-logs", ssaAuthHandler.GetAuditLogs)

	// Organization Management
	ssa.Get("/subscription-plans", tenantHandler.ListSubscriptionPlans)
	ssa.Post("/subscription-plans", tenantHandler.CreateSubscriptionPlan)
	ssa.Get("/subscription-plans/:id", tenantHandler.GetSubscriptionPlan)
	ssa.Put("/subscription-plans/:id", tenantHandler.UpdateSubscriptionPlan)
	ssa.Delete("/subscription-plans/:id", tenantHandler.DeleteSubscriptionPlan)

	ssa.Get("/stats", tenantHandler.GetOrganizationStats)
	ssa.Get("/organizations", tenantHandler.ListOrganizations)
	ssa.Post("/organizations", tenantHandler.CreateOrganization)
	ssa.Get("/organizations/:id", tenantHandler.GetOrganization)

	// Domain Verification
	ssa.Post("/organizations/:id/verify", tenantHandler.VerifyDomain)
	ssa.Post("/organizations/:id/resend-verification", tenantHandler.ResendVerification)

	// Subscription Management
	ssa.Put("/organizations/:id/subscription", tenantHandler.UpdateOrganizationSubscription)

	// Database Creation
	ssa.Post("/organizations/:id/database", tenantHandler.CreateDatabase)

	// Tenant Admin Management
	ssa.Post("/organizations/:id/admin", tenantHandler.CreateTenantAdmin)

	// Organization Status Management
	ssa.Post("/organizations/:id/suspend", tenantHandler.SuspendOrganization)
	ssa.Post("/organizations/:id/activate", tenantHandler.ActivateOrganization)

	// Organization Deletion (with OTP verification)
	ssa.Post("/organizations/:id/request-deletion", tenantHandler.RequestDeletion)
	ssa.Post("/organizations/:id/confirm-deletion", tenantHandler.ConfirmDeletion)

	log.Printf("🔐 SSA Routes configured at /api/v1/ssa/*")

	port := cfg.App.Port
	log.Printf("🚀 Server starting on port %s", port)
	log.Printf("📊 Environment: %s", cfg.App.Env)
	log.Printf("🔒 CORS Origin: %s", cfg.CORS.Origin)

	go func() {
		if err := app.Listen(":" + port); err != nil {
			log.Fatalf("❌ Server failed to start: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down server...")
	if err := app.Shutdown(); err != nil {
		log.Fatalf("❌ Server forced to shutdown: %v", err)
	}

	log.Println("✅ Server exited properly")
	log.Println("Closing database connection...")
}

func customErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	return c.Status(code).JSON(fiber.Map{
		"error": err.Error(),
		"code":  code,
	})
}
