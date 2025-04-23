// cmd/server/main.go
package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"

	"redo.ai/internal/server"
	"redo.ai/internal/storage"
)

func main() {
	_ = godotenv.Load()
	db := storage.Connect(os.Getenv("DATABASE_URL"))
	defer db.Close()

	srv := server.New(db)

	log.Println("Server running on :8080")
	if err := srv.Start(":8080"); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
