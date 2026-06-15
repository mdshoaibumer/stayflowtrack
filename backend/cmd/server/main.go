package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/stayflow/stayflow-track/internal/config"
	"github.com/stayflow/stayflow-track/internal/modules/auth/handler"
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
	saasrazorpay "github.com/stayflow/stayflow-track/internal/modules/saas/razorpay"
	saasrepo "github.com/stayflow/stayflow-track/internal/modules/saas/repository"
	saasservice "github.com/stayflow/stayflow-track/internal/modules/saas/service"
	"github.com/stayflow/stayflow-track/internal/platform/database"
	"github.com/stayflow/stayflow-track/internal/platform/logger"
	"github.com/stayflow/stayflow-track/internal/platform/storage"
	"github.com/stayflow/stayflow-track/internal/shared/audit"
	"github.com/stayflow/stayflow-track/internal/shared/middleware"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	log := logger.New(cfg.Log.Level, cfg.Log.Format)
	log.Info().Str("env", cfg.App.Env).Msg("starting stayflow-track server")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := database.Connect(ctx, cfg.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer db.Close()

	store, err := storage.New(cfg.Storage)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize storage")
	}

	// Audit logger
	auditLog := audit.New(db)
	_ = auditLog // Used by handlers for recording audit entries

	// Repositories
	authRepo := authrepo.New(db)
	propRepo := proprepo.New(db)
	guestRepo := guestrepo.New(db)
	resRepo := resrepo.New(db)
	calendarRepo := calendarrepo.New(db)
	checkinoutRepo := checkinoutrepo.New(db)
	billingRepo := billingrepo.New(db)
	hkRepo := hkrepo.New(db)
	laundryRepo := laundryrepo.New(db)
	dashboardRepo := dashboardrepo.New(db)
	notifRepo := notifrepo.New(db)
	saasRepo := saasrepo.New(db)

	// Notification provider (env-driven: NOTIFICATION_PROVIDER=log|gupshup)
	var notifProvider notifprovider.Provider
	switch cfg.Notifications.Provider {
	case "gupshup":
		notifProvider = notifprovider.NewGupshupProvider(notifprovider.GupshupConfig{
			APIKey:  cfg.Notifications.GupshupAPIKey,
			AppName: cfg.Notifications.GupshupApp,
		})
		log.Info().Msg("using gupshup notification provider")
	default:
		notifProvider = notifprovider.NewLogProvider()
		if cfg.App.Env == "production" {
			log.Warn().Msg("using log notification provider in production — set NOTIFICATION_PROVIDER=gupshup for real delivery")
		}
	}

	// Razorpay client (nil in dev if no keys configured)
	var razorpayClient *saasrazorpay.Client
	if cfg.Razorpay.KeyID != "" {
		razorpayClient = saasrazorpay.NewClient(saasrazorpay.Config{
			KeyID:         cfg.Razorpay.KeyID,
			KeySecret:     cfg.Razorpay.KeySecret,
			WebhookSecret: cfg.Razorpay.WebhookSecret,
		})
	}

	// Services
	authSvc := authservice.New(authRepo, cfg.JWT)
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
	saasSvc := saasservice.New(saasRepo, razorpayClient)
	opsSvc := opsservice.New(db)

	// Handlers
	authHandler := handler.New(authSvc, log)
	propHandler := prophandler.New(propSvc, log)
	guestHandler := guesthandler.New(guestSvc, log)
	resHandler := reshandler.New(resSvc, log)
	calendarHandler := calendarhandler.New(calendarSvc, log)
	checkinoutHandler := checkinouthandler.New(checkinoutSvc, log)
	billingHandler := billinghandler.New(billingSvc, log)
	hkHandler := hkhandler.New(hkSvc, log)
	laundryHandler := laundryhandler.New(laundrySvc, log)
	dashboardHandler := dashboardhandler.New(dashboardSvc, log)
	notifHandler := notifhandler.New(notifSvc, log)
	saasHandler := saashandler.New(saasSvc, razorpayClient, log)
	opsHandler := opshandler.New(opsSvc, log)

	// Middleware
	authMw := authmiddleware.New(authSvc, log)

	// Rate limiters
	authLimiter := middleware.NewRateLimiter(50, 1*time.Minute)     // 50 req/min for auth
	globalLimiter := middleware.NewRateLimiter(1000, 1*time.Minute) // 1000 req/min global

	// Router
	r := chi.NewRouter()
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(middleware.RequestLogger(log))
	r.Use(chimiddleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CORS.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Tenant-ID"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	r.Use(globalLimiter.Limit)
	r.Use(middleware.MetricsMiddleware)
	r.Use(middleware.MaxBodySize(1 << 20)) // 1MB global limit

	// Health checks
	healthHandler := middleware.NewHealthHandler(func() error {
		return db.Ping(context.Background())
	})

	// Health check
	r.Get("/health", healthHandler.Health)
	r.Get("/ready", healthHandler.Ready)
	r.Get("/metrics", middleware.MetricsHandler)

	// API v1
	r.Route("/api/v1", func(r chi.Router) {
		// Public routes
		r.Route("/auth", func(r chi.Router) {
			r.Use(authLimiter.Limit)
			r.Post("/register", authHandler.RegisterTenant)
			r.Post("/login", authHandler.Login)
			r.Post("/refresh", authHandler.RefreshToken)
			r.Post("/password-reset/request", authHandler.RequestPasswordReset)
			r.Post("/password-reset/confirm", authHandler.ConfirmPasswordReset)
		})

		// Public: plans listing and webhooks
		r.Get("/plans", saasHandler.ListPlans)

		// Webhook (public, no auth)
		r.Post("/webhooks/notifications", notifHandler.Webhook)
		r.Post("/webhooks/razorpay", saasHandler.RazorpayWebhook)

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(authMw.Authenticate)
			r.Use(authmiddleware.TenantContext)

			r.Post("/auth/logout", authHandler.Logout)

			// User management
			r.Route("/users", func(r chi.Router) {
				r.Use(authmiddleware.RequireRole("super_admin", "property_admin"))
				r.Post("/", authHandler.CreateUser)
			})

			// Properties
			r.Route("/properties", func(r chi.Router) {
				r.Use(authmiddleware.RequireRole("super_admin", "property_admin"))
				r.Post("/", propHandler.CreateProperty)
				r.Get("/", propHandler.ListProperties)
				r.Get("/{propertyID}", propHandler.GetProperty)
				r.Put("/{propertyID}", propHandler.UpdateProperty)

				// Unit Types
				r.Post("/{propertyID}/unit-types", propHandler.CreateUnitType)
				r.Get("/{propertyID}/unit-types", propHandler.ListUnitTypes)

				// Units
				r.Post("/{propertyID}/units", propHandler.CreateUnit)
				r.Get("/{propertyID}/units", propHandler.ListUnits)
				r.Put("/{propertyID}/units/{unitID}", propHandler.UpdateUnit)
				r.Delete("/{propertyID}/units/{unitID}", propHandler.DeleteUnit)
				r.Patch("/{propertyID}/units/{unitID}/status", propHandler.ChangeUnitStatus)
				r.Get("/{propertyID}/units/search", propHandler.SearchUnits)
			})

			// Guests
			r.Route("/guests", func(r chi.Router) {
				r.Post("/", guestHandler.CreateGuest)
				r.Get("/", guestHandler.ListGuests)
				r.Get("/{guestID}", guestHandler.GetGuest)
				r.Put("/{guestID}", guestHandler.UpdateGuest)
				r.Get("/{guestID}/history", guestHandler.GetGuestHistory)
				r.Get("/search", guestHandler.SearchGuests)
				r.With(middleware.MaxBodySize(10<<20)).Post("/{guestID}/documents", guestHandler.UploadDocument)
				r.Get("/{guestID}/documents", guestHandler.ListDocuments)
			})

			// Reservations
			r.Route("/reservations", func(r chi.Router) {
				r.Post("/", resHandler.CreateReservation)
				r.Get("/", resHandler.ListReservations)
				r.Get("/{reservationID}", resHandler.GetReservation)
				r.Put("/{reservationID}", resHandler.UpdateReservation)
				r.Post("/{reservationID}/cancel", resHandler.CancelReservation)
				r.Post("/{reservationID}/confirm", resHandler.ConfirmReservation)
				r.Post("/{reservationID}/check-in", resHandler.CheckIn)
				r.Post("/{reservationID}/check-out", resHandler.CheckOut)
				r.Get("/availability", resHandler.CheckAvailability)
			})

			// Calendar
			r.Route("/calendar", func(r chi.Router) {
				r.Get("/{propertyID}", calendarHandler.GetCalendarView)
				r.Post("/move", calendarHandler.MoveBooking)
				r.Get("/{propertyID}/occupancy", calendarHandler.GetOccupancyStats)
			})

			// Operations (Check-in/Check-out)
			r.Route("/operations", func(r chi.Router) {
				r.Post("/check-in", checkinoutHandler.CheckIn)
				r.Post("/check-out", checkinoutHandler.CheckOut)
				r.Post("/walk-in", checkinoutHandler.WalkIn)
				r.Get("/pre-checkout", checkinoutHandler.GetPreCheckoutSummary)
				r.Post("/no-show", opsHandler.MarkNoShow)
				r.Post("/extend-stay", opsHandler.ExtendStay)
				r.Post("/room-move", opsHandler.MoveRoom)
				r.Post("/maintenance-block", opsHandler.CreateMaintenanceBlock)
				r.Post("/refund-deposit", opsHandler.RefundDeposit)
			})

			// Billing
			r.Route("/billing", func(r chi.Router) {
				// Folios
				r.Get("/folios/{folioID}", billingHandler.GetFolio)
				r.Get("/folios/{folioID}/summary", billingHandler.GetFolioSummary)
				r.Get("/folios/reservation/{reservationID}", billingHandler.GetFolioByReservation)
				r.Get("/folios/{folioID}/items", billingHandler.ListLineItems)
				r.Get("/folios/{folioID}/payments", billingHandler.ListPayments)

				// Charges
				r.Post("/charges", billingHandler.AddCharge)
				r.Post("/charges/void", billingHandler.VoidLineItem)

				// Payments
				r.Post("/payments", billingHandler.RecordPayment)

				// Invoices
				r.Get("/invoices", billingHandler.ListInvoices)
				r.Get("/invoices/{invoiceID}", billingHandler.GetInvoice)
				r.Post("/invoices/{invoiceID}/pdf", billingHandler.GenerateInvoicePDF)
				r.Get("/invoices/{invoiceID}/pdf", billingHandler.GetInvoicePDF)
			})

			// Housekeeping
			r.Route("/housekeeping", func(r chi.Router) {
				r.Post("/tasks", hkHandler.CreateTask)
				r.Get("/tasks", hkHandler.ListTasks)
				r.Get("/tasks/{taskID}", hkHandler.GetTask)
				r.Post("/tasks/assign", hkHandler.AssignTask)
				r.Post("/tasks/status", hkHandler.UpdateStatus)
				r.Get("/stats/{propertyID}", hkHandler.GetStats)
			})

			// Laundry
			r.Route("/laundry", func(r chi.Router) {
				r.Post("/orders", laundryHandler.CreateOrder)
				r.Get("/orders", laundryHandler.ListOrders)
				r.Get("/orders/{orderID}", laundryHandler.GetOrder)
				r.Post("/orders/status", laundryHandler.UpdateStatus)
				r.Post("/orders/{orderID}/post-to-folio", laundryHandler.PostToFolio)
				r.Get("/stats/{propertyID}", laundryHandler.GetStats)
				// Rate Card management
				r.Post("/rate-card", laundryHandler.CreateRateCard)
				r.Get("/rate-card/{propertyID}", laundryHandler.ListRateCards)
				r.Put("/rate-card", laundryHandler.UpdateRateCard)
			})

			// Dashboard
			r.Route("/dashboard", func(r chi.Router) {
				r.Get("/{propertyID}", dashboardHandler.GetDashboard)
				r.Get("/{propertyID}/revenue-trend", dashboardHandler.GetRevenueTrend)
				r.Get("/{propertyID}/daily-collection", dashboardHandler.GetDailyCollection)
				r.Get("/{propertyID}/outstanding-dues", dashboardHandler.GetOutstandingDues)
				r.Get("/{propertyID}/end-of-day", dashboardHandler.GetEndOfDaySummary)
				r.Post("/{propertyID}/close-day", dashboardHandler.CloseDay)
			})

			// Notifications
			r.Route("/notifications", func(r chi.Router) {
				r.Post("/templates", notifHandler.UpsertTemplate)
				r.Get("/templates", notifHandler.ListTemplates)
				r.Post("/send-test", notifHandler.SendTest)
				r.Get("/logs", notifHandler.ListLogs)
			})

			// Subscription management
			r.Route("/subscription", func(r chi.Router) {
				r.Get("/", saasHandler.GetSubscription)
				r.Post("/", saasHandler.CreateSubscription)
				r.Put("/plan", saasHandler.ChangePlan)
				r.Post("/cancel", saasHandler.CancelSubscription)
				r.Post("/checkout", saasHandler.CreateCheckout)
				r.Post("/verify-payment", saasHandler.VerifyPayment)
				r.Get("/billing-events", saasHandler.ListBillingEvents)
			})

			// Onboarding
			r.Route("/onboarding", func(r chi.Router) {
				r.Post("/{propertyID}/init", saasHandler.InitOnboarding)
				r.Get("/{propertyID}/steps", saasHandler.GetOnboardingSteps)
				r.Post("/complete-step", saasHandler.CompleteStep)
			})

			// Admin (platform admin only)
			r.Route("/admin", func(r chi.Router) {
				r.Use(authMw.RequirePlatformAdmin)
				r.Get("/tenants", saasHandler.AdminListTenants)
				r.Get("/metrics", saasHandler.AdminGetMetrics)
				r.Get("/plans", saasHandler.AdminListPlans)
			})
		})
	})

	srv := &http.Server{
		Addr:         ":" + cfg.App.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info().Str("port", cfg.App.Port).Msg("server listening")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server error")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down server")
	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatal().Err(err).Msg("server forced to shutdown")
	}

	log.Info().Msg("server stopped")
}
