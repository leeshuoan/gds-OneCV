package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/leeshuoan/gds-OneCV/mocks"
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

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unfulfilled mock expectations: %s", err)
		}
	})
}
