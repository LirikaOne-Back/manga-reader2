package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jmoiron/sqlx"
	"manga-reader2/internal/api/handler"
	customMiddleware "manga-reader2/internal/api/middleware"
	"manga-reader2/internal/common/logger"
	"manga-reader2/internal/domain/repository"
	"manga-reader2/internal/infrastructure/auth"
	"manga-reader2/internal/infrastructure/db"
	"manga-reader2/internal/infrastructure/repository/postgres"
	"manga-reader2/internal/infrastructure/repository/redis"
	"manga-reader2/internal/usecase"
	"net/http"
)

// SetupRoutes настраивает все маршруты приложения
func SetupRoutes(
	r *chi.Mux,
	postgresDB *db.PostgresDB,
	redisClient *db.RedisClient,
	jwtService *auth.JWTService,
	log logger.Logger,
) {
	mangaRepo := postgres.NewMangaRepository(postgresDB.GetDB(), log)
	chapterRepo := postgres.NewChapterRepository(postgresDB.GetDB(), log)
	pageRepo := postgres.NewPageRepository(postgresDB.GetDB(), log)
	userRepo := postgres.NewUserRepository(postgresDB.GetDB(), log)

	cacheRepo := redis.NewCacheRepository(redisClient, log)
	analyticsRepo := redis.NewAnalyticsRepository(redisClient, log)

	mangaUseCase := usecase.NewMangaUseCase(mangaRepo, cacheRepo, analyticsRepo, log)
	chapterUseCase := usecase.NewChapterUseCase(chapterRepo, mangaRepo, cacheRepo, analyticsRepo, log)
	pageUseCase := usecase.NewPageUseCase(pageRepo, chapterRepo, cacheRepo, analyticsRepo, log)
	userUseCase := usecase.NewUserUseCase(userRepo, jwtService, log)
	analyticsUseCase := usecase.NewAnalyticsUseCase(analyticsRepo, mangaRepo, chapterRepo, log)

	mangaHandler := handler.NewMangaHandler(mangaUseCase, log)
	chapterHandler := handler.NewChapterHandler(chapterUseCase, log)
	pageHandler := handler.NewPageHandler(pageUseCase, log)
	userHandler := handler.NewUserHandler(userUseCase, log)
	analyticsHandler := handler.NewAnalyticsHandler(analyticsUseCase, log)

	authMiddleware := customMiddleware.Authentication(jwtService, log)

	adminMiddleware := customMiddleware.RequireRole("admin")

	// Общие маршруты
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// API v1
	r.Route("/api/v1", func(r chi.Router) {
		// Маршруты для пользователей
		r.Route("/users", func(r chi.Router) {
			r.Post("/register", userHandler.Register)
			r.Post("/login", userHandler.Login)
			r.Post("/refresh", userHandler.RefreshToken)

			// Маршруты, требующие аутентификации
			r.Group(func(r chi.Router) {
				r.Use(authMiddleware)

				r.Get("/me", userHandler.GetProfile)
				r.Put("/me", userHandler.UpdateProfile)
				r.Post("/logout", userHandler.Logout)

				// Закладки
				r.Get("/bookmarks", userHandler.GetBookmarks)
				r.Post("/bookmarks", userHandler.AddBookmark)
				r.Delete("/bookmarks/{mangaID}", userHandler.RemoveBookmark)

				// История чтения
				r.Get("/history", userHandler.GetReadingHistory)
				r.Delete("/history/{id}", userHandler.RemoveFromHistory)
			})

			// Маршруты для администраторов
			r.Group(func(r chi.Router) {
				r.Use(authMiddleware)
				r.Use(adminMiddleware)

				r.Get("/", userHandler.ListUsers)
				r.Get("/{id}", userHandler.GetUser)
				r.Put("/{id}", userHandler.UpdateUser)
				r.Delete("/{id}", userHandler.DeleteUser)
			})
		})

		// Маршруты для манги
		r.Route("/manga", func(r chi.Router) {
			r.Get("/", mangaHandler.List)
			r.Get("/popular", mangaHandler.GetPopular)
			r.Get("/{id}", mangaHandler.GetByID)
			r.Get("/{id}/chapters", mangaHandler.GetChapters)

			// Маршруты для администраторов
			r.Group(func(r chi.Router) {
				r.Use(authMiddleware)
				r.Use(adminMiddleware)

				r.Post("/", mangaHandler.Create)
				r.Put("/{id}", mangaHandler.Update)
				r.Delete("/{id}", mangaHandler.Delete)
			})
		})

		// Маршруты для глав
		r.Route("/chapters", func(r chi.Router) {
			r.Get("/{id}", chapterHandler.GetByID)
			r.Get("/{id}/pages", chapterHandler.GetPages)

			// Маршруты для администраторов
			r.Group(func(r chi.Router) {
				r.Use(authMiddleware)
				r.Use(adminMiddleware)

				r.Post("/", chapterHandler.Create)
				r.Put("/{id}", chapterHandler.Update)
				r.Delete("/{id}", chapterHandler.Delete)
			})
		})

		// Маршруты для страниц
		r.Route("/pages", func(r chi.Router) {
			r.Get("/{id}", pageHandler.GetByID)
			r.Get("/{id}/image", pageHandler.ServeImage)

			// Маршруты для администраторов
			r.Group(func(r chi.Router) {
				r.Use(authMiddleware)
				r.Use(adminMiddleware)

				r.Post("/", pageHandler.Create)
				r.Post("/upload", pageHandler.UploadImage)
				r.Put("/{id}", pageHandler.Update)
				r.Delete("/{id}", pageHandler.Delete)
			})
		})

		// Маршруты для аналитики
		r.Route("/analytics", func(r chi.Router) {
			r.Get("/manga/top", analyticsHandler.GetTopManga)
			r.Get("/chapters/top", analyticsHandler.GetTopChapters)

			// Маршруты для администраторов
			r.Group(func(r chi.Router) {
				r.Use(authMiddleware)
				r.Use(adminMiddleware)

				r.Post("/reset/daily", analyticsHandler.ResetDailyStats)
				r.Post("/reset/weekly", analyticsHandler.ResetWeeklyStats)
				r.Post("/reset/monthly", analyticsHandler.ResetMonthlyStats)
				r.Get("/stats", analyticsHandler.GetStats)
			})
		})
	})

	// Маршрут для Swagger UI
	r.Get("/swagger/*", http.StripPrefix("/swagger/", http.FileServer(http.Dir("./docs/swagger"))).ServeHTTP)
}

// Эти функции-заглушки будут заменены на реальные реализации позже
func setupRepositories(db *sqlx.DB, redisClient *db.RedisClient, log logger.Logger) (
	repository.MangaRepository,
	repository.ChapterRepository,
	repository.PageRepository,
	repository.UserRepository,
	repository.CacheRepository,
	repository.AnalyticsRepository,
) {
	mangaRepo := postgres.NewMangaRepository(db, log)
	chapterRepo := postgres.NewChapterRepository(db, log)
	pageRepo := postgres.NewPageRepository(db, log)
	userRepo := postgres.NewUserRepository(db, log)
	cacheRepo := redis.NewCacheRepository(redisClient, log)
	analyticsRepo := redis.NewAnalyticsRepository(redisClient, log)

	return mangaRepo, chapterRepo, pageRepo, userRepo, cacheRepo, analyticsRepo
}
