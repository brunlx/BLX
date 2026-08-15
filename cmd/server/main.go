// pentest-commander: gerador de comandos de pentest.
// Escolha uma ferramenta, responda às perguntas e receba comandos prontos.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/Brunlx/BLX/internal/api"
	"github.com/Brunlx/BLX/internal/static"
	"github.com/Brunlx/BLX/internal/tools"
)

// version é injetada no build via -ldflags "-X main.version=<versão>".
var version = "dev"

func main() {
	printVersion := flag.Bool("version", false, "exibe a versão e sai")
	noBrowser := flag.Bool("no-browser", false, "não abre o navegador ao iniciar (uso servidor)")
	flag.Parse()
	if *printVersion {
		fmt.Println(version)
		return
	}

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
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
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

	if !*noBrowser {
		browserURL := "http://" + addr
		if host == "0.0.0.0" || host == "::" {
			browserURL = "http://127.0.0.1:" + port
		}
		go func() {
			if waitServerReady(browserURL, 3*time.Second) {
				openBrowser(browserURL)
			}
		}()
	}

	<-ctx.Done()
	log.Println("desligando servidor...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("erro ao desligar: %v", err)
	}
	log.Println("servidor encerrado.")
}

// waitServerReady aguarda o servidor responder em baseURL até timeout.
func waitServerReady(baseURL string, timeout time.Duration) bool {
	client := &http.Client{Timeout: 500 * time.Millisecond}
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := client.Get(baseURL + "/api/health")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return true
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	return false
}

// openBrowser abre a URL no navegador padrão do sistema.
func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	if err := cmd.Start(); err != nil {
		log.Printf("aviso: não foi possível abrir o navegador: %v", err)
		return
	}
	go cmd.Wait()
}
