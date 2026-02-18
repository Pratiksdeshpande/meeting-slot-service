package app_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"meeting-slot-service/cmd/server/app"
	"meeting-slot-service/internal/handler"
	"meeting-slot-service/internal/service"

	"github.com/stretchr/testify/assert"
)

// newTestApp builds a minimal *app.App with nil services so that NewRouter
// can be called without a database.  Handler methods are never invoked in
// these tests — we only probe the routing table.
func newTestApp() *app.App {
	userHandler := handler.NewUserHandler(service.NewUserService(nil))
	eventHandler := handler.NewEventHandler(service.NewEventService(nil, nil, nil))
	availabilityHandler := handler.NewAvailabilityHandler(
		service.NewAvailabilityService(nil, nil, nil, nil),
		service.NewRecommendationService(nil, nil, nil),
	)

	return &app.App{
		UserHandler:         userHandler,
		EventHandler:        eventHandler,
		AvailabilityHandler: availabilityHandler,
	}
}

// routeExists returns true when the router matches the given method+path with
// a status code other than 405 Method Not Allowed or 404 Not Found.
func routeExists(t *testing.T, router http.Handler, method, path string) bool {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	// 404 → route not registered; 405 → route exists but wrong method
	return rr.Code != http.StatusNotFound && rr.Code != http.StatusMethodNotAllowed
}

func TestNewRouter_HealthCheck(t *testing.T) {
	router := app.NewRouter(newTestApp())

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "OK", rr.Body.String())
}

func TestNewRouter_UserRoutes(t *testing.T) {
	router := app.NewRouter(newTestApp())

	routes := []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/api/v1/users"},
		{http.MethodGet, "/api/v1/users"},
		{http.MethodGet, "/api/v1/users/abc"},
		{http.MethodPut, "/api/v1/users/abc"},
		{http.MethodDelete, "/api/v1/users/abc"},
	}

	for _, r := range routes {
		t.Run(r.method+" "+r.path, func(t *testing.T) {
			assert.True(t, routeExists(t, router, r.method, r.path),
				"expected route to be registered")
		})
	}
}

func TestNewRouter_EventRoutes(t *testing.T) {
	router := app.NewRouter(newTestApp())

	routes := []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/api/v1/events"},
		{http.MethodGet, "/api/v1/events"},
		{http.MethodGet, "/api/v1/events/abc"},
		{http.MethodPut, "/api/v1/events/abc"},
		{http.MethodDelete, "/api/v1/events/abc"},
		{http.MethodPost, "/api/v1/events/abc/participants"},
		{http.MethodGet, "/api/v1/events/abc/participants"},
		{http.MethodDelete, "/api/v1/events/abc/participants/user1"},
	}

	for _, r := range routes {
		t.Run(r.method+" "+r.path, func(t *testing.T) {
			assert.True(t, routeExists(t, router, r.method, r.path),
				"expected route to be registered")
		})
	}
}

func TestNewRouter_AvailabilityRoutes(t *testing.T) {
	router := app.NewRouter(newTestApp())

	routes := []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/api/v1/events/abc/participants/user1/availability"},
		{http.MethodPut, "/api/v1/events/abc/participants/user1/availability"},
		{http.MethodGet, "/api/v1/events/abc/participants/user1/availability"},
		{http.MethodGet, "/api/v1/events/abc/recommendations"},
	}

	for _, r := range routes {
		t.Run(r.method+" "+r.path, func(t *testing.T) {
			assert.True(t, routeExists(t, router, r.method, r.path),
				"expected route to be registered")
		})
	}
}

func TestNewRouter_UnregisteredRoute(t *testing.T) {
	router := app.NewRouter(newTestApp())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/nonexistent", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestNewRouter_WrongMethod(t *testing.T) {
	router := app.NewRouter(newTestApp())

	// /health only accepts GET
	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}
