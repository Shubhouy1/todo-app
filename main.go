package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/Shubhouy1/todo-app/database"
	"github.com/Shubhouy1/todo-app/router"
)

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func main() {
	r := router.SetupRouter()

	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "local")
	dbPassword := getEnv("DB_PASSWORD", "local")
	dbName := getEnv("DB_NAME", "mercury-dev")
	sslMode := getEnv("DB_SSLMODE", string(database.SSLModeDisabled))
	serverPort := getEnv("SERVER_PORT", "8080")

	err := database.CreateAndMigrate(
		dbHost,
		dbPort,
		dbUser,
		dbPassword,
		dbName,
		database.SSLMode(sslMode),
	)
	if err != nil {
		panic(err)
	}

	fmt.Println("Server running on port", serverPort)

	if err := http.ListenAndServe(":"+serverPort, r); err != nil {
		panic(err)
	}
}
