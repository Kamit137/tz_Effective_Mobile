// @title           Subscription Service API
// @version         1.0
// @description     REST API сервис для управления подписками
// @host            localhost:8080
// @BasePath        /
// @schemes         http

package main

import (
	"log"
	"net/http"
	"tz/internal/handlers"
	"tz/internal/repository"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	httpSwagger "github.com/swaggo/http-swagger"

	_ "tz/docs"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found")
	}
	err = repository.InitDB()
	if err != nil {
		log.Fatal("Failed to init DB:", err)
	}
	http.HandleFunc("/subscriptions", handlers.Subscriptions)
	http.HandleFunc("/sumAllSub", handlers.Sum)
	http.HandleFunc("/subscriptions/", handlers.SubscriptionsID)

	http.Handle("/swagger/", httpSwagger.WrapHandler)
	log.Println("Server starting on port :8080")

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
