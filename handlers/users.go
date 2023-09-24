package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/leeshuoan/gds-OneCV/utils"
	"github.com/leeshuoan/gds-OneCV/models"
	"github.com/lib/pq"
)

func Register(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	var request RegistrationRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		sendJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	if request.Teacher == "" || len(request.Students) == 0 {
		sendJSONError(w, http.StatusBadRequest, "Both 'teacher' and 'students' fields are required in the request body")
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
			return
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

func CommonStudents(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	teacherEmails, ok := r.URL.Query()["teacher"]
	if !ok || len(teacherEmails) < 1 {
		sendJSONError(w, http.StatusBadRequest, "At least one teacher is required in the query parameter")
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
		return
	}

	args = append(args, len(teacherEmails))

	rows, err := sqlStatement.Query(args...)
	if err != nil {
		sendJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	defer rows.Close()

	var students []string
	for rows.Next() {
		var studentEmail string
		if err := rows.Scan(&studentEmail); err != nil {
			sendJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		students = append(students, studentEmail)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(CommonStudentsResponse{Students: students})
}

func Suspend(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	var request SuspendRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		sendJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	if request.Student == "" {
		sendJSONError(w, http.StatusBadRequest, "'student' is required in the request body")
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
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func RetrieveForNotifications(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	var request NotificationRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		sendJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	if request.Teacher == "" || request.Notification == "" {
		sendJSONError(w, http.StatusBadRequest, "Both 'teacher' and 'notification' fields are required in the request body")
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
		return
	}
	defer rows.Close()

	var students []string
	for rows.Next() {
		var studentEmail string
		if err := rows.Scan(&studentEmail); err != nil {
			sendJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		students = append(students, studentEmail)

	}
	response := NotificationResponse{Recipients: students}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
