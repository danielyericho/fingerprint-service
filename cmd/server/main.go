package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"fingerprint-service/internal/api"
)

func switchToExeDir() error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("os.Executable: %w", err)
	}
	dir := filepath.Dir(exe)
	if err := os.Chdir(dir); err != nil {
		return fmt.Errorf("os.Chdir(%s): %w", dir, err)
	}
	return nil
}

func runHTTPServer(ctx context.Context, addr string) (err error) {
	srv, err := api.NewServer(nil)
	if err != nil {
		return fmt.Errorf("init server: %w", err)
	}
	defer func() {
		_ = srv.Close()
	}()

	server := &http.Server{
		Addr:    addr,
		Handler: srv.Handler(),
	}

	log.Printf("fingerprint-service listening on %s", addr)

	serverErr := make(chan error, 1)
	go func() {
		serverErr <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_ = server.Shutdown(shutdownCtx)
		err := <-serverErr
		if err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("serve: %w", err)
		}
		return nil

	case err := <-serverErr:
		if err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("serve: %w", err)
		}
		return nil
	}
}

func main() {
	addr := flag.String("addr", ":8080", "HTTP listen address")
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Printf("fingerprint-service %s\n", version)
		return
	}

	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
	log.Printf("fingerprint-service %s starting (pid=%d, os=%s, arch=%s)", version, os.Getpid(), runtime.GOOS, runtime.GOARCH)

	if runtime.GOOS == "windows" {
		if err := switchToExeDir(); err != nil {
			log.Fatalf("switch to exe dir: %v", err)
		}
	}

	if handled, err := runWindowsService(func(ctx context.Context) error {
		return runHTTPServer(ctx, *addr)
	}); handled {
		if err != nil {
			log.Fatalf("service execution failed: %v", err)
		}
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := runHTTPServer(ctx, *addr); err != nil {
		log.Fatalf("%v", err)
	}
}
