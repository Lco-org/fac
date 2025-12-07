package main

import (
	"cso-erp/internal/config"
	"cso-erp/internal/database"
	"cso-erp/internal/handlers"
	"cso-erp/internal/middleware"
	"cso-erp/internal/models"
	"cso-erp/internal/services"
	"log"
	"os"
	"os/signal"
	"syscall"

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

	authService := services.NewAuthService(db, cfg, emailService)
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

	authHandler := handlers.NewAuthHandler(authService, cfg)
	userHandler := handlers.NewUserHandler(userService)
	containerHandler := handlers.NewContainerHandler(containerService)
	transactionHandler := handlers.NewTransactionHandler(transactionService)
	reportHandler := handlers.NewReportHandler(reportService)
	projectHandler := handlers.NewProjectHandler(projectService)
	employeeHandler := handlers.NewEmployeeHandler(employeeService)
	memberHandler := handlers.NewMemberHandler(memberService)
	dashboardHandler := handlers.NewDashboardHandler(dashboardService)
	transferHandler := handlers.NewTransferHandler(transferService)
	receiptHandler := handlers.NewReceiptHandler(receiptService)

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

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "healthy",
			"service": "CSO ERP API",
		})
	})

	api := app.Group("/api/v1")

	auth := api.Group("/auth")
	auth.Post("/login", middleware.RateLimitMiddleware(&cfg.Security), authHandler.Login)
	auth.Post("/logout", authHandler.Logout)
	auth.Post("/register", middleware.RateLimitMiddleware(&cfg.Security), authHandler.Register)
	auth.Post("/verify-2fa", middleware.RateLimitMiddleware(&cfg.Security), authHandler.Verify2FA)
	auth.Post("/resend-otp", middleware.RateLimitMiddleware(&cfg.Security), authHandler.ResendOTP)

	protected := api.Group("", middleware.AuthMiddleware(&cfg.JWT))

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

	dashboardWidgets := protected.Group("/admin/dashboard-widgets", middleware.RequireRole(models.RoleAdmin))
	dashboardWidgets.Get("", dashboardHandler.GetWidgetConfigs)
	dashboardWidgets.Put("/:id", dashboardHandler.UpdateWidgetConfig)
	dashboardWidgets.Post("/bulk-order", dashboardHandler.BulkUpdateWidgetOrder)

	containers := protected.Group("/containers", middleware.RequireRole(models.RoleAdmin))
	containers.Get("", containerHandler.ListContainers)
	containers.Post("", containerHandler.CreateContainer)
	containers.Get("/:id", containerHandler.GetContainer)
	containers.Put("/:id", containerHandler.UpdateContainer)
	containers.Delete("/:id", containerHandler.DeleteContainer)

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
	reports.Get("/daily-position", middleware.RequireAnyAuthenticated(), reportHandler.GetDailyPosition)
	reports.Get("/project-summary", middleware.RequireAnyAuthenticated(), reportHandler.GetProjectSummary)

	projects := protected.Group("/projects", middleware.RequireRole(models.RoleAdmin))
	projects.Get("", projectHandler.ListProjects)
	projects.Post("", projectHandler.CreateProject)
	projects.Get("/:id", projectHandler.GetProject)
	projects.Put("/:id", projectHandler.UpdateProject)
	projects.Delete("/:id", projectHandler.DeleteProject)

	employees := protected.Group("/employees", middleware.RequireRole(models.RoleAdmin))
	employees.Get("", employeeHandler.ListEmployees)
	employees.Post("", employeeHandler.CreateEmployee)
	employees.Get("/:id", employeeHandler.GetEmployee)
	employees.Put("/:id", employeeHandler.UpdateEmployee)
	employees.Delete("/:id", employeeHandler.DeleteEmployee)

	members := protected.Group("/members", middleware.RequireRole(models.RoleAdmin))
	members.Get("", memberHandler.ListMembers)
	members.Post("", memberHandler.CreateMember)
	members.Get("/:id", memberHandler.GetMember)
	members.Put("/:id", memberHandler.UpdateMember)
	members.Delete("/:id", memberHandler.DeleteMember)

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
