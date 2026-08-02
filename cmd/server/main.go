// pentest-commander: gerador de comandos de pentest.
// Escolha uma ferramenta, responda às perguntas e receba comandos prontos.
package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Brunlx/BLX/internal/api"
	"github.com/Brunlx/BLX/internal/static"
	"github.com/Brunlx/BLX/internal/tools"
)

func main() {
	host := os.Getenv("HOST")
	if host == "" {
		host = "127.0.0.1"
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := net.JoinHostPort(host, port)

	staticHandler, err := static.Handler()
	if err != nil {
		log.Fatalf("erro ao carregar frontend embutido: %v", err)
	}

	server := api.New(tools.NewCatalog())
	handler := server.Routes(staticHandler)

	httpServer := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// Graceful shutdown on SIGINT/SIGTERM.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		if host == "0.0.0.0" || host == "::" {
			log.Printf("BLX ouvindo em %s — outras máquinas da rede podem acessar. Libere a porta no firewall se necessário.", addr)
		} else {
			log.Printf("BLX ouvindo em http://%s", addr)
		}
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("servidor encerrado inesperadamente: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("desligando servidor...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("erro ao desligar: %v", err)
	}
	log.Println("servidor encerrado.")
}
