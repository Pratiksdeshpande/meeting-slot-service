package middleware

import (
	"log"
	"meeting-slot-service/internal/utils"
	"net/http"
	"runtime/debug"
)

// Recovery middleware recovers from panics and returns 500
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v", err)
				log.Printf("Stack trace:\n%s", debug.Stack())

				utils.WriteInternalError(w, "Internal server error")
			}
		}()

		next.ServeHTTP(w, r)
	})
}
