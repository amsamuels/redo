// cmd/server/main.go
package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"

	"redo.ai/internal/server"
	"redo.ai/internal/storage"
	"redo.ai/logger"
)

func main() {
	// Initialize the logger
	err := logger.Init("program.log")
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	defer logger.Close()

	_ = godotenv.Load()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		logger.Fatal("DATABASE_URL is not set")
	}

	db := storage.Connect(dsn)
	defer db.Close()

	port := os.Getenv("PORT")
	if port == "" {
		logger.Fatal("PORT environment variable is not set")
	}

	srv := server.New(db)
	fmt.Println("Starting server...")
	logger.Info("Starting server on :%s", port)
	if err := srv.Start(port); err != nil {
		fmt.Printf("Server failed: %v\n", err)
		logger.Fatal("Server error: %v", err)
	}
}
