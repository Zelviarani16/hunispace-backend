package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/jackc/pgx/v5/pgxpool"

	"hunispace/app/model"
	"hunispace/app/repository"
	"hunispace/app/service"
	"hunispace/config"
	"hunispace/database"
	_ "hunispace/docs"
	"hunispace/helper"
	"hunispace/middleware"
	"hunispace/route"
)

func main() {
	config.LoadEnv()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	secret := config.GetEnv("JWT_SECRET", "")
	if len(secret) < 32 {
		log.Fatal("JWT_SECRET wajib diisi di .env (minimal 32 karakter)")
	}

	pool, err := database.NewPool(ctx)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()

	if config.GetEnv("AUTO_MIGRATE", "true") == "true" {
		if err := database.Migrate(ctx, pool, config.GetEnv("MIGRATIONS_DIR", "")); err != nil {
			log.Fatalf("migration gagal: %v", err)
		}
		log.Println("migration selesai")
	}

	jwtManager := helper.NewJWTManager(
		secret,
		config.GetEnv("JWT_ISSUER", "hunispace"),
		time.Duration(config.GetEnvInt("JWT_ACCESS_TTL_MINUTES", 60))*time.Minute,
	)

	seedOwner(ctx, pool, jwtManager)

	app := fiber.New(fiber.Config{AppName: "HuniSpace"})
	app.Use(middleware.ErrorResponse())
	app.Use(recover.New())
	app.Use(logger.New())

	route.Setup(app, route.Deps{
		Pool:   pool,
		Tx:     repository.NewTxManager(pool),
		JWT:    jwtManager,
		Auth:   middleware.RequireAuth(jwtManager),
		Owner:  middleware.RequireRole(model.RoleOwner),
		Tenant: middleware.RequireRole(model.RoleTenant),
	})

	port := config.GetEnv("APP_PORT", "3000")
	go func() {
		if err := app.Listen(":" + port); err != nil {
			log.Printf("server berhenti: %v", err)
			stop()
		}
	}()
	log.Printf("HuniSpace berjalan di http://localhost:%s", port)

	<-ctx.Done()
	log.Println("sinyal berhenti diterima, menutup server")
	if err := app.ShutdownWithTimeout(10 * time.Second); err != nil {
		log.Printf("shutdown: %v", err)
	}
}

func seedOwner(ctx context.Context, pool *pgxpool.Pool, jwtManager *helper.JWTManager) {
	email := config.GetEnv("SEED_OWNER_EMAIL", "")
	password := config.GetEnv("SEED_OWNER_PASSWORD", "")
	if email == "" || password == "" {
		return
	}
	users := service.NewUserService(pool, repository.NewUserRepository(), jwtManager)
	if err := users.EnsureOwner(ctx, config.GetEnv("SEED_OWNER_NAME", "Pemilik Kos"), email, password); err != nil {
		log.Fatalf("gagal membuat akun owner awal: %v", err)
	}
	log.Printf("akun owner awal siap: %s", email)
}
