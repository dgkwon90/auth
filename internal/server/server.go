// Package server provides the HTTP server and routing for the authentication service.
package server

// @title Auth API
// @version 1.0
// @description 인증/회원관리 서비스 API 문서
// @host localhost:3000
// @BasePath /api/v1
// @schemes http
// @contact.name Auth API Support
// @contact.email support@example.com

import (
	"auth/internal/config"
	"auth/internal/handler"
	"auth/internal/middleware"
	"auth/internal/repository"
	"auth/internal/service"
	"auth/internal/service/email"
	"auth/pkg/database"

	// docs 패키지는 Swagger 문서 생성을 위해 필요합니다. 실제 코드에서는 사용되지 않습니다.
	_ "auth/docs"

	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/swagger"
	"github.com/jackc/pgx/v4/pgxpool"
)

// APIPrefix is the base path for all API routes.
// APIVersion is the version path for the API.
const (
	APIPrefix  = "/api"
	APIVersion = "/v1"
)

// Server wraps the Fiber app and database pool.
type Server struct {
	App    *fiber.App
	DbPool *pgxpool.Pool
}

// NewServer creates and configures a new HTTP server for the authentication service.
func NewServer(cfg config.Config) *Server {
	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		JSONEncoder: sonic.Marshal,
		JSONDecoder: sonic.Unmarshal,
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		},
	})

	app.Use(logger.New())
	app.Use(cors.New())

	if err := database.Connect(cfg.DatabaseURL); err != nil {
		panic(err)
	}
	dbPool := database.GetPool()

	jwtService := service.NewJwtService(cfg.JwtSecret)
	emailService := email.NewEmailService(cfg.SMTPServer, cfg.SMTPPort, cfg.SMTPID, cfg.SMTPPassword)
	authService := service.NewAuthService(dbPool,
		repository.NewUserRepository(dbPool),
		repository.NewProfileRepository(dbPool),
		jwtService, emailService,
	)
	authHandler := handler.NewAuthHandler(authService)

	api := app.Group(APIPrefix).Group(APIVersion)
	auth := api.Group("/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)
	auth.Post("/refresh-token", authHandler.RefreshToken)
	auth.Post("/email/recover", authHandler.FindEmail)
	auth.Post("/password/forgot", authHandler.ForgotPassword)
	auth.Post("/password/reset", authHandler.ResetPassword)
	auth.Post("/logout", middleware.JwtMiddleware(jwtService), authHandler.Logout)

	users := api.Group("/users")
	users.Use(middleware.JwtMiddleware(jwtService))
	users.Get("/me", authHandler.GetProfile)
	users.Put("/me", authHandler.UpdateProfile)
	users.Delete("/me", authHandler.DeleteProfile)
	users.Put("/me/password", authHandler.ChangePassword)

	app.Get("/swagger/*", swagger.HandlerDefault)

	return &Server{App: app, DbPool: dbPool}
}

// Close gracefully closes the database connection pool.
func (s *Server) Close() {
	if s.DbPool != nil {
		s.DbPool.Close()
	}
}
