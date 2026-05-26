package main

import (
	"backend/internal/config"
	"backend/internal/database"
	"backend/internal/handler"
	"backend/internal/helper"
	"backend/internal/middleware"
	"backend/internal/repository"
	"backend/internal/routes"
	"backend/internal/service"
	"errors"
	"log"

	"github.com/gofiber/fiber/v3"
)

func main() {
	// 1. Config
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	helper.SwaggerAppURL = cfg.AppURL

	// 2. Database + migrations + seeding
	db := database.Connect(cfg.DBDSN)
	database.RunMigrations(db)

	// 3. Repositories
	userRepo := repository.NewGormUserRepo(db)
	refreshTokenRepo := repository.NewGormRefreshTokenRepo(db)
	menuRepo := repository.NewGormMenuRepo(db)
	roleRepo := repository.NewGormRoleRepo(db)
	auditLogRepo := repository.NewGormAuditLogRepo(db)

	// 4. Services
	mailer := helper.NewMailer(
		cfg.SMTP.Host,
		cfg.SMTP.Port,
		cfg.SMTP.Username,
		cfg.SMTP.Password,
		cfg.SMTP.From,
	)
	authService := service.NewAuthService(
		userRepo, refreshTokenRepo,
		cfg.JWTPrivateKey, cfg.JWTExpiry, cfg.RefreshTokenExpiry,
		mailer,
		cfg.AppURL,
	)
	userService := service.NewUserService(userRepo)
	menuService := service.NewMenuService(menuRepo, roleRepo)
	auditLogService := service.NewAuditLogService(auditLogRepo)

	// 5. Handlers
	handlers := handler.NewHandlers(
		handler.NewAuthHandler(authService, auditLogService),
		handler.NewUserHandler(userService, auditLogService),
		handler.NewMenuHandler(menuService, auditLogService),
	)

	// 6. Fiber
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			var e *fiber.Error
			if errors.As(err, &e) {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{"error": err.Error()})
		},
	})

	// 7. Global middleware
	app.Use(middleware.Logger())
	app.Use(middleware.Recover())
	app.Use(middleware.CORS())

	// 8. Routes
	routes.Register(app, handlers, cfg.JWTPublicKey)

	// 9. Start
	log.Printf("🚀 Server running on :%s", cfg.Port)
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
