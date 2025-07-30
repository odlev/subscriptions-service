package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/odlev/subscriptions/docs"
	"github.com/odlev/subscriptions/internal/config"
	"github.com/odlev/subscriptions/internal/handlers"
	"github.com/odlev/subscriptions/internal/storage"
	"github.com/odlev/subscriptions/pkg/sl"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/files" 
)

// @title           Subscription service API
// @version         1.0
// @description     Api for managing subscriptions
// @host            localhost:8080
// @BasePath        /
// @schemes http
func main() {

	cfg := config.MustLoad()

	log := newLogger(cfg.Environment)

	db, err := storage.InitPostgres(log, *cfg)
	if err != nil {
		log.Error("error initialization database", sl.Err(err))
	}

	router := gin.Default()

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.POST("/new", handlers.CreateSubscription(log, db))
	router.GET("/get/:id", handlers.GetSubscription(log, db))
	router.DELETE("/delete/:id", handlers.DeleteSubscription(log, db))
	router.PATCH("/update/:id", handlers.UpdateSubscription(log, db))
	router.GET("/list", handlers.GetListSubscriptions(log, db))

	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Error("failed to start server", "err", err)
	}

}

func newLogger(environment string) *slog.Logger {
	var log *slog.Logger

	switch environment {
	case "local":
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case "prod":
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}
