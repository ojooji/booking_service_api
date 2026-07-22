package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/ojooji/booking-service-api/internal/config"
	"github.com/ojooji/booking-service-api/internal/handler"
	pgxrepo "github.com/ojooji/booking-service-api/internal/repository/pgx"
	"github.com/ojooji/booking-service-api/internal/router"
	"github.com/ojooji/booking-service-api/internal/service"
)

func startPostgres(ctx context.Context) (testcontainers.Container, string, error) {
	req := testcontainers.ContainerRequest{
		Image:        "postgres:16-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "booking",
			"POSTGRES_USER":     "postgres",
			"POSTGRES_PASSWORD": "postgres",
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections").
			WithOccurrence(2).WithStartupTimeout(60 * time.Second),
	}
	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, "", err
	}

	host, err := c.Host(ctx)
	if err != nil {
		return nil, "", err
	}
	port, err := c.MappedPort(ctx, "5432")
	if err != nil {
		return nil, "", err
	}
	dsn := fmt.Sprintf("postgres://postgres:postgres@%s:%s/booking?sslmode=disable", host, port.Port())
	return c, dsn, nil
}

func runMigrations(pool *pgxpool.Pool, migrationsDir string) error {
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return err
	}

	var upFiles []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".up.sql") {
			upFiles = append(upFiles, e.Name())
		}
	}

	for _, f := range upFiles {
		data, err := os.ReadFile(filepath.Join(migrationsDir, f))
		if err != nil {
			return fmt.Errorf("read %s: %w", f, err)
		}
		if _, err := pool.Exec(context.Background(), string(data)); err != nil {
			return fmt.Errorf("exec %s: %w", f, err)
		}
	}
	return nil
}

func setupTestServer(t *testing.T, dsn string) *chi.Mux {
	pool, err := pgxpool.New(context.Background(), dsn)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	err = runMigrations(pool, "../../migrations")
	require.NoError(t, err)

	userRepo := pgxrepo.NewUserRepository(pool)
	serviceRepo := pgxrepo.NewServiceRepository(pool)
	employeeRepo := pgxrepo.NewEmployeeRepository(pool)
	bookingRepo := pgxrepo.NewBookingRepository(pool)

	cfg := &config.Config{
		ServerPort:    "8080",
		JWTSecret:     "test-secret",
		JWTExpiration: 3600,
	}

	authSvc := service.NewAuthService(userRepo, cfg)
	userSvc := service.NewUserService(userRepo)
	serviceSvc := service.NewServiceService(serviceRepo)
	employeeSvc := service.NewEmployeeService(employeeRepo, userRepo)
	bookingSvc := service.NewBookingService(bookingRepo, employeeRepo, serviceRepo)

	authH := handler.NewAuthHandler(authSvc)
	userH := handler.NewUserHandler(userSvc)
	svcH := handler.NewServiceHandler(serviceSvc)
	empH := handler.NewEmployeeHandler(employeeSvc)
	bookingH := handler.NewBookingHandler(bookingSvc)

	return router.New(cfg, authH, userH, svcH, empH, bookingH)
}

func TestIntegration_RegisterAndLogin(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()
	postgres, dsn, err := startPostgres(ctx)
	require.NoError(t, err)
	defer func() { _ = postgres.Terminate(ctx) }()

	r := setupTestServer(t, dsn)

	// Register a user
	regBody := map[string]string{
		"email":      "jane@example.com",
		"password":   "password123",
		"first_name": "Jane",
		"last_name":  "Doe",
	}
	regJSON, _ := json.Marshal(regBody)

	req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewReader(regJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	var regResp struct {
		Status string          `json:"status"`
		Data   json.RawMessage `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &regResp))
	assert.Equal(t, "success", regResp.Status)

	// Login
	loginBody := map[string]string{"email": "jane@example.com", "password": "password123"}
	loginJSON, _ := json.Marshal(loginBody)
	req = httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(loginJSON))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var loginResp struct {
		Status string `json:"status"`
		Data   struct {
			AccessToken string `json:"access_token"`
			UserID      string `json:"user_id"`
			Role        string `json:"role"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &loginResp))
	assert.NotEmpty(t, loginResp.Data.AccessToken)
	assert.Equal(t, "user", loginResp.Data.Role)
}

func TestIntegration_ServiceCatalog(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()
	postgres, dsn, err := startPostgres(ctx)
	require.NoError(t, err)
	defer func() { _ = postgres.Terminate(ctx) }()

	r := setupTestServer(t, dsn)

	req := httptest.NewRequest("GET", "/api/v1/services", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		Status string          `json:"status"`
		Data   json.RawMessage `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "success", resp.Status)
}
