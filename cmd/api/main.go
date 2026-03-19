package main

import (
	"context"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/izzahnin/jalur-berlian-backend/docs"
	"github.com/izzahnin/jalur-berlian-backend/internal/handler"
	"github.com/izzahnin/jalur-berlian-backend/internal/repository"
	"github.com/izzahnin/jalur-berlian-backend/internal/usecase"
	"github.com/izzahnin/jalur-berlian-backend/pkg/database"
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
func main() {
	// ---------------------------
	// Setup Infrastruktur
	// ---------------------------
	// PostgreSQL connection - read from environment (REQUIRED, no fallback for secrets!)
	dbSource := os.Getenv("DB_SOURCE")
	if dbSource == "" {
		log.Fatal("❌ ERROR: DB_SOURCE environment variable is not set!\n" +
			"Please set it in .env.local or environment:\n" +
			"  DB_SOURCE=postgres://user:password@host:5432/dbname?sslmode=disable\n" +
			"\nExample:\n" +
			"  DB_SOURCE=postgres://admin:mypassword@localhost:5432/jalur_berlian_db?sslmode=disable")
	}
	db, err := database.NewPostgres(dbSource)
	if err != nil {
		log.Fatalf("Gagal inisialisasi database: %v", err)
	}

	// Redis connection - read from environment (REQUIRED)
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		log.Fatal("❌ ERROR: REDIS_ADDR environment variable is not set!\n" +
			"Please set it in .env.local or environment:\n" +
			"  REDIS_ADDR=host:port (e.g., localhost:6379)")
	}
	redisClient := database.NewRedis(redisAddr, "", 0)
	defer redisClient.Close()
	if err := redisClient.Ping(context.Background()); err != nil {
		log.Fatalf("redis ping: %v", err)
	}

	// ---------------------------
	// Repositories (Data Layer)
	// ---------------------------
	truckRepo := repository.NewTruckRepository(db)
	orderRepo := repository.NewOrderRepository(db)
	locRepo := repository.NewRedisLocationRepo(redisClient)
	userRepo := repository.NewUserRepository(db)
	// auditLogRepo will be needed in STEP 3 (logging) when handlers log mutations
	// auditLogRepo := repository.NewAuditLogRepository(db)

	// ---------------------------
	// Usecases (Business Logic)
	// ---------------------------
	truckUsecase := usecase.NewTruckUsecase(truckRepo)
	orderUsecase := usecase.NewOrderUsecase(orderRepo, truckRepo)
	locationUsecase := usecase.NewLocationUsecase(locRepo)

	// JWT Secret dari environment variable (REQUIRED, no hardcoded secrets!)
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("❌ ERROR: JWT_SECRET environment variable is not set!\n" +
			"This secret MUST be set in .env.local or environment.\n" +
			"Generate a secure secret using:\n" +
			"  Windows: [BitConverter]::ToString([System.Security.Cryptography.RandomNumberGenerator]::GetBytes(32)) -replace '-','' | ForEach-Object {$_}\n" +
			"  Linux/Mac: openssl rand -hex 32\n" +
			"\nThen set it in .env.local:\n" +
			"  JWT_SECRET=your_generated_secret_here")
	}
	authUsecase := usecase.NewAuthUsecase(userRepo, jwtSecret)

	// ---------------------------
	// ---------------------------
	// Router Setup
	// ---------------------------
	r := gin.Default()

	// Create handler dengan dependency injection
	h := handler.NewHandler(
		truckRepo, orderRepo, locRepo, userRepo,
		truckUsecase, orderUsecase, locationUsecase, authUsecase,
		jwtSecret,
	)

	// Register semua routes (21 endpoints)
	h.RegisterAllRoutes(r)

	// Swagger Documentation endpoints
	// Serve swagger.json spec file
	r.GET("/swagger/swagger.json", func(c *gin.Context) {
		// This endpoint serves the Swagger/OpenAPI spec
		c.File("./docs/swagger.json")
	})

	// Serve Swagger UI HTML page
	r.GET("/swagger/docs", func(c *gin.Context) {
		// Swagger UI HTML page - standard layout
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
                    // Auto-add Bearer prefix to Authorization header
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

	// ---------------------------
	// Start Server
	// ---------------------------
	log.Println("Server PT. Jalur Berlian berjalan di port :8080")
	log.Println("📚 Swagger UI: http://localhost:8080/swagger/docs")
	r.Run(":8080")
}