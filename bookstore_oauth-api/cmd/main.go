package main

import (
	"net/http"

	"github.com/FacuBar/bookstore_oauth-api/pkg/core/services"
	"github.com/FacuBar/bookstore_oauth-api/pkg/infraestructure/clients"
	"github.com/FacuBar/bookstore_oauth-api/pkg/infraestructure/http/rest"
	"github.com/FacuBar/bookstore_oauth-api/pkg/infraestructure/repositories"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}

	db := clients.ConnectDB()

	atr := repositories.NewAccessTokenRepository(db, &http.Client{})
	ats := services.NewAccessTokenService(atr)

	router := rest.Handler(ats)
	router.Run(":8081")
}
