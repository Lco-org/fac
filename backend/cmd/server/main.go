package main

import (
	"cso-erp/internal/config"
	"cso-erp/internal/database"
	"cso-erp/internal/handlers"
	"cso-erp/internal/middleware"
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
	// Load configuration
	cfg := config.Load()

	// Connect to database
	db, err := database.Connect(&cfg.Database)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize services
	authService := services.NewAuthService(db, cfg)
	transactionService := services.NewTransactionService(db)
	reportService := services.NewReportService(db)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService, cfg)
	transactionHandler := handlers.NewTransactionHandler(transactionService)
	reportHandler := handlers.NewReportHandler(reportService)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName:      "CSO ERP System",
		ErrorHandler: customErrorHandler,
	})

	// Global middleware
	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${method} ${path} (${latency})\n",
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORS.Origin,
		AllowCredentials: cfg.CORS.Credentials,
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS",
	}))

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "healthy",
			"service": "CSO ERP API",
		})
	})

	// API routes
	api := app.Group("/api/v1")

	// Public routes
	auth := api.Group("/auth")
	auth.Post("/login", middleware.RateLimitMiddleware(&cfg.Security), authHandler.Login)
	auth.Post("/logout", authHandler.Logout)
	auth.Post("/register", middleware.RateLimitMiddleware(&cfg.Security), authHandler.Register)

	// Protected routes
	protected := api.Group("", middleware.AuthMiddleware(&cfg.JWT))

	// User routes
	users := protected.Group("/users")
	users.Get("/me", authHandler.GetCurrentUser)

	// Transaction routes (Finance & Admin can create, only Admin can delete)
	transactions := protected.Group("/transactions")
	transactions.Post("/", middleware.RequireFinanceOrAdmin(), transactionHandler.CreateTransaction)
	transactions.Get("/:id", middleware.RequireAnyAuthenticated(), transactionHandler.GetTransaction)
	transactions.Post("/:id/reverse", middleware.RequireFinanceOrAdmin(), transactionHandler.ReverseTransaction)

	// Report routes (All authenticated users can view)
	reports := protected.Group("/reports")
	reports.Get("/daily-position", middleware.RequireAnyAuthenticated(), reportHandler.GetDailyPosition)
	reports.Get("/project-summary", middleware.RequireAnyAuthenticated(), reportHandler.GetProjectSummary)
	reports.Get("/deficit", middleware.RequireAnyAuthenticated(), reportHandler.GetDeficitReport)

	// Start server
	port := cfg.App.Port
	log.Printf("🚀 Server starting on port %s", port)
	log.Printf("📊 Environment: %s", cfg.App.Env)
	log.Printf("🔒 CORS Origin: %s", cfg.CORS.Origin)

	// Graceful shutdown
	go func() {
		if err := app.Listen(":" + port); err != nil {
			log.Fatalf("❌ Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal
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
