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
)

func main() {
	//load config
	cfg := config.Must_Load()
	fmt.Println(cfg)
	//database setup
	//setup router
	router := http.NewServeMux()
	router.HandleFunc("POST /api/students", student.New())

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
		err := server.ListenAndServe()
		if err != nil {
			log.Fatal("failed to start a server")
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
