package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"

	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
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
	calendarhandler "github.com/stayflow/stayflow-track/internal/modules/calendar/handler"
	calendarrepo "github.com/stayflow/stayflow-track/internal/modules/calendar/repository"
	calendarservice "github.com/stayflow/stayflow-track/internal/modules/calendar/service"
	checkinouthandler "github.com/stayflow/stayflow-track/internal/modules/checkinout/handler"
	checkinoutrepo "github.com/stayflow/stayflow-track/internal/modules/checkinout/repository"
	checkinoutservice "github.com/stayflow/stayflow-track/internal/modules/checkinout/service"
	dashboardhandler "github.com/stayflow/stayflow-track/internal/modules/dashboard/handler"
	dashboardrepo "github.com/stayflow/stayflow-track/internal/modules/dashboard/repository"
	dashboardservice "github.com/stayflow/stayflow-track/internal/modules/dashboard/service"
	guesthandler "github.com/stayflow/stayflow-track/internal/modules/guest/handler"
	guestrepo "github.com/stayflow/stayflow-track/internal/modules/guest/repository"
	guestservice "github.com/stayflow/stayflow-track/internal/modules/guest/service"
	hkhandler "github.com/stayflow/stayflow-track/internal/modules/housekeeping/handler"
	hkrepo "github.com/stayflow/stayflow-track/internal/modules/housekeeping/repository"
	hkservice "github.com/stayflow/stayflow-track/internal/modules/housekeeping/service"
	laundryhandler "github.com/stayflow/stayflow-track/internal/modules/laundry/handler"
	laundryrepo "github.com/stayflow/stayflow-track/internal/modules/laundry/repository"
	laundryservice "github.com/stayflow/stayflow-track/internal/modules/laundry/service"
	notifhandler "github.com/stayflow/stayflow-track/internal/modules/notifications/handler"
	notifprovider "github.com/stayflow/stayflow-track/internal/modules/notifications/provider"
	notifrepo "github.com/stayflow/stayflow-track/internal/modules/notifications/repository"
	notifservice "github.com/stayflow/stayflow-track/internal/modules/notifications/service"
	opshandler "github.com/stayflow/stayflow-track/internal/modules/operations/handler"
	opsservice "github.com/stayflow/stayflow-track/internal/modules/operations/service"
	prophandler "github.com/stayflow/stayflow-track/internal/modules/property/handler"
	proprepo "github.com/stayflow/stayflow-track/internal/modules/property/repository"
	propservice "github.com/stayflow/stayflow-track/internal/modules/property/service"
	reshandler "github.com/stayflow/stayflow-track/internal/modules/reservation/handler"
	resrepo "github.com/stayflow/stayflow-track/internal/modules/reservation/repository"
	resservice "github.com/stayflow/stayflow-track/internal/modules/reservation/service"
	saashandler "github.com/stayflow/stayflow-track/internal/modules/saas/handler"
	saasrepo "github.com/stayflow/stayflow-track/internal/modules/saas/repository"
	saasservice "github.com/stayflow/stayflow-track/internal/modules/saas/service"
	"github.com/stayflow/stayflow-track/internal/platform/database"
	"github.com/stayflow/stayflow-track/internal/platform/storage"
	"github.com/stayflow/stayflow-track/internal/shared/audit"
	"github.com/stayflow/stayflow-track/internal/shared/middleware"
)

const (
	pgPort   = 15433
	httpPort = "8080"
)

