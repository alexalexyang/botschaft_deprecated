package main

import (
	"log"
	"net/http"

	"github.com/alexalexyang/botschaft/controllers"
	"github.com/gorilla/mux"
)

func main() {
	// http.HandleFunc("/", controllers.IndexHandler)
	log.Fatal(http.ListenAndServe(":3000", initRouter()))
}

func initRouter() *mux.Router {
	router := mux.NewRouter()
	// root should be .Methods("GET")
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	router.HandleFunc("/", controllers.IndexHandler)
	router.HandleFunc("/createuser", controllers.CreateUserHandler)
	router.HandleFunc("/createbot", controllers.CreateBotHandler)
	router.HandleFunc("/createbotpois", controllers.CreateBotPoisHandler)
	// router.HandleFunc("/createentry", controllers.CreateHandler).Methods("POST")
	return router
}
