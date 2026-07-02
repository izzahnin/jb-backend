package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	_ "github.com/izzahnin/jalur-berlian-backend/docs"
	"github.com/izzahnin/jalur-berlian-backend/internal/handler"
	"github.com/izzahnin/jalur-berlian-backend/internal/repository"
	"github.com/izzahnin/jalur-berlian-backend/internal/usecase"
	"github.com/izzahnin/jalur-berlian-backend/pkg/database"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
)

// @title Jalur Berlian Fleet Management API
// @version 1.0
// @description Real-time fleet management and order tracking system for PT. Jalur Berlian Makassar
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.jalurberlian.id/support
// @contact.email support@jalurberlian.id

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

// @schemes http https

// Config holds application configuration
type Config struct {
	DBSource      string
	RedisAddr     string
	RedisPassword string
	JWTSecret     string
	Port          string
	GinMode       string
	CORSOrigins   []string
}

// Application holds all dependencies
type Application struct {
	Config    *Config
	DB        *sqlx.DB
	Redis     *database.RedisClient
	Router    *gin.Engine
	Handler   *handler.Handler
}

func init() {
	// Reject unknown JSON keys so PATCH requests with misspelled fields do not silently no-op.
	binding.EnableDecoderDisallowUnknownFields = true
}

// loadEnvFiles loads environment variables from .env files
// Priority: .env.local (secrets) > .env (template)
// Useful for: go run, local development, testing
// Note: Docker-compose loads this automatically
func loadEnvFiles() {
	// Try .env.local first (highest priority - secrets)
	if err := godotenv.Load(".env.local"); err == nil {
		fmt.Println("📄 Loaded .env.local")
	}

	// Then try .env (fallback - template)
	if err := godotenv.Load(".env"); err == nil {
		fmt.Println("📄 Loaded .env")
	}

	// If both fail, that's okay - env vars might be set by system/docker
}

// LoadConfig loads environment variables with validation
func LoadConfig() (*Config, error) {
	// Parse CORS origins from comma-separated env var
	corsOrigins := []string{"http://localhost:3000", "http://localhost:3001"}
	if envOrigins := os.Getenv("CORS_ALLOWED_ORIGINS"); envOrigins != "" {
		corsOrigins = strings.Split(envOrigins, ",")
		for i := range corsOrigins {
			corsOrigins[i] = strings.TrimSpace(corsOrigins[i])
		}
	}

	cfg := &Config{
		DBSource:      os.Getenv("DB_SOURCE"),
		RedisAddr:     os.Getenv("REDIS_ADDR"),
		RedisPassword: os.Getenv("REDIS_PASSWORD"),
		JWTSecret:     os.Getenv("JWT_SECRET"),
		Port:          os.Getenv("PORT"),
		GinMode:       os.Getenv("GIN_MODE"),
		CORSOrigins:   corsOrigins,
	}

	// Validate required configuration
	if cfg.DBSource == "" {
		return nil, fmt.Errorf("missing required config: DB_SOURCE not set in environment")
	}
	if cfg.RedisAddr == "" {
		return nil, fmt.Errorf("missing required config: REDIS_ADDR not set in environment")
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("missing required config: JWT_SECRET not set in environment")
	}

	// Set defaults for optional config
	if cfg.Port == "" {
		cfg.Port = "8080"
	}
	if cfg.GinMode == "" {
		cfg.GinMode = "release"
	}

	return cfg, nil
}

// InitializeApplication initializes all infrastructure dependencies
func InitializeApplication(ctx context.Context, cfg *Config) (*Application, error) {
	app := &Application{Config: cfg}

	// Initialize PostgreSQL
	fmt.Println("🔧 Initializing PostgreSQL connection...")
	db, err := database.NewPostgres(cfg.DBSource)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize PostgreSQL: %w", err)
	}
	
	// Verify database connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("database health check failed: %w", err)
	}
	fmt.Println("✅ PostgreSQL connected")
	app.DB = db

	// Initialize Redis
	fmt.Println("🔧 Initializing Redis connection...")
	redis := database.NewRedis(cfg.RedisAddr, cfg.RedisPassword, 0)
	
	// Verify Redis connection
	if err := redis.Ping(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("redis health check failed: %w", err)
	}
	fmt.Println("✅ Redis connected")
	app.Redis = redis

	// Initialize Repositories (Data Layer)
	fmt.Println("🔧 Initializing repositories...")
	truckRepo := repository.NewTruckRepository(app.DB)
	orderRepo := repository.NewOrderRepository(app.DB)
	tripRepo := repository.NewTripRepository(app.DB)
	driverRepo := repository.NewDriverRepository(app.DB)
	customerRepo := repository.NewCustomerRepository(app.DB)
	auditRepo := repository.NewAuditLogRepository(app.DB)
	locRepo := repository.NewPostgresLocationRepo(app.DB)
	userRepo := repository.NewUserRepository(app.DB)
	fmt.Println("✅ Repositories initialized")

	// Initialize Usecases (Business Logic Layer)
	fmt.Println("🔧 Initializing usecases...")
	truckUsecase := usecase.NewTruckUsecase(truckRepo, tripRepo)
	orderUsecase := usecase.NewOrderUsecase(orderRepo, customerRepo)
	tripUsecase := usecase.NewTripUsecase(tripRepo, orderRepo, truckRepo, driverRepo, auditRepo)
	locationUsecase := usecase.NewLocationUsecase(locRepo, tripRepo, redis)
	customerUsecase := usecase.NewCustomerUsecase(customerRepo)
	driverUsecase := usecase.NewDriverUsecase(driverRepo, tripRepo)
	authUsecase := usecase.NewAuthUsecase(userRepo, cfg.JWTSecret)
	userUsecase := usecase.NewUserUsecase(userRepo)
	fmt.Println("✅ Usecases initialized")

	// Initialize Router (Presentation Layer)
	fmt.Println("🔧 Initializing router...")
	gin.SetMode(cfg.GinMode)
	router := gin.Default()

	// Setup CORS middleware
	router.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORSOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * 3600,
	}))

	// Initialize Handler with all dependencies
	h := handler.NewHandler(
		truckRepo, orderRepo, tripRepo, driverRepo, customerRepo, auditRepo, locRepo, userRepo,
		truckUsecase, orderUsecase, tripUsecase, locationUsecase, customerUsecase, driverUsecase, authUsecase, userUsecase,
		cfg.JWTSecret, redis,
	)

	// Register all routes
	h.RegisterAllRoutes(router)

	// Setup Swagger endpoints
	setupSwaggerEndpoints(router)

	fmt.Println("✅ Router configured")

	app.Router = router
	app.Handler = h

	return app, nil
}