func main() {
	log := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "15:04:05"}).
		With().Timestamp().Logger()

	fmt.Println("═══════════════════════════════════════════════════")
	fmt.Println("  StayFlow E2E Test Environment Launcher")
	fmt.Println("═══════════════════════════════════════════════════")

	// ─── 1. Start Embedded PostgreSQL ───
	log.Info().Uint32("port", pgPort).Msg("starting embedded PostgreSQL...")

	pg := embeddedpostgres.NewDatabase(embeddedpostgres.DefaultConfig().
		Username("stayflow").
		Password("stayflow_e2e").
		Database("stayflow_e2e").
		Port(pgPort).
		StartTimeout(120 * time.Second))

	if err := pg.Start(); err != nil {
		log.Fatal().Err(err).Msg("failed to start embedded PostgreSQL")
	}
	defer func() {
		log.Info().Msg("stopping embedded PostgreSQL...")
		_ = pg.Stop()
	}()

	log.Info().Msg("✓ Embedded PostgreSQL started")

	// ─── 2. Connect and run migrations ───
	dsn := fmt.Sprintf("postgres://stayflow:stayflow_e2e@localhost:%d/stayflow_e2e?sslmode=disable", pgPort)
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to embedded PostgreSQL")
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatal().Err(err).Msg("failed to ping database")
	}

	migrationsDir := filepath.Join(".", "migrations")
	if err := runMigrations(ctx, pool, migrationsDir, log); err != nil {
		log.Fatal().Err(err).Msg("failed to run migrations")
	}
	log.Info().Msg("✓ All migrations applied")
	tenantDB := database.NewTenantPool(pool)

	// ─── 3. Build and start the HTTP server (mirrors cmd/server) ───
	jwtCfg := config.JWTConfig{
		AccessSecret:      "e2e-access-secret-must-be-32-chars!",
		RefreshSecret:     "e2e-refresh-secret-must-be-32-chars",
		AccessExpiration:  30 * time.Minute,
		RefreshExpiration: 168 * time.Hour,
		Issuer:            "stayflow-e2e",
	}

	storageCfg := config.StorageConfig{Provider: "local", Bucket: filepath.Join(os.TempDir(), "stayflow-e2e-uploads")}
	_ = os.MkdirAll(storageCfg.Bucket, 0o755)
	store, _ := storage.New(storageCfg)

	// Repos
	authRepo := authrepo.New(pool)
	propRepo := proprepo.New(tenantDB)
	guestRepo := guestrepo.New(tenantDB)
	resRepo := resrepo.New(tenantDB)
	calendarRepo := calendarrepo.New(tenantDB)
	checkinoutRepo := checkinoutrepo.New(tenantDB)
	billingRepo := billingrepo.New(tenantDB)
	hkRepo := hkrepo.New(tenantDB)
	laundryRepo := laundryrepo.New(tenantDB)
	dashboardRepo := dashboardrepo.New(tenantDB)
	notifRepo := notifrepo.New(tenantDB)
	saasRepo := saasrepo.New(tenantDB)

	// Services
	notifProvider := notifprovider.NewLogProvider()
	authSvc := authservice.New(authRepo, jwtCfg)
	propSvc := propservice.New(propRepo)
	guestSvc := guestservice.New(guestRepo, store)
	notifSvc := notifservice.New(notifRepo, notifProvider, log)
	resSvc := resservice.New(resRepo, propRepo, notifSvc)
	calendarSvc := calendarservice.New(calendarRepo)
	checkinoutSvc := checkinoutservice.New(checkinoutRepo, notifSvc)
	billingSvc := billingservice.New(billingRepo, store)
	hkSvc := hkservice.New(hkRepo)
	laundrySvc := laundryservice.New(laundryRepo)
	dashboardSvc := dashboardservice.New(dashboardRepo)
	saasSvc := saasservice.New(saasRepo, nil) // no razorpay in e2e
	opsSvc := opsservice.New(tenantDB)

	auditLog := audit.New(pool)

	// Handlers
	authH := authhandler.New(authSvc, log, auditLog)
	propH := prophandler.New(propSvc, log)
	guestH := guesthandler.New(guestSvc, log)
	resH := reshandler.New(resSvc, log)
	calendarH := calendarhandler.New(calendarSvc, log)
	checkinoutH := checkinouthandler.New(checkinoutSvc, log, auditLog)
	billingH := billinghandler.New(billingSvc, log, auditLog)
	hkH := hkhandler.New(hkSvc, log)
	laundryH := laundryhandler.New(laundrySvc, log)
	dashboardH := dashboardhandler.New(dashboardSvc, log)
	notifH := notifhandler.New(notifSvc, log)
	saasH := saashandler.New(saasSvc, nil, log)
	opsH := opshandler.New(opsSvc, log)

	// Middleware
	authMw := authmiddleware.New(authSvc, log)

	// Router (exact mirror of cmd/server/main.go routes)
	r := chi.NewRouter()
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(middleware.RequestLogger(log))
	r.Use(chimiddleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://127.0.0.1:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Tenant-ID"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	r.Use(middleware.MaxBodySize(1 << 20))

	// Health checks
	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	r.Get("/ready", func(w http.ResponseWriter, _ *http.Request) {
		if err := pool.Ping(context.Background()); err != nil {
			http.Error(w, `{"status":"not_ready"}`, http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ready"}`))
	})

	// API v1
	r.Route("/api/v1", func(r chi.Router) {
		// Public auth routes
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authH.RegisterTenant)
			r.Post("/login", authH.Login)
			r.Post("/refresh", authH.RefreshToken)
			r.Post("/password-reset/request", authH.RequestPasswordReset)
			r.Post("/password-reset/confirm", authH.ConfirmPasswordReset)
		})

		r.Get("/plans", saasH.ListPlans)
		r.Post("/webhooks/notifications", notifH.Webhook)
		r.Post("/webhooks/razorpay", saasH.RazorpayWebhook)

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(authMw.Authenticate)
			r.Use(authmiddleware.TenantContext)

			r.Post("/auth/logout", authH.Logout)

			// Users
			r.Route("/users", func(r chi.Router) {
				r.Use(authmiddleware.RequireRole("super_admin", "property_admin"))
				r.Post("/", authH.CreateUser)
			})

			// Properties
			r.Route("/properties", func(r chi.Router) {
				r.Use(authmiddleware.RequireRole("super_admin", "property_admin"))
				r.Post("/", propH.CreateProperty)
				r.Get("/", propH.ListProperties)
				r.Get("/{propertyID}", propH.GetProperty)
				r.Put("/{propertyID}", propH.UpdateProperty)
				r.Post("/{propertyID}/unit-types", propH.CreateUnitType)
				r.Get("/{propertyID}/unit-types", propH.ListUnitTypes)
				r.Post("/{propertyID}/units", propH.CreateUnit)
				r.Get("/{propertyID}/units", propH.ListUnits)
				r.Put("/{propertyID}/units/{unitID}", propH.UpdateUnit)
				r.Delete("/{propertyID}/units/{unitID}", propH.DeleteUnit)
				r.Patch("/{propertyID}/units/{unitID}/status", propH.ChangeUnitStatus)
				r.Get("/{propertyID}/units/search", propH.SearchUnits)
			})

			// Guests
			r.Route("/guests", func(r chi.Router) {
				r.Post("/", guestH.CreateGuest)
				r.Get("/", guestH.ListGuests)
				r.Get("/{guestID}", guestH.GetGuest)
				r.Put("/{guestID}", guestH.UpdateGuest)
				r.Get("/{guestID}/history", guestH.GetGuestHistory)
				r.Get("/search", guestH.SearchGuests)
				r.With(middleware.MaxBodySize(10<<20)).Post("/{guestID}/documents", guestH.UploadDocument)
				r.Get("/{guestID}/documents", guestH.ListDocuments)
			})

			// Reservations
			r.Route("/reservations", func(r chi.Router) {
				r.Post("/", resH.CreateReservation)
				r.Get("/", resH.ListReservations)
				r.Get("/{reservationID}", resH.GetReservation)
				r.Put("/{reservationID}", resH.UpdateReservation)
				r.Post("/{reservationID}/cancel", resH.CancelReservation)
				r.Post("/{reservationID}/confirm", resH.ConfirmReservation)
				r.Post("/{reservationID}/check-in", resH.CheckIn)
				r.Post("/{reservationID}/check-out", resH.CheckOut)
				r.Get("/availability", resH.CheckAvailability)
			})

			// Calendar
			r.Route("/calendar", func(r chi.Router) {
				r.Get("/{propertyID}", calendarH.GetCalendarView)
				r.Post("/move", calendarH.MoveBooking)
				r.Get("/{propertyID}/occupancy", calendarH.GetOccupancyStats)
			})

			// Operations
			r.Route("/operations", func(r chi.Router) {
				r.Use(authmiddleware.RequireRole("super_admin", "property_admin", "receptionist"))
				r.Post("/check-in", checkinoutH.CheckIn)
				r.Post("/check-out", checkinoutH.CheckOut)
				r.Post("/walk-in", checkinoutH.WalkIn)
				r.Get("/pre-checkout", checkinoutH.GetPreCheckoutSummary)
				r.Post("/no-show", opsH.MarkNoShow)
				r.Post("/extend-stay", opsH.ExtendStay)
				r.Post("/room-move", opsH.MoveRoom)
				r.Post("/maintenance-block", opsH.CreateMaintenanceBlock)
				r.Post("/refund-deposit", opsH.RefundDeposit)
			})

			// Billing
			r.Route("/billing", func(r chi.Router) {
				r.Get("/folios/{folioID}", billingH.GetFolio)
				r.Get("/folios/{folioID}/summary", billingH.GetFolioSummary)
				r.Get("/folios/reservation/{reservationID}", billingH.GetFolioByReservation)
				r.Get("/folios/{folioID}/items", billingH.ListLineItems)
				r.Get("/folios/{folioID}/payments", billingH.ListPayments)
				r.Post("/charges", billingH.AddCharge)
				r.Post("/charges/void", billingH.VoidLineItem)
				r.Post("/payments", billingH.RecordPayment)
				r.Get("/invoices", billingH.ListInvoices)
				r.Get("/invoices/{invoiceID}", billingH.GetInvoice)
				r.Post("/invoices/{invoiceID}/pdf", billingH.GenerateInvoicePDF)
				r.Get("/invoices/{invoiceID}/pdf", billingH.GetInvoicePDF)
			})

			// Housekeeping
			r.Route("/housekeeping", func(r chi.Router) {
				r.Post("/tasks", hkH.CreateTask)
				r.Get("/tasks", hkH.ListTasks)
				r.Get("/tasks/{taskID}", hkH.GetTask)
				r.Post("/tasks/assign", hkH.AssignTask)
				r.Post("/tasks/status", hkH.UpdateStatus)
				r.Get("/stats/{propertyID}", hkH.GetStats)
			})

			// Laundry
			r.Route("/laundry", func(r chi.Router) {
				r.Post("/orders", laundryH.CreateOrder)
				r.Get("/orders", laundryH.ListOrders)
				r.Get("/orders/{orderID}", laundryH.GetOrder)
				r.Post("/orders/status", laundryH.UpdateStatus)
				r.Post("/orders/{orderID}/post-to-folio", laundryH.PostToFolio)
				r.Get("/stats/{propertyID}", laundryH.GetStats)
				r.Post("/rate-card", laundryH.CreateRateCard)
				r.Get("/rate-card/{propertyID}", laundryH.ListRateCards)
				r.Put("/rate-card", laundryH.UpdateRateCard)
			})

			// Dashboard
			r.Route("/dashboard", func(r chi.Router) {
				r.Get("/{propertyID}", dashboardH.GetDashboard)
				r.Get("/{propertyID}/revenue-trend", dashboardH.GetRevenueTrend)
				r.Get("/{propertyID}/daily-collection", dashboardH.GetDailyCollection)
				r.Get("/{propertyID}/outstanding-dues", dashboardH.GetOutstandingDues)
				r.Get("/{propertyID}/end-of-day", dashboardH.GetEndOfDaySummary)
				r.Post("/{propertyID}/close-day", dashboardH.CloseDay)
			})

			// Notifications
			r.Route("/notifications", func(r chi.Router) {
				r.Post("/templates", notifH.UpsertTemplate)
				r.Get("/templates", notifH.ListTemplates)
				r.Post("/send-test", notifH.SendTest)
				r.Get("/logs", notifH.ListLogs)
			})

			// Subscription
			r.Route("/subscription", func(r chi.Router) {
				r.Get("/", saasH.GetSubscription)
				r.Post("/", saasH.CreateSubscription)
				r.Put("/plan", saasH.ChangePlan)
				r.Post("/cancel", saasH.CancelSubscription)
				r.Post("/checkout", saasH.CreateCheckout)
				r.Post("/verify-payment", saasH.VerifyPayment)
				r.Get("/billing-events", saasH.ListBillingEvents)
			})

			// Onboarding
			r.Route("/onboarding", func(r chi.Router) {
				r.Post("/{propertyID}/init", saasH.InitOnboarding)
				r.Get("/{propertyID}/steps", saasH.GetOnboardingSteps)
				r.Post("/complete-step", saasH.CompleteStep)
			})

			// Admin
			r.Route("/admin", func(r chi.Router) {
				r.Use(authMw.RequirePlatformAdmin)
				r.Get("/tenants", saasH.AdminListTenants)
				r.Get("/metrics", saasH.AdminGetMetrics)
				r.Get("/plans", saasH.AdminListPlans)
			})
		})
	})

	// ─── 4. Start HTTP server ───
	srv := &http.Server{
		Addr:              ":" + httpPort,
		Handler:           r,
		ReadTimeout:       15 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Info().Str("port", httpPort).Msg("E2E backend server listening")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server error")
		}
	}()

	fmt.Println("═══════════════════════════════════════════════════")
	fmt.Println("  ✓ Backend ready at http://localhost:" + httpPort)
	fmt.Printf("  ✓ PostgreSQL at localhost:%d\n", pgPort)
	fmt.Println("  Now run: cd frontend && npx playwright test")
	fmt.Println("  Press Ctrl+C to stop")
	fmt.Println("═══════════════════════════════════════════════════")

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down E2E environment...")
	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
	log.Info().Msg("done")
}

func runMigrations(ctx context.Context, pool *pgxpool.Pool, dir string, log zerolog.Logger) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".up.sql") {
			files = append(files, filepath.Join(dir, entry.Name()))
		}
	}
	sort.Strings(files)

	for _, f := range files {
		sql, err := os.ReadFile(f)
		if err != nil {
			return fmt.Errorf("read %s: %w", f, err)
		}
		if _, err := pool.Exec(ctx, string(sql)); err != nil {
			return fmt.Errorf("exec %s: %w", filepath.Base(f), err)
		}
		log.Debug().Str("file", filepath.Base(f)).Msg("migration applied")
	}
	return nil
}
