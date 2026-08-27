package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"

	"github.com/eprisi/absensi-next/services/api/internal/config"
	"github.com/eprisi/absensi-next/services/api/internal/handler"
	appmiddleware "github.com/eprisi/absensi-next/services/api/internal/middleware"
	"github.com/eprisi/absensi-next/services/api/internal/platform/authtoken"
	"github.com/eprisi/absensi-next/services/api/internal/platform/db"
	"github.com/eprisi/absensi-next/services/api/internal/platform/storage"
	"github.com/eprisi/absensi-next/services/api/internal/repo"
	"github.com/eprisi/absensi-next/services/api/internal/usecase/attendance"
	"github.com/eprisi/absensi-next/services/api/internal/usecase/auth"
	configusecase "github.com/eprisi/absensi-next/services/api/internal/usecase/config"
	"github.com/eprisi/absensi-next/services/api/internal/usecase/department"
	"github.com/eprisi/absensi-next/services/api/internal/usecase/employee"
	"github.com/eprisi/absensi-next/services/api/internal/usecase/leave"
	"github.com/eprisi/absensi-next/services/api/internal/usecase/monitoring"
	"github.com/eprisi/absensi-next/services/api/internal/usecase/notification"
	"github.com/eprisi/absensi-next/services/api/internal/usecase/profile"
	"github.com/eprisi/absensi-next/services/api/internal/usecase/recap"
	"github.com/eprisi/absensi-next/services/api/internal/usecase/task"
)

