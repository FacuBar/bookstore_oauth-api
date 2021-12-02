package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

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
	srv := http.Server{
		Handler: router,
		Addr:    ":8081",
	}

	go func() {
		if err = srv.ListenAndServe(); err != nil {
			log.Fatalf("Error while serving", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	<-quit
	log.Println("Shutdown server ...")
	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second)

	defer cancel()

	db.Close()

	go func() {
		if err := srv.Shutdown(ctx); err != nil {
			log.Fatal("Server Shutdown:", err)
		}
	}()

	log.Print("Server exiting")
}
