package app_test

import (
	"net/http"
	"testing"
	"time"

	"meeting-slot-service/cmd/server/app"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestNewServer_DefaultTimeouts(t *testing.T) {
	router := mux.NewRouter()
	srv := app.NewServer(":9999", router)

	assert.NotNil(t, srv)

	inner := srv.HTTPServer()
	assert.Equal(t, ":9999", inner.Addr)
	assert.Equal(t, 15*time.Second, inner.ReadTimeout)
	assert.Equal(t, 15*time.Second, inner.WriteTimeout)
	assert.Equal(t, 60*time.Second, inner.IdleTimeout)
}

func TestNewServer_UsesProvidedRouter(t *testing.T) {
	router := mux.NewRouter()
	router.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	srv := app.NewServer(":9998", router)
	assert.NotNil(t, srv)
	assert.Equal(t, router, srv.HTTPServer().Handler)
}