func main() {
	_ = godotenv.Load() // optional, ignored if .env is absent (e.g. in containers using real env vars)

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	if err := db.Migrate(cfg.DatabaseURL, "migrations"); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	gdb, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	sqlDB, err := gdb.DB()
	if err != nil {
		log.Fatalf("db handle: %v", err)
	}
	defer sqlDB.Close()

	issuer := authtoken.NewIssuer(cfg.JWTSecret)
	employeeRepo := repo.NewEmployeeRepo(gdb)
	userRepo := repo.NewUserRepo(gdb)
	revokedTokenRepo := repo.NewRevokedTokenRepo(gdb)
	authService := auth.NewService(employeeRepo, userRepo, issuer, revokedTokenRepo, cfg.AccessTokenTTL, cfg.RefreshTokenTTL)
	authHandler := handler.NewAuthHandler(authService)

	officeLocationRepo := repo.NewOfficeLocationRepo(gdb)
	fieldAssignmentRepo := repo.NewFieldAssignmentRepo(gdb)
	shiftScheduleRepo := repo.NewShiftScheduleRepo(gdb)
	attendanceService := attendance.NewService(gdb, officeLocationRepo, fieldAssignmentRepo, shiftScheduleRepo)
	photoStore := storage.NewLocalStore(cfg.StorageDir)
	attendanceHandler := handler.NewAttendanceHandler(attendanceService, photoStore)

	// D-17: periodically flag attendance rows left open (no checkout) well
	// past their shift. In-process ticker for dev/demo visibility -- a real
	// scheduler (external cron, etc.) is a Phase 5 decision, not made yet.
	go runFlagNoCheckoutJob(attendanceService, cfg.FlagJobInterval, cfg.FlagJobStaleAfter)

	notificationRepo := repo.NewNotificationRepo(gdb)
	notificationPreferenceRepo := repo.NewNotificationPreferenceRepo(gdb)
	notificationService := notification.NewService(notificationRepo, notificationPreferenceRepo)
	notificationHandler := handler.NewNotificationHandler(notificationService)

	profileService := profile.NewService(employeeRepo, userRepo, photoStore)
	profileHandler := handler.NewProfileHandler(profileService)

	leaveRepo := repo.NewLeaveRequestRepo(gdb)
	leaveService := leave.NewService(leaveRepo)
	leaveHandler := handler.NewLeaveHandler(leaveService, notificationService)

	taskRepo := repo.NewTaskRepo(gdb)
	taskService := task.NewService(taskRepo)
	taskHandler := handler.NewTaskHandler(taskService)

	shiftRepo := repo.NewShiftRepo(gdb)
	weeklyShiftDefaultRepo := repo.NewWeeklyShiftDefaultRepo(gdb)
	workScheduleRepo := repo.NewWorkScheduleRepo(gdb)
	companySettingsRepo := repo.NewCompanySettingsRepo(gdb)
	configService := configusecase.NewService(shiftRepo, weeklyShiftDefaultRepo, workScheduleRepo, officeLocationRepo, fieldAssignmentRepo, companySettingsRepo)
	configHandler := handler.NewConfigHandler(configService, photoStore)

	departmentRepo := repo.NewDepartmentRepo(gdb)
	departmentService := department.NewService(departmentRepo)
	departmentHandler := handler.NewDepartmentHandler(departmentService)

	employeeService := employee.NewService(employeeRepo)
	employeeHandler := handler.NewEmployeeHandler(employeeService)

	attendanceRepo := repo.NewAttendanceRepo(gdb)
	monitoringService := monitoring.NewService(attendanceRepo, photoStore)
	monitoringHandler := handler.NewMonitoringHandler(monitoringService)

	recapService := recap.NewService(employeeRepo, attendanceRepo, leaveRepo, shiftScheduleRepo)
	recapHandler := handler.NewRecapHandler(recapService)

	// D-4: rate limit login attempts by client IP. See middleware.LoginRateLimiter
	// doc comment re: in-memory/single-instance limitation.
	loginLimiter := appmiddleware.NewLoginRateLimiter(cfg.LoginRateLimitPerMinute, time.Minute)
	rateLimitByIP := loginLimiter.Middleware(func(req *http.Request) string { return req.RemoteAddr })

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	// Dev-only permissive CORS so the Next.js dev server (localhost:3000) can
	// call the API directly from the browser. Tighten this before production.
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	healthHandler := handler.NewHealthHandler(sqlDB)
	r.Get("/health", healthHandler.Health)

	r.Route("/api/v1", func(api chi.Router) {
		api.Group(func(pub chi.Router) {
			pub.Use(rateLimitByIP)
			pub.Post("/auth/employee/login", authHandler.EmployeeLogin)
			pub.Post("/auth/admin/login", authHandler.AdminLogin)
		})
		api.Post("/auth/refresh", authHandler.Refresh)
		api.Post("/auth/logout", authHandler.Logout)

		api.Group(func(employeeRoutes chi.Router) {
			employeeRoutes.Use(appmiddleware.RequireAuth(issuer, authtoken.AudienceEmployee))
			employeeRoutes.Post("/attendance/check-in", attendanceHandler.CheckIn)
			employeeRoutes.Get("/attendance/today", attendanceHandler.Today)
			employeeRoutes.Get("/attendance/geofence", attendanceHandler.Geofence)

			employeeRoutes.Post("/leave-requests", leaveHandler.Submit)
			employeeRoutes.Get("/leave-requests", leaveHandler.ListOwn)

			employeeRoutes.Post("/tasks", taskHandler.Create)
			employeeRoutes.Get("/tasks", taskHandler.ListOwn)
			employeeRoutes.Put("/tasks/{id}", taskHandler.Update)
		})

		// Shared by both audiences -- the handlers branch on claims.Audience
		// themselves (see handler.identity()), so this uses RequireAnyAuth
		// instead of duplicating routes under employeeRoutes/adminRoutes
		// (registering the identical path twice on chi's shared mux would
		// silently let the second registration win for both callers).
		api.Group(func(selfRoutes chi.Router) {
			selfRoutes.Use(appmiddleware.RequireAnyAuth(issuer, authtoken.AudienceEmployee, authtoken.AudienceAdmin))

			selfRoutes.Get("/me", profileHandler.Get)
			selfRoutes.Patch("/me", profileHandler.Update)
			selfRoutes.Post("/me/avatar", profileHandler.UploadAvatar)
			selfRoutes.Get("/me/avatar", profileHandler.Avatar)
			selfRoutes.Post("/me/change-password", profileHandler.ChangePassword)

			selfRoutes.Get("/notifications", notificationHandler.List)
			selfRoutes.Get("/notifications/unread-count", notificationHandler.UnreadCount)
			selfRoutes.Patch("/notifications/{id}/read", notificationHandler.MarkRead)
			selfRoutes.Patch("/notifications/read-all", notificationHandler.MarkAllRead)
			selfRoutes.Get("/notification-preferences", notificationHandler.Preferences)
			selfRoutes.Put("/notification-preferences", notificationHandler.SetPreference)
		})

		api.Group(func(adminRoutes chi.Router) {
			adminRoutes.Use(appmiddleware.RequireAuth(issuer, authtoken.AudienceAdmin))
			adminRoutes.Get("/admin/leave-requests", leaveHandler.ListForAdmin)
			adminRoutes.Post("/admin/leave-requests/{id}/review", leaveHandler.Review)

			adminRoutes.Get("/admin/departments", departmentHandler.List)
			adminRoutes.Post("/admin/departments", departmentHandler.Create)

			adminRoutes.Get("/admin/employees", employeeHandler.List)
			adminRoutes.Post("/admin/employees", employeeHandler.Create)
			adminRoutes.Put("/admin/employees/{id}", employeeHandler.Update)

			adminRoutes.Get("/admin/attendance/monitoring", monitoringHandler.List)
			adminRoutes.Get("/admin/attendance/{id}/photo", monitoringHandler.Photo)

			adminRoutes.Get("/admin/reports/recap", recapHandler.Get)
			adminRoutes.Get("/admin/reports/recap/export", recapHandler.Export)

			// Readable by any admin; RBAC per docs/openapi.yaml (D-7).
			adminRoutes.Get("/admin/config/company", configHandler.GetCompanySettings)
			adminRoutes.Get("/admin/config/company/logo", configHandler.CompanyLogo)

			adminRoutes.Get("/admin/config/shifts", configHandler.ListShifts)
			adminRoutes.Get("/admin/config/office-locations", configHandler.ListOfficeLocations)
			adminRoutes.Get("/admin/config/field-assignments", configHandler.ListFieldAssignments)
			adminRoutes.Post("/admin/config/work-schedules", configHandler.SetWorkSchedule)
			adminRoutes.Post("/admin/config/weekly-shift-defaults", configHandler.SetWeeklyShiftDefault)
			adminRoutes.Post("/admin/config/field-assignments", configHandler.CreateFieldAssignment)

			// Mutating shift/office-location config is superadmin-only,
			// per docs/openapi.yaml (D-7 RBAC) -- these directly control
			// lateness thresholds and the enforced geofence radius (D-1/D-10).
			adminRoutes.Group(func(superadminRoutes chi.Router) {
				superadminRoutes.Use(appmiddleware.RequireRole("superadmin"))
				superadminRoutes.Post("/admin/config/shifts", configHandler.CreateShift)
				superadminRoutes.Put("/admin/config/shifts/{id}", configHandler.UpdateShift)
				superadminRoutes.Post("/admin/config/office-locations", configHandler.CreateOfficeLocation)
				superadminRoutes.Put("/admin/config/office-locations/{id}", configHandler.UpdateOfficeLocation)
				superadminRoutes.Put("/admin/config/company", configHandler.UpdateCompanySettings)
				superadminRoutes.Post("/admin/config/company/logo", configHandler.UploadCompanyLogo)
			})
		})

		// dashboard and reports routes are mounted here in subsequent Phase 2
		// work, per docs/openapi.yaml -- response shapes not yet finalized
		// (see docs/PROGRESS.md Phase 1 blockers).
	})

	addr := ":" + cfg.Port
	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func runFlagNoCheckoutJob(svc attendance.Service, interval, staleAfter time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Printf("flag-no-checkout job: running every %s, threshold %s", interval, staleAfter)

	for range ticker.C {
		n, err := svc.FlagStaleOpenAttendances(context.Background(), staleAfter)
		if err != nil {
			log.Printf("flag-no-checkout job: error: %v", err)
			continue
		}
		if n > 0 {
			log.Printf("flag-no-checkout job: flagged %d attendance row(s)", n)
		}
	}
}