// Cleanup gracefully closes all resources
func (app *Application) Cleanup() error {
	fmt.Println("\n🧹 Cleaning up resources...")
	
	if app.Redis != nil {
		if err := app.Redis.Close(); err != nil {
			fmt.Printf("⚠️  Warning: Redis cleanup error: %v\n", err)
		}
	}
	
	if app.DB != nil {
		if err := app.DB.Close(); err != nil {
			fmt.Printf("⚠️  Warning: Database cleanup error: %v\n", err)
		}
	}
	
	fmt.Println("✅ Cleanup completed")
	return nil
}

// setupSwaggerEndpoints configures Swagger documentation endpoints
func setupSwaggerEndpoints(router *gin.Engine) {
	router.GET("/swagger/swagger.json", func(c *gin.Context) {
		c.File("./docs/swagger.json")
	})

	router.GET("/swagger/docs", func(c *gin.Context) {
		html := `<!DOCTYPE html>
<html lang="id">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Jalur Berlian Fleet Management API</title>
    <link rel="stylesheet" type="text/css" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@4/swagger-ui.css">
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@4/swagger-ui-bundle.js"></script>
    <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@4/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = function() {
            SwaggerUIBundle({
                url: "/swagger/swagger.json",
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIBundle.SwaggerUIStandalonePreset
                ],
                plugins: [
                    SwaggerUIBundle.plugins.DownloadUrl
                ],
                layout: "BaseLayout",
                requestInterceptor: (request) => {
                    if (request.headers.Authorization && !request.headers.Authorization.startsWith('Bearer ')) {
                        request.headers.Authorization = 'Bearer ' + request.headers.Authorization;
                    }
                    return request;
                }
            })
        }
    </script>
</body>
</html>`
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(200, html)
	})
}

func main() {
	separator := strings.Repeat("=", 60)
	fmt.Println("\n" + separator)
	fmt.Println("🚀 PT. Jalur Berlian Fleet Management API")
	fmt.Println(separator)

	// Load environment variables from .env files (for local development)
	fmt.Println("\n📋 Loading environment files...")
	loadEnvFiles()

	// Load configuration
	fmt.Println("📋 Loading configuration...")
	cfg, err := LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Configuration error: %v\n", err)
		fmt.Fprintf(os.Stderr, "\n📝 Required environment variables:\n")
		fmt.Fprintf(os.Stderr, "  - DB_SOURCE: postgres://user:password@host:port/dbname?sslmode=disable\n")
		fmt.Fprintf(os.Stderr, "  - REDIS_ADDR: host:port (e.g., localhost:6379)\n")
		fmt.Fprintf(os.Stderr, "  - JWT_SECRET: (generate with: openssl rand -hex 32)\n")
		os.Exit(1)
	}
	fmt.Println("✅ Configuration loaded")

	// Initialize application
	fmt.Println("\n⚙️  Initializing application...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	app, err := InitializeApplication(ctx, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Initialization error: %v\n", err)
		os.Exit(1)
	}
	defer app.Cleanup()

	fmt.Println("\n✅ Application initialized successfully")

	// Start HTTP server
	fmt.Printf("\n🌐 Starting server on http://localhost:%s\n", cfg.Port)
	fmt.Println("📚 Swagger UI: http://localhost:" + cfg.Port + "/swagger/docs")
	fmt.Println(separator + "\n")

	// Create HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      app.Router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "❌ Server error: %v\n", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown: Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	fmt.Printf("\n🛑 Received signal: %v\n", sig)

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	fmt.Println("\n⏳ Shutting down server gracefully...")
	if err := srv.Shutdown(shutdownCtx); err != nil {
		fmt.Fprintf(os.Stderr, "⚠️  Shutdown error: %v\n", err)
	}

	fmt.Println("✅ Server stopped gracefully")
	fmt.Println(separator)
}