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
	oauth_grpc "github.com/FacuBar/bookstore_oauth-api/pkg/infraestructure/http/grpc/oauth"
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
	defer db.Close()

	ur := repositories.NewUsersRepository(db)
	us := services.NewUsersService(ur)

	atr := repositories.NewAccessTokenRepository(db, &http.Client{})
	ats := services.NewAccessTokenService(atr, us)

	router := rest.Handler(ats)
	srv := http.Server{
		Handler: router,
		Addr:    ":8081",
	}

	_, err = clients.NewRabbitMQ(os.Getenv("RMQ_URI"))
	if err != nil {
		log.Fatalf("rabitmq error: %v\n", err)
	}

	go func() {
		if err = srv.ListenAndServe(); err != nil {
			log.Fatalf("Error while serving: %v", err)
		}
	}()

	OauthGrpcServer, err := oauth_grpc.NewGRPCServer("0.0.0.0:10000", ats)
	if err != nil {
		log.Fatalf("couldn't serve grpc server, err: %v", err)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	<-quit
	log.Println("Shutdown server ...")
	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second)
	defer cancel()

	go OauthGrpcServer.GracefulStop()
	go func() {
		if err := srv.Shutdown(ctx); err != nil {
			log.Fatal("Server Shutdown:", err)
		}
	}()

	log.Print("Server exiting")
}
