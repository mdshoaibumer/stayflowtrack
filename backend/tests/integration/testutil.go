package integration

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"

	"github.com/stayflow/stayflow-track/internal/config"
	authhandler "github.com/stayflow/stayflow-track/internal/modules/auth/handler"
	authmiddleware "github.com/stayflow/stayflow-track/internal/modules/auth/middleware"
	authrepo "github.com/stayflow/stayflow-track/internal/modules/auth/repository"
	authservice "github.com/stayflow/stayflow-track/internal/modules/auth/service"
	billinghandler "github.com/stayflow/stayflow-track/internal/modules/billing/handler"
	billingrepo "github.com/stayflow/stayflow-track/internal/modules/billing/repository"
	billingservice "github.com/stayflow/stayflow-track/internal/modules/billing/service"
	guesthandler "github.com/stayflow/stayflow-track/internal/modules/guest/handler"
	guestrepo "github.com/stayflow/stayflow-track/internal/modules/guest/repository"
	guestservice "github.com/stayflow/stayflow-track/internal/modules/guest/service"
	hkhandler "github.com/stayflow/stayflow-track/internal/modules/housekeeping/handler"
	hkrepo "github.com/stayflow/stayflow-track/internal/modules/housekeeping/repository"
	hkservice "github.com/stayflow/stayflow-track/internal/modules/housekeeping/service"
	notifprovider "github.com/stayflow/stayflow-track/internal/modules/notifications/provider"
	notifrepo "github.com/stayflow/stayflow-track/internal/modules/notifications/repository"
	notifservice "github.com/stayflow/stayflow-track/internal/modules/notifications/service"
	prophandler "github.com/stayflow/stayflow-track/internal/modules/property/handler"
	proprepo "github.com/stayflow/stayflow-track/internal/modules/property/repository"
	propservice "github.com/stayflow/stayflow-track/internal/modules/property/service"
	reshandler "github.com/stayflow/stayflow-track/internal/modules/reservation/handler"
	resrepo "github.com/stayflow/stayflow-track/internal/modules/reservation/repository"
	resservice "github.com/stayflow/stayflow-track/internal/modules/reservation/service"
	"github.com/stayflow/stayflow-track/internal/platform/database"
	"github.com/stayflow/stayflow-track/internal/platform/email"
	"github.com/stayflow/stayflow-track/internal/platform/storage"
	"github.com/stayflow/stayflow-track/internal/shared/audit"
)

// TestEnv holds the test server and database for integration tests.
type TestEnv struct {
	Server   *httptest.Server
	DB       *pgxpool.Pool
	Postgres *embeddedpostgres.EmbeddedPostgres
}

// projectRoot returns the backend/ directory path regardless of where the test runs.
func projectRoot() string {
	_, filename, _, _ := runtime.Caller(0)
	// tests/integration/testutil.go -> go up 2 levels to backend/
	return filepath.Join(filepath.Dir(filename), "..", "..")
}

