// Package main is the entry point for the GophKeeper server.
package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/ar11/gophkeeper/internal/server"
	"github.com/ar11/gophkeeper/internal/storage"
	pb "github.com/ar11/gophkeeper/pkg/api/proto"
	"google.golang.org/grpc"
)

var (
	// Version is the application version (set at build time).
	Version = "dev"
	// BuildDate is the build date (set at build time).
	BuildDate = "unknown"
)

func main() {
	// Parse command-line flags
	var (
		address   = flag.String("addr", ":50051", "Server address")
		dbDSN     = flag.String("db", getEnv("DB_DSN", "postgres://localhost:5432/gophkeeper?sslmode=disable"), "Database DSN")
		jwtSecret = flag.String("jwt-secret", getEnv("JWT_SECRET", "change-this-secret-key"), "JWT secret key")
		version   = flag.Bool("version", false, "Print version information")
	)
	flag.Parse()

	// Print version information
	if *version {
		fmt.Printf("GophKeeper Server\n")
		fmt.Printf("Version: %s\n", Version)
		fmt.Printf("Build Date: %s\n", BuildDate)
		return
	}

	log.Printf("Starting GophKeeper Server %s (built %s)", Version, BuildDate)

	// Initialize storage
	log.Printf("Connecting to database: %s", maskDSN(*dbDSN))
	store, err := storage.NewPostgresStorage(*dbDSN)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer store.Close()

	// Initialize schema
	log.Println("Initializing database schema...")
	if err := store.InitSchema(); err != nil {
		log.Fatalf("Failed to initialize schema: %v", err)
	}

	// Create gRPC server
	grpcServer := grpc.NewServer()
	gophkeeperServer := server.NewServer(store, *jwtSecret)
	pb.RegisterGophKeeperServer(grpcServer, gophkeeperServer)

	// Start listening
	listener, err := net.Listen("tcp", *address)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", *address, err)
	}

	log.Printf("Server listening on %s", *address)

	// Handle graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh
		log.Println("Shutting down server...")
		grpcServer.GracefulStop()
	}()

	// Start serving
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func maskDSN(dsn string) string {
	// Simple masking for logging purposes
	if len(dsn) > 20 {
		return dsn[:10] + "..." + dsn[len(dsn)-10:]
	}
	return "***"
}

