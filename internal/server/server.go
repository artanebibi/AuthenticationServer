package server

import (
	"AuthServer/internal/repository"
	"AuthServer/internal/server/handlers"
	"fmt"
	"net/http"
	"os"
	"strconv"

	_ "github.com/joho/godotenv/autoload"

	"AuthServer/internal/database"
)

func NewServer() *http.Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))

	dbService := database.New() // This is of type database.Service
	database.RunMigrations(dbService.DB())
	//
	userRepo := repository.NewUserRepository(dbService) // Pass the service

	serverHandler := &handlers.Server{
		Port:     port,
		Db:       dbService,
		UserRepo: userRepo,
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: serverHandler.RegisterRoutes(),
	}

	return srv
}
