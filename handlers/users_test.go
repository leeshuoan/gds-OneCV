package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/leeshuoan/gds-OneCV/mocks"
	"github.com/lib/pq"
)

func TestRegister(t *testing.T) {
	db, mock := mocks.NewMock()
	defer db.Close()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Register(w, r, db)
	})

	t.Run("Successful Registration", func(t *testing.T) {
		mock.ExpectExec(`INSERT INTO registrations`).WithArgs("teacher@example.com", "studentjon@example.com").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(`INSERT INTO registrations`).WithArgs("teacher@example.com", "studenthon@example.com").WillReturnResult(sqlmock.NewResult(1, 1))

		req := httptest.NewRequest("POST", "/register", strings.NewReader(`{"teacher": "teacher@example.com", "students": ["studentjon@example.com", "studenthon@example.com"]}`))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusNoContent {
			t.Errorf("Expected status %d; got %d", http.StatusNoContent, status)
		}
	})

	t.Run("Missing Teacher Field", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/register", strings.NewReader(`{"students": ["student@example.com"]}`))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusBadRequest {
			t.Errorf("Expected status %d; got %d", http.StatusBadRequest, status)
		}

		expectedErrorMessage := "Both 'teacher' and 'students' fields are required in the request body"
		responseBody := rr.Body.String()
		if !strings.Contains(responseBody, expectedErrorMessage) {
			t.Errorf("Expected error message '%s' in response body; got '%s'", expectedErrorMessage, responseBody)
		}
	})

	t.Run("Missing Students Field", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/register", strings.NewReader(`{"teacher": "teacher@example.com"}`))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusBadRequest {
			t.Errorf("Expected status %d; got %d", http.StatusBadRequest, status)
		}

		expectedErrorMessage := "Both 'teacher' and 'students' fields are required in the request body"
		responseBody := rr.Body.String()
		if !strings.Contains(responseBody, expectedErrorMessage) {
			t.Errorf("Expected error message '%s' in response body; got '%s'", expectedErrorMessage, responseBody)
		}
	})

	t.Run("Duplicate Student Registration", func(t *testing.T) {
		mock.ExpectExec(`INSERT INTO registrations`).WithArgs("teacher@example.com", "student@example.com").WillReturnError(&pq.Error{Code: "23505"})

		req := httptest.NewRequest("POST", "/register", strings.NewReader(`{"teacher": "teacher@example.com", "students": ["student@example.com"]}`))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusBadRequest {
			t.Errorf("Expected status %d; got %d", http.StatusBadRequest, status)
		}

		expectedErrorMessage := "student@example.com is already registered with this teacher"
		responseBody := rr.Body.String()
		if !strings.Contains(responseBody, expectedErrorMessage) {
			t.Errorf("Expected error message '%s' in response body; got '%s'", expectedErrorMessage, responseBody)
		}
	})

	t.Run("Non-Existent Teacher", func(t *testing.T) {
		mock.ExpectExec(`INSERT INTO registrations`).WithArgs("teacher@example.com", "student@example.com").WillReturnError(&pq.Error{Constraint: "registrations_teacher_email_fkey"})

		req := httptest.NewRequest("POST", "/register", strings.NewReader(`{"teacher": "teacher@example.com", "students": ["student@example.com"]}`))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusBadRequest {
			t.Errorf("Expected status %d; got %d", http.StatusBadRequest, status)
		}

		expectedErrorMessage := "Teacher teacher@example.com does not exist in the database"
		responseBody := rr.Body.String()
		if !strings.Contains(responseBody, expectedErrorMessage) {
			t.Errorf("Expected error message '%s' in response body; got '%s'", expectedErrorMessage, responseBody)
		}
	})

	t.Run("Non-Existent Student", func(t *testing.T) {
		mock.ExpectExec(`INSERT INTO registrations`).WithArgs("teacher@example.com", "student@example.com").WillReturnError(&pq.Error{Constraint: "registrations_student_email_fkey"})

		req := httptest.NewRequest("POST", "/register", strings.NewReader(`{"teacher": "teacher@example.com", "students": ["student@example.com"]}`))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusBadRequest {
			t.Errorf("Expected status %d; got %d", http.StatusBadRequest, status)
		}

		expectedErrorMessage := "Student student@example.com does not exist in the database"
		responseBody := rr.Body.String()
		if !strings.Contains(responseBody, expectedErrorMessage) {
			t.Errorf("Expected error message '%s' in response body; got '%s'", expectedErrorMessage, responseBody)
		}
	})
}

