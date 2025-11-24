package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/example/prreview/internal/app"
	"github.com/example/prreview/internal/config"
)

func main() {
	cfg := config.LoadFromEnv()
	logger := log.New(os.Stdout, "", log.LstdFlags)

	a, err := app.NewApp(cfg, logger)
	if err != nil {
		logger.Fatalf("init app: %v", err)
	}

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      a.Router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		logger.Printf("listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Println("shutting down")
	_ = srv.Close()
}
