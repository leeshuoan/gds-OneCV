package utils

import (
	"net/http"
	"encoding/json"
	"strings"
)

func SendJSONError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	response := map[string]string{"message": message}
	json.NewEncoder(w).Encode(response)
}

func ParseMentionedStudents(notificationText string) []string {
	mentionedStudents := []string{}
	words := strings.Fields(notificationText)
	for _, word := range words {
		if strings.HasPrefix(word, "@") && strings.Contains(word, "@") {
			mentionedStudents = append(mentionedStudents, strings.TrimPrefix(word, "@"))
		}
	}

	return mentionedStudents
}