func TestCommonStudents(t *testing.T) {
	db, mock := mocks.NewMock()
	defer db.Close()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		CommonStudents(w, r, db)
	})

	t.Run("Successful Common Students", func(t *testing.T) {
		mock.ExpectPrepare("SELECT student_email").ExpectQuery().
			WithArgs("teacherken@gmail.com", "teacherjoe@gmail.com", 2).
			WillReturnRows(sqlmock.NewRows([]string{"student_email"}).
				AddRow("commonstudent1@gmail.com").
				AddRow("commonstudent2@gmail.com"))

		req := httptest.NewRequest("GET", "/common-students?teacher=teacherken%40gmail.com&teacher=teacherjoe%40gmail.com", nil)

		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("Expected status %d; got %d", http.StatusOK, status)
		}

		expectedResponse := `{"students":["commonstudent1@gmail.com","commonstudent2@gmail.com"]}`
		if !strings.Contains(rr.Body.String(), expectedResponse) {
			t.Errorf("Expected response body %s; got %s", expectedResponse, rr.Body.String())
		}
	})

	t.Run("No Teacher in Query", func(t *testing.T) {
		mock.ExpectPrepare("SELECT student_email").ExpectQuery().
			WithArgs("teacherken@gmail.com", "teacherjoe@gmail.com", 2).
			WillReturnRows(sqlmock.NewRows([]string{"student_email"}).
				AddRow("commonstudent1@gmail.com").
				AddRow("commonstudent2@gmail.com"))

		req := httptest.NewRequest("GET", "/common-students", nil)

		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusBadRequest {
			t.Errorf("Expected status %d; got %d", http.StatusBadRequest, status)
		}

		expectedErrorMessage := `{"message":"At least one teacher is required in the query parameter"}`
		if !strings.Contains(rr.Body.String(), expectedErrorMessage) {
			t.Errorf("Expected response body %s; got %s", expectedErrorMessage, rr.Body.String())
		}
	})
}

func TestSuspend(t *testing.T) {
	db, mock := mocks.NewMock()
	defer db.Close()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Suspend(w, r, db)
	})

	t.Run("Successful Suspension", func(t *testing.T) {
		mock.ExpectExec("UPDATE students SET is_suspended = true").
			WithArgs("studentmary@gmail.com").
			WillReturnResult(sqlmock.NewResult(0, 1))

		req := httptest.NewRequest("POST", "/suspend", strings.NewReader(`{"student": "studentmary@gmail.com"}`))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusNoContent {
			t.Errorf("Expected status %d; got %d", http.StatusNoContent, status)
		}
	})

	t.Run("Missing Student in Request Body", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/suspend", strings.NewReader(`{}`))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusBadRequest {
			t.Errorf("Expected status %d; got %d", http.StatusBadRequest, status)
		}

		expectedErrorMessage := `{"message":"'student' is required in the request body"}`
		if !strings.Contains(rr.Body.String(), expectedErrorMessage) {
			t.Errorf("Expected response body %s; got %s", expectedErrorMessage, rr.Body.String())
		}
	})

	t.Run("Student Not Found in Database", func(t *testing.T) {
		mock.ExpectExec("UPDATE students SET is_suspended = true").
			WithArgs("nonexistentstudent@gmail.com").
			WillReturnError(&pq.Error{Code: "23503"})

		req := httptest.NewRequest("POST", "/suspend", strings.NewReader(`{"student": "nonexistentstudent@gmail.com"}`))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusBadRequest {
			t.Errorf("Expected status %d; got %d", http.StatusBadRequest, status)
		}

		expectedErrorMessage := `{"message":"Student nonexistentstudent@gmail.com does not exist in the database"}`
		if !strings.Contains(rr.Body.String(), expectedErrorMessage) {
			t.Errorf("Expected response body %s; got %s", expectedErrorMessage, rr.Body.String())
		}
	})
}

