package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leeshuoan/gds-OneCV/db"
	"github.com/leeshuoan/gds-OneCV/handlers"
)

func main() {
	router := mux.NewRouter()
	db := db.OpenConnection()
	defer db.Close()

	router.HandleFunc("/api/register", func(w http.ResponseWriter, r *http.Request) {
		handlers.Register(w, r, db)
	}).Methods("POST")
	router.HandleFunc("/api/commonstudents", func(w http.ResponseWriter, r *http.Request) {
		handlers.CommonStudents(w, r, db)
	}).Methods("GET")
	router.HandleFunc("/api/suspend", func(w http.ResponseWriter, r *http.Request) {
		handlers.Suspend(w, r, db)
	}).Methods("POST")
	router.HandleFunc("/api/retrievefornotifications", func(w http.ResponseWriter, r *http.Request) {
		handlers.RetrieveForNotifications(w, r, db)
	}).Methods("POST")

	fmt.Println("Server at 8080")
	log.Fatal(http.ListenAndServe(":8000", router))
}
