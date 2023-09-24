package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
	"github.com/lib/pq"
)

type RegistrationRequest struct {
	Teacher  string   `json:"teacher"`
	Students []string `json:"students"`
}

type SuspendRequest struct {
	Student string `json:"student"`
}

type NotificationRequest struct {
	Teacher      string `json:"teacher"`
	Notification string `json:"notification"`
}

type CommonStudentsResponse struct {
	Students []string `json:"students"`
}

type NotificationResponse struct {
	Recipients []string `json:"recipients"`
}

func sendJSONError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	response := map[string]string{"message": message}
	json.NewEncoder(w).Encode(response)
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
	router.HandleFunc("/api/commonstudents", CommonStudents).Methods("GET")
	router.HandleFunc("/api/suspend", Suspend).Methods("POST")
	router.HandleFunc("/api/retrievefornotifications", RetrieveForNotifications).Methods("POST")

	fmt.Println("Server at 8080")
	log.Fatal(http.ListenAndServe(":8000", router))
}

func Register(w http.ResponseWriter, r *http.Request) {
	db := openConnection()
	var request RegistrationRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		sendJSONError(w, http.StatusBadRequest, err.Error())
		db.Close()
		return
	}

	if request.Teacher == "" || len(request.Students) == 0 {
		sendJSONError(w, http.StatusBadRequest, "Both 'teacher' and 'students' fields are required in the request body")
		db.Close()
		return
	}

	for _, studentEmail := range request.Students {
		sqlStatement := `INSERT INTO registrations (teacher_email, student_email) VALUES ($1, $2)`
		_, err := db.Exec(sqlStatement, request.Teacher, studentEmail)

		if err != nil {
			if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
				errorMessage := fmt.Sprintf("%s is already registered with this teacher", studentEmail)
				sendJSONError(w, http.StatusBadRequest, errorMessage)
			} else if pqErr != nil {
				constraintName := pqErr.Constraint
				if constraintName == "registrations_teacher_email_fkey" {
					errorMessage := fmt.Sprintf("Teacher %s does not exist in the database", request.Teacher)
					sendJSONError(w, http.StatusBadRequest, errorMessage)
				} else if constraintName == "registrations_student_email_fkey" {
					errorMessage := fmt.Sprintf("Student %s does not exist in the database", studentEmail)
					sendJSONError(w, http.StatusBadRequest, errorMessage)
				} else {
					sendJSONError(w, http.StatusBadRequest, err.Error())
				}
			} else {
				sendJSONError(w, http.StatusBadRequest, err.Error())
			}
			db.Close()
			return
		}
	}

	w.WriteHeader(http.StatusNoContent)

	defer db.Close()
}

func CommonStudents(w http.ResponseWriter, r *http.Request) {
	db := openConnection()

	teacherEmails, ok := r.URL.Query()["teacher"]
	if !ok || len(teacherEmails) < 1 {
		sendJSONError(w, http.StatusBadRequest, "At least one teacher is required in the query parameter")
		db.Close()
		return
	}

	placeholders := make([]string, len(teacherEmails))
	args := make([]interface{}, len(teacherEmails))
	for i, email := range teacherEmails {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = email
	}

	query := fmt.Sprintf(`
			SELECT student_email
			FROM registrations
			WHERE teacher_email IN (%s)
			GROUP BY student_email
			HAVING COUNT(DISTINCT teacher_email) = $%d
	`, strings.Join(placeholders, ","), len(teacherEmails)+1)

	sqlStatement, err := db.Prepare(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		db.Close()
		return
	}

	args = append(args, len(teacherEmails))

	rows, err := sqlStatement.Query(args...)
	if err != nil {
		sendJSONError(w, http.StatusBadRequest, err.Error())
		db.Close()
		return
	}
	defer rows.Close()

	var students []string
	for rows.Next() {
		var studentEmail string
		if err := rows.Scan(&studentEmail); err != nil {
			sendJSONError(w, http.StatusBadRequest, err.Error())
			db.Close()
			return
		}
		students = append(students, studentEmail)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(CommonStudentsResponse{Students: students})

	defer db.Close()
}

func Suspend(w http.ResponseWriter, r *http.Request) {
	db := openConnection()
	var request SuspendRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		sendJSONError(w, http.StatusBadRequest, err.Error())
		db.Close()
		return
	}

	if request.Student == "" {
		sendJSONError(w, http.StatusBadRequest, "'student' is required in the request body")
		db.Close()
		return
	}

	sqlStatement := `UPDATE students SET is_suspended = true WHERE student_email = $1`
	_, err := db.Exec(sqlStatement, request.Student)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23503" {
			errorMessage := fmt.Sprintf("Student %s does not exist in the database", request.Student)
			sendJSONError(w, http.StatusBadRequest, errorMessage)
		} else {
			sendJSONError(w, http.StatusBadRequest, err.Error())
		}
		db.Close()
		return
	}

	w.WriteHeader(http.StatusNoContent)

	defer db.Close()
}

func RetrieveForNotifications(w http.ResponseWriter, r *http.Request) {
	db := openConnection()
	var request NotificationRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		sendJSONError(w, http.StatusBadRequest, err.Error())
		db.Close()
		return
	}

	if request.Teacher == "" || request.Notification == "" {
		sendJSONError(w, http.StatusBadRequest, "Both 'teacher' and 'notification' fields are required in the request body")
		db.Close()
		return
	}

	teacherEmail := request.Teacher
	notification := request.Notification

	mentionedStudents := parseMentionedStudents(notification)

	query := `
		SELECT DISTINCT r.student_email
		FROM registrations r, students s
		WHERE r.student_email = s.student_email AND teacher_email = $1 AND is_suspended = false
		UNION
		SELECT DISTINCT student_email
		FROM students
		WHERE student_email = ANY($2) AND is_suspended = false
	`
	rows, err := db.Query(query, teacherEmail, pq.Array(mentionedStudents))
	if err != nil {
		sendJSONError(w, http.StatusBadRequest, err.Error())
		db.Close()
		return
	}
	defer rows.Close()

	var students []string
	for rows.Next() {
		var studentEmail string
		if err := rows.Scan(&studentEmail); err != nil {
			sendJSONError(w, http.StatusBadRequest, err.Error())
			db.Close()
			return
		}
		students = append(students, studentEmail)

	}
	response := NotificationResponse{Recipients: students}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	defer db.Close()
}

func parseMentionedStudents(notificationText string) []string {
	mentionedStudents := []string{}
	words := strings.Fields(notificationText)
	for _, word := range words {
		if strings.HasPrefix(word, "@") && strings.Contains(word, "@") {
			mentionedStudents = append(mentionedStudents, strings.TrimPrefix(word, "@"))
		}
	}

	return mentionedStudents
}
