package hashmiddleware

import (
	"bytes"
	"io"
	"net/http"

	"github.com/yadmabramov/admAlerting/internal/utils"
)

func HashCheckMiddleware(key string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if key != "" && r.Method == http.MethodPost {
				// Читаем тело запроса
				body, err := io.ReadAll(r.Body)
				if err != nil {
					http.Error(w, "Failed to read request body", http.StatusBadRequest)
					return
				}
				r.Body = io.NopCloser(bytes.NewBuffer(body))

				// Проверяем хеш
				receivedHash := r.Header.Get("HashSHA256")
				if !utils.VerifyHash(body, key, receivedHash) {
					http.Error(w, "Invalid hash", http.StatusBadRequest)
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