// SetupTestEnv starts an embedded PostgreSQL, runs migrations, and returns a ready test server.
func SetupTestEnv() (*TestEnv, error) {
	root := projectRoot()
	migrationsDir := filepath.Join(root, "migrations")

	port := uint32(15432)

	pg := embeddedpostgres.NewDatabase(embeddedpostgres.DefaultConfig().
		Username("testuser").
		Password("testpass").
		Database("stayflow_test").
		Port(port).
		StartTimeout(90 * time.Second))

	if err := pg.Start(); err != nil {
		return nil, fmt.Errorf("start embedded postgres: %w", err)
	}

	dsn := fmt.Sprintf("postgres://testuser:testpass@localhost:%d/stayflow_test?sslmode=disable", port)

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		_ = pg.Stop()
		return nil, fmt.Errorf("connect to embedded postgres: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		_ = pg.Stop()
		return nil, fmt.Errorf("ping embedded postgres: %w", err)
	}

	// Run migrations
	if err := runMigrations(ctx, pool, migrationsDir); err != nil {
		pool.Close()
		_ = pg.Stop()
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	jwtCfg := config.JWTConfig{
		AccessSecret:      "integration-test-access-secret-32chars!",
		RefreshSecret:     "integration-test-refresh-secret-32chars",
		AccessExpiration:  15 * time.Minute,
		RefreshExpiration: 168 * time.Hour,
		Issuer:            "stayflow-integration-test",
	}

	server := buildServer(pool, jwtCfg)

	return &TestEnv{
		Server:   server,
		DB:       pool,
		Postgres: pg,
	}, nil
}

// Teardown cleans up the test environment.
func (e *TestEnv) Teardown() {
	e.Server.Close()
	e.DB.Close()
	if e.Postgres != nil {
		_ = e.Postgres.Stop()
	}
}

func runMigrations(ctx context.Context, pool *pgxpool.Pool, dir string) error {
	files, err := collectMigrationFiles(dir)
	if err != nil {
		return err
	}

	for _, f := range files {
		sql, err := os.ReadFile(f)
		if err != nil {
			return fmt.Errorf("read %s: %w", f, err)
		}
		if _, err := pool.Exec(ctx, string(sql)); err != nil {
			return fmt.Errorf("exec %s: %w", filepath.Base(f), err)
		}
	}
	return nil
}

func buildServer(pool *pgxpool.Pool, jwtCfg config.JWTConfig) *httptest.Server {
	log := zerolog.New(os.Stdout).With().Timestamp().Logger()
	tenantDB := database.NewTenantPool(pool)

	// Repos
	authRepo := authrepo.New(pool)
	propRepo := proprepo.New(tenantDB)
	guestRepo := guestrepo.New(tenantDB)
	resRepo := resrepo.New(tenantDB)
	billingRepo := billingrepo.New(tenantDB)
	hkRepo := hkrepo.New(tenantDB)
	notifRepo := notifrepo.New(tenantDB)

	// Services
	authSvc := authservice.New(authRepo, jwtCfg)
	propSvc := propservice.New(propRepo)

	storageCfg := config.StorageConfig{Provider: "local", Bucket: os.TempDir()}
	store, _ := storage.New(storageCfg)

	guestSvc := guestservice.New(guestRepo, store)
	notifSvc := notifservice.New(notifRepo, notifprovider.NewLogProvider(), log)
	resSvc := resservice.New(resRepo, propRepo, notifSvc)
	billingSvc := billingservice.New(billingRepo, store)
	hkSvc := hkservice.New(hkRepo)

	auditLog := audit.New(pool)

	// Handlers
	emailSender := email.New(config.EmailConfig{Enabled: false}, "stayflow-test")
	authH := authhandler.New(authSvc, log, auditLog, emailSender, "app.test.local")
	propH := prophandler.New(propSvc, log)
	guestH := guesthandler.New(guestSvc, log)
	resH := reshandler.New(resSvc, log)
	billingH := billinghandler.New(billingSvc, log, auditLog)
	hkH := hkhandler.New(hkSvc, log)

	// Middleware
	authMw := authmiddleware.New(authSvc, log)

	// Router
	r := chi.NewRouter()
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	r.Route("/api/v1", func(r chi.Router) {
		// Public auth routes
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authH.RegisterTenant)
			r.Post("/login", authH.Login)
			r.Post("/refresh", authH.RefreshToken)
		})

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(authMw.Authenticate)
			r.Use(authmiddleware.TenantContext)

			r.Post("/auth/logout", authH.Logout)

			// Properties
			r.Route("/properties", func(r chi.Router) {
				r.Post("/", propH.CreateProperty)
				r.Get("/", propH.ListProperties)
				r.Get("/{propertyID}", propH.GetProperty)
				r.Put("/{propertyID}", propH.UpdateProperty)
				r.Post("/{propertyID}/unit-types", propH.CreateUnitType)
				r.Get("/{propertyID}/unit-types", propH.ListUnitTypes)
				r.Post("/{propertyID}/units", propH.CreateUnit)
				r.Get("/{propertyID}/units", propH.ListUnits)
			})

			// Guests
			r.Route("/guests", func(r chi.Router) {
				r.Post("/", guestH.CreateGuest)
				r.Get("/", guestH.ListGuests)
				r.Get("/{guestID}", guestH.GetGuest)
				r.Put("/{guestID}", guestH.UpdateGuest)
			})

			// Reservations
			r.Route("/reservations", func(r chi.Router) {
				r.Post("/", resH.CreateReservation)
				r.Get("/", resH.ListReservations)
				r.Get("/{reservationID}", resH.GetReservation)
				r.Post("/{reservationID}/confirm", resH.ConfirmReservation)
				r.Post("/{reservationID}/cancel", resH.CancelReservation)
			})

			// Billing
			r.Route("/billing", func(r chi.Router) {
				r.Get("/folios/reservation/{reservationID}", billingH.GetFolioByReservation)
				r.Post("/charges", billingH.AddCharge)
				r.Post("/payments", billingH.RecordPayment)
			})

			// Housekeeping
			r.Route("/housekeeping", func(r chi.Router) {
				r.Post("/tasks", hkH.CreateTask)
				r.Get("/tasks", hkH.ListTasks)
				r.Post("/tasks/status", hkH.UpdateStatus)
			})
		})
	})

	return httptest.NewServer(r)
}

// collectMigrationFiles returns only the .up.sql files sorted in order.
func collectMigrationFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".up.sql") {
			files = append(files, filepath.Join(dir, entry.Name()))
		}
	}
	sort.Strings(files)
	return files, nil
}
