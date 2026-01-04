package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"net/http"

	"github.com/emiliopalmerini/treni/internal/app"
	"github.com/emiliopalmerini/treni/internal/server"
)

func main() {
	cfg, err := app.New()
	if err != nil {
		log.Fatal(err)
	}

	httpSrv := server.NewHTTPServer(cfg)
	go func() {
		log.Printf("http server listening on %s", cfg.Addr)
		if err := httpSrv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := httpSrv.Shutdown(ctx); err != nil {
		log.Printf("http shutdown error: %v", err)
	}
}
