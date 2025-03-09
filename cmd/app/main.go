package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"manga-reader2/config"
	customMiddleware "manga-reader2/internal/api/middleware"
	"manga-reader2/internal/api/router"
	"manga-reader2/internal/common/logger"
	"manga-reader2/internal/infrastructure/auth"
	"manga-reader2/internal/infrastructure/db"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// @title          Manga Reader API
// @version        1.0
// @description    API для чтения манги

// @host     localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey Bearer
// @in                         header
// @name                       Authorization
// @description                Введите токен в формате: Bearer {token}
func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		fmt.Printf("Ошибка загрузки конфигурации: %v\n", err)
		os.Exit(1)
	}

	log := logger.NewLogger(cfg.Log.Level)
	log.Info("Запуск приложения...")

	ctx := context.Background()

	pgConfig := db.PostgresConfig{
		Host:        cfg.Postgres.Host,
		Port:        cfg.Postgres.Port,
		User:        cfg.Postgres.User,
		Password:    cfg.Postgres.Password,
		DBName:      cfg.Postgres.DBName,
		SSLMode:     cfg.Postgres.SSLMode,
		MaxOpenConn: 25,
		MaxIdleConn: 5,
		MaxLifetime: 5 * time.Minute,
	}

	postgresDB, err := db.NewPostgresDB(ctx, pgConfig, log)
	if err != nil {
		log.Error("Ошибка подключения к PostgreSQL", "error", err.Error())
		os.Exit(1)
	}
	defer postgresDB.Close()

	redisConfig := db.RedisConfig{
		Host:     cfg.Redis.Host,
		Port:     cfg.Redis.Port,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	}

	redisClient, err := db.NewRedisClient(ctx, redisConfig, log)
	if err != nil {
		log.Error("Ошибка подключения к Redis", "error", err.Error())
		os.Exit(1)
	}
	defer redisClient.Close()

	jwtService := auth.NewJWTService(
		cfg.JWT.Secret,
		cfg.JWT.RefreshSecret,
		cfg.JWT.ExpirationHours,
		cfg.JWT.RefreshExpDays,
	)

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(customMiddleware.RequestLogging(log))
	r.Use(customMiddleware.Recovery(log))
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(customMiddleware.CORS)

	router.SetupRoutes(r, postgresDB, redisClient, jwtService, log)

	server := &http.Server{
		Addr:         cfg.Server.Address(),
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  120 * time.Second,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Info("HTTP-сервер запущен", "address", cfg.Server.Address())
		if err = server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("Ошибка запуска HTTP-сервера", "error", err.Error())
			quit <- os.Interrupt
		}
	}()

	<-quit
	log.Info("Получен сигнал завершения, начинаем грациозное завершение...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("Ошибка грациозного завершения сервера", "error", err.Error())
	}

	log.Info("Сервер успешно остановлен")
}
