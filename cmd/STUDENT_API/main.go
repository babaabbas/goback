package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/babaabbas/goback/internal/config"
	"github.com/babaabbas/goback/internal/http/handlers/student"
	"github.com/babaabbas/goback/internal/storage/sqlite"
)

func main() {
	//load config
	cfg := config.Must_Load()
	fmt.Println(cfg)
	//database setup
	storage, err := sqlite.New(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize DB: %v", err)
	}
	slog.Info("Storage initialized", slog.String("env", cfg.Env), slog.String("Version", "1."))
	//setup router
	router := http.NewServeMux()
	router.HandleFunc("POST /api/students", student.New(storage))

	//setup server
	server := http.Server{
		Addr:    cfg.Addr,
		Handler: router,
	}
	slog.Info("server started", slog.String("address", cfg.Addr))
	fmt.Printf("Server started %s", cfg.HTTPServer.Addr)
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	<-done
	slog.Info("shutting down the server")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Failde to shutdown the server", slog.String("error", err.Error()))
	}
	slog.Info("Server shutdown successfully")
}