func TestRetrieveForNotifications(t *testing.T) {
	db, mock := mocks.NewMock()
	defer db.Close()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		RetrieveForNotifications(w, r, db)
	})

	t.Run("Successful Notification Retrieval with mentions", func(t *testing.T) {
		mock.ExpectQuery(`SELECT DISTINCT r.student_email`).
			WithArgs("teacherken@gmail.com", pq.Array([]string{"studentagnes@gmail.com", "studentmiche@gmail.com"})).
			WillReturnRows(sqlmock.NewRows([]string{"student_email"}).
				AddRow("studentbob@gmail.com").
				AddRow("studentagnes@gmail.com").
				AddRow("studentmiche@gmail.com"))

		reqBody := `{"teacher": "teacherken@gmail.com", "notification": "Hello students! @studentagnes@gmail.com @studentmiche@gmail.com"}`
		req := httptest.NewRequest("POST", "/notifications", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("Expected status %d; got %d", http.StatusOK, status)
		}

		expectedResponse := `{"recipients":["studentbob@gmail.com","studentagnes@gmail.com","studentmiche@gmail.com"]}`
		if !strings.Contains(rr.Body.String(), expectedResponse) {
			t.Errorf("Expected response body to contain %s; got %s", expectedResponse, rr.Body.String())
		}
	})

	t.Run("Successful Notification Retrieval without mentions", func(t *testing.T) {
		mock.ExpectQuery(`SELECT DISTINCT r.student_email`).
			WithArgs("teacherken@gmail.com", pq.Array([]string{})).
			WillReturnRows(sqlmock.NewRows([]string{"student_email"}).
				AddRow("studentbob@gmail.com"))

		reqBody := `{"teacher": "teacherken@gmail.com", "notification": "Hey everybody!"}`
		req := httptest.NewRequest("POST", "/notifications", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("Expected status %d; got %d", http.StatusOK, status)
		}

		expectedResponse := `{"recipients":["studentbob@gmail.com"]}`
		if !strings.Contains(rr.Body.String(), expectedResponse) {
			t.Errorf("Expected response body to contain %s; got %s", expectedResponse, rr.Body.String())
		}
	})

	t.Run("Missing Teacher in Request Body", func(t *testing.T) {
		reqBody := `{"notification": "Hello @student@example.com"}`
		req := httptest.NewRequest("POST", "/notifications", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusBadRequest {
			t.Errorf("Expected status %d; got %d", http.StatusBadRequest, status)
		}

		expectedErrorMessage := `{"message":"Both 'teacher' and 'notification' fields are required in the request body"}`
		if !strings.Contains(rr.Body.String(), expectedErrorMessage) {
			t.Errorf("Expected response body to contain %s; got %s", expectedErrorMessage, rr.Body.String())
		}
	})

	t.Run("Missing Notification in Request Body", func(t *testing.T) {
		reqBody := `{"teacher": "teacherken@gmail.com"}`
		req := httptest.NewRequest("POST", "/notifications", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusBadRequest {
			t.Errorf("Expected status %d; got %d", http.StatusBadRequest, status)
		}

		expectedErrorMessage := `{"message":"Both 'teacher' and 'notification' fields are required in the request body"}`
		if !strings.Contains(rr.Body.String(), expectedErrorMessage) {
			t.Errorf("Expected response body to contain %s; got %s", expectedErrorMessage, rr.Body.String())
		}
	})
}
