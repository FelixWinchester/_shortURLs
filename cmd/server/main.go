package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/winchester/shorturls/internal/cache"
	"github.com/winchester/shorturls/internal/config"
	"github.com/winchester/shorturls/internal/database"
	"github.com/winchester/shorturls/internal/handlers"
	"github.com/winchester/shorturls/internal/middleware"
	"github.com/winchester/shorturls/internal/repository"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	pg, err := database.NewPostgresPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("postgres connection failed: %v", err)
	}
	defer pg.Close()
	log.Println("postgres connected")

	rdb, err := database.NewRedisClient(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		log.Fatalf("redis connection failed: %v", err)
	}
	defer rdb.Close()
	log.Println("redis connected")

	if err := database.RunMigrations(ctx, pg); err != nil {
		log.Fatalf("migrations failed: %v", err)
	}
	log.Println("migrations applied")

	r := gin.Default()

	linkRepo := repository.NewLinkRepo(pg)
	analyticsRepo := repository.NewAnalyticsRepo(pg)
	linkCache := cache.NewLinkCache(rdb)
	linkHandler := handlers.NewLinkHandler(linkRepo, linkCache)
	rateLimiter := cache.NewRateLimiter(rdb, 20, time.Minute)

	redirectHandler := handlers.NewRedirectHandler(linkRepo, analyticsRepo, linkCache)
	analyticsHandler := handlers.NewAnalyticsHandler(linkRepo, analyticsRepo)
	qrHandler := handlers.NewQRHandler(linkRepo, analyticsRepo, rdb)

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "ShortURLs",
			"version": "1.0",
			"endpoints": gin.H{
				"POST /links":            "Create a short link",
				"GET /links/{alias}":      "Get link info",
				"PATCH /links/{alias}":    "Update link",
				"DELETE /links/{alias}":   "Delete link",
				"GET /{alias}":            "Redirect to original URL",
				"GET /links/{alias}/qr":    "Generate QR code",
				"GET /analytics":          "Global analytics + top 5",
				"GET /analytics/{alias}":  "Per-link analytics",
			},
		})
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	links := r.Group("/links", middleware.RateLimiter(rateLimiter))
	{
		links.POST("", linkHandler.Create)
		links.GET("/:alias", linkHandler.Get)
		links.PATCH("/:alias", linkHandler.Update)
		links.DELETE("/:alias", linkHandler.Delete)
		links.GET("/:alias/qr", qrHandler.GetQR)
	}

	analytics := r.Group("/analytics", middleware.RateLimiter(rateLimiter))
	{
		analytics.GET("", analyticsHandler.GetGlobal)
		analytics.GET("/:alias", analyticsHandler.GetByAlias)
	}

	r.GET("/:alias", middleware.RateLimiter(rateLimiter), redirectHandler.Redirect)

	srv := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()
	log.Printf("server started on :%s", cfg.ServerPort)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("forced shutdown: %v", err)
	}
}
