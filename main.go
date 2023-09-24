package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/lib/pq"
)

type RegistrationRequest struct {
	Teacher  string   `json:"teacher"`
	Students []string `json:"students"`
}

func openConnection() *sql.DB {
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	connStr := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", dbUser, dbPassword, dbName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	return db
}

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/api/register", Register).Methods("POST")

	fmt.Println("Server at 8080")
	log.Fatal(http.ListenAndServe(":8000", router))
}

func Register(w http.ResponseWriter, r *http.Request) {
	db := openConnection()
	var request RegistrationRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		db.Close()
		return
	}

	if request.Teacher == "" || len(request.Students) == 0 {
		http.Error(w, "Both 'teacher' and 'students' fields are required in the request body", http.StatusBadRequest)
		db.Close()
		return
	}

	for _, studentEmail := range request.Students {
		sqlStatement := `INSERT INTO registrations (teacher_email, student_email) VALUES ($1, $2)`
		_, err := db.Exec(sqlStatement, request.Teacher, studentEmail)

		if err != nil {
			if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
				errorMessage := fmt.Sprintf("%s is already registered with this teacher", studentEmail)
				http.Error(w, errorMessage, http.StatusBadRequest)
			} else if pqErr != nil {
				constraintName := pqErr.Constraint
				if constraintName == "registrations_teacher_email_fkey" {
					errorMessage := fmt.Sprintf("Teacher %s does not exist in the database", request.Teacher)
					http.Error(w, errorMessage, http.StatusBadRequest)
				} else if constraintName == "registrations_student_email_fkey" {
					errorMessage := fmt.Sprintf("Student %s does not exist in the database", studentEmail)
					http.Error(w, errorMessage, http.StatusBadRequest)
				} else {
					http.Error(w, err.Error(), http.StatusBadRequest)
				}
			} else {
				http.Error(w, err.Error(), http.StatusBadRequest)
			}
			db.Close()
			return
		}
	}

	w.WriteHeader(http.StatusNoContent)

	defer db.Close()
}
