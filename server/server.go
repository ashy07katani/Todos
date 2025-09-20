package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"todos/config"
	"todos/mail"
	"todos/router"
)

func StartServer() {
	appConfig := config.LoadConfiguration()
	appHostAndPort := fmt.Sprintf("%s:%d", appConfig.Host, appConfig.Port)

	db, err := config.DBinit(&appConfig.DBconfig)
	authConfig := &config.AuthConfig{
		JWTSecret:  appConfig.AuthConfig.JWTSecret,
		AccessTTL:  appConfig.AuthConfig.AccessTTL,
		RefreshTTL: appConfig.AuthConfig.RefreshTTL,
	}
	mailConfig := &mail.Mail{
		From:     appConfig.Mail.From,
		Password: appConfig.Mail.Password,
		Host:     appConfig.Mail.Host,
		Port:     appConfig.Mail.Port,
	}
	if err != nil {
		panic("Cannot Start the application")
	}
	r := router.NewRouter(db, authConfig, mailConfig)
	serv := http.Server{
		Addr:    appHostAndPort,
		Handler: r,
	}
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		if err := serv.ListenAndServe(); err != nil {
			log.Println("Server Exited Successfully")
		}
	}()
	<-sig
	log.Println("Shutting down gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Existing connections ko close karne ka request
	if err := serv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

}
