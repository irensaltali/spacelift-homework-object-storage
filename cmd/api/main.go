package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/irensaltali/object-storage-gateway/internal/api"
	"github.com/irensaltali/object-storage-gateway/internal/discovery"
	"github.com/irensaltali/object-storage-gateway/internal/storage"
)

const (
	serverAddr          = ":3000"
	discoveryTimeout    = 10 * time.Second
	shutdownGracePeriod = 10 * time.Second
)

func main() {
	printCredits()

	if err := run(); err != nil {
		log.Printf("application error: %v", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, cancel := context.WithTimeout(context.Background(), discoveryTimeout)
	defer cancel()

	instances, err := discovery.DiscoverInstances(ctx)
	if err != nil {
		return fmt.Errorf("failed to discover minio instances: %w", err)
	}
	if len(instances) == 0 {
		return fmt.Errorf("no minio instances found")
	}

	log.Printf("discovered %d minio instance(s)", len(instances))
	for _, inst := range instances {
		log.Printf("  - instance %s at %s:%s", inst.ID, inst.Host, inst.Port)
	}

	bucketName := os.Getenv("BUCKET_NAME")
	if bucketName == "" {
		bucketName = "objects"
	}

	gateway, err := storage.NewGateway(instances, storage.WithBucketName(bucketName))
	if err != nil {
		return fmt.Errorf("failed to create gateway: %w", err)
	}
	defer func() {
		if closeErr := gateway.Close(); closeErr != nil {
			log.Printf("failed to close gateway: %v", closeErr)
		}
	}()

	server := &http.Server{
		Addr:              serverAddr,
		Handler:           api.NewRouter(gateway),
		ReadHeaderTimeout: 5 * time.Second,
	}

	serverErrCh := make(chan error, 1)
	go func() {
		log.Printf("API server is running on %s", serverAddr)
		if listenErr := server.ListenAndServe(); listenErr != nil && !errors.Is(listenErr, http.ErrServerClosed) {
			serverErrCh <- listenErr
		}
	}()

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(signalCh)

	select {
	case serverErr := <-serverErrCh:
		return fmt.Errorf("http server failed: %w", serverErr)
	case sig := <-signalCh:
		log.Printf("received signal %s, starting graceful shutdown", sig)

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownGracePeriod)
		defer shutdownCancel()

		if shutdownErr := server.Shutdown(shutdownCtx); shutdownErr != nil {
			return fmt.Errorf("graceful shutdown failed: %w", shutdownErr)
		}

		return nil
	}
}

func printCredits() {
	println(`
   /$$
   /_/
 /$$$$$$                                      /$$$$$$            /$$   /$$               /$$|
|_  $$_/                                     /$$__  $$          | $$  | $$              | $$|
  | $$    /$$$$$$   /$$$$$$  /$$$$$$$       | $$  \__/  /$$$$$$ | $$ /$$$$$$    /$$$$$$ | $$ /$$
  | $$   /$$__  $$ /$$__  $$| $$__  $$      |  $$$$$$  |____  $$| $$|_  $$_/   |____  $$| $$| $$
  | $$  | $$  \__/| $$$$$$$$| $$  \ $$       \____  $$  /$$$$$$$| $$  | $$      /$$$$$$$| $$| $$
  | $$  | $$      | $$_____/| $$  | $$       /$$  \ $$ /$$__  $$| $$  | $$ /$$ /$$__  $$| $$| $$
 /$$$$$$| $$      |  $$$$$$$| $$  | $$      |  $$$$$$/|  $$$$$$$| $$  |  $$$$/|  $$$$$$$| $$| $$
|______/|__/       \_______/|__/  |__/       \______/  \_______/|__/   \___/   \_______/|__/|__/


                                                   /$$ /$$  /$$$$$$   /$$           /$$                                                                           /$$
                                                  | $$|__/ /$$__  $$ | $$          | $$                                                                          | $$
  /$$$$$$$  /$$$$$$   /$$$$$$   /$$$$$$$  /$$$$$$ | $$ /$$| $$  \__//$$$$$$        | $$$$$$$   /$$$$$$  /$$$$$$/$$$$   /$$$$$$  /$$  /$$  /$$  /$$$$$$   /$$$$$$ | $$   /$$
 /$$_____/ /$$__  $$ |____  $$ /$$_____/ /$$__  $$| $$| $$| $$$$   |_  $$_/        | $$__  $$ /$$__  $$| $$_  $$_  $$ /$$__  $$| $$ | $$ | $$ /$$__  $$ /$$__  $$| $$  /$$/
|  $$$$$$ | $$  \ $$  /$$$$$$$| $$      | $$$$$$$$| $$| $$| $$_/     | $$          | $$  \ $$| $$  \ $$| $$ \ $$ \ $$| $$$$$$$$| $$ | $$ | $$| $$  \ $$| $$  \__/| $$$$$$/
 \____  $$| $$  | $$ /$$__  $$| $$      | $$_____/| $$| $$| $$       | $$ /$$      | $$  | $$| $$  | $$| $$ | $$ | $$| $$_____/| $$ | $$ | $$| $$  | $$| $$      | $$_  $$
 /$$$$$$$/| $$$$$$$/|  $$$$$$$|  $$$$$$$|  $$$$$$$| $$| $$| $$       |  $$$$/      | $$  | $$|  $$$$$$/| $$ | $$ | $$|  $$$$$$$|  $$$$$/$$$$/|  $$$$$$/| $$      | $$ \  $$
|_______/ | $$____/  \_______/ \_______/ \_______/|__/|__/|__/        \___/        |__/  |__/ \______/ |__/ |__/ |__/ \_______/ \_____/\___/  \______/ |__/      |__/  \__/
          | $$
          | $$
          |__/
`)
}
