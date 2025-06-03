package server

import (
	"auth/internal/config"
	"auth/internal/handler"
	"auth/internal/middleware"
	"auth/internal/repository"
	"auth/internal/service"
	"auth/internal/service/email"
	"auth/pkg/database"

	_ "auth/docs"

	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/swagger"
	"github.com/jackc/pgx/v4/pgxpool"
)

const (
	ApiPrefix  = "/api"
	ApiVersion = "/v1"
)

type Server struct {
	App    *fiber.App
	DbPool *pgxpool.Pool
}

// @title Auth API
// @version 1.0
// @description 인증/회원관리 서비스 API 문서
// @host localhost:3000
// @BasePath /api/v1
// @schemes http
// @contact.name Auth API Support
// @contact.email support@example.com
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

	if err := database.Connect(cfg.DatabaseUrl); err != nil {
		panic(err)
	}
	dbPool := database.GetPool()

	jwtService := service.NewJwtService(cfg.JwtSecret)
	emailService := email.NewEmailService(cfg.SmtpServer, cfg.SmtpPort, cfg.SmtpId, cfg.SmtpPassword)
	authService := service.NewAuthService(dbPool,
		repository.NewUserRepository(dbPool),
		repository.NewProfileRepository(dbPool),
		jwtService, emailService,
	)
	authHandler := handler.NewAuthHandler(authService)

	api := app.Group(ApiPrefix).Group(ApiVersion)
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

func (s *Server) Close() {
	if s.DbPool != nil {
		s.DbPool.Close()
	}
}
