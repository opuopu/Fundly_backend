package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Dodul01/fundly/internal/application/auth"
	"github.com/Dodul01/fundly/internal/application/transfer"
	userapp "github.com/Dodul01/fundly/internal/application/user"
	"github.com/Dodul01/fundly/internal/config"
	infrahttp "github.com/Dodul01/fundly/internal/infrastructure/http"
	"github.com/Dodul01/fundly/internal/infrastructure/persistence/postgres"
	"github.com/Dodul01/fundly/internal/infrastructure/security"
)

func main() {
	cfg := config.Load()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. Initialize PostgreSQL Connection Pool
	dbCfg := postgres.DefaultConfig(cfg.DatabaseURL)
	pool, err := postgres.NewPool(ctx, dbCfg)
	if err != nil {
		log.Printf("⚠️ PostgreSQL connection failed: %v", err)
		log.Printf("💡 To start PostgreSQL: run 'docker compose up -d' or configure DATABASE_URL in .env")
	} else {
		defer pool.Close()
		log.Println("✅ Connected to PostgreSQL database successfully.")

		// 2. Run Database Migrations if configured
		if cfg.AutoMigrate {
			if err := postgres.ApplyMigrations(ctx, pool); err != nil {
				log.Printf("⚠️ Failed to apply database migrations: %v", err)
			} else {
				log.Println("✅ Database migrations applied successfully.")
			}
		}
	}

	// 3. Initialize Repositories and Security Providers
	var authHandler *infrahttp.AuthHandler
	var userHandler *infrahttp.UserHandler
	var authMiddleware *infrahttp.AuthMiddleware
	if pool != nil {
		userRepo := postgres.NewUserRepository(pool)
		accRepo := postgres.NewAccountRepository(pool)
		txRepo := postgres.NewTransactionRepository(pool)
		_ = transfer.NewService(accRepo, txRepo)

		hasher := security.NewBcryptHasher()
		otpGen := security.NewCryptoOTPGenerator()
		jwtService := security.NewJWTService(os.Getenv("JWT_SECRET"), 24*time.Hour, 7*24*time.Hour)
		socialService := security.NewSocialAuthService()
		blacklist := security.NewMemoryTokenBlacklist()

		registerUC := auth.NewRegisterUseCase(userRepo, hasher, otpGen)
		verifyOTPUC := auth.NewVerifyOTPUseCase(userRepo, jwtService)
		resendOTPUC := auth.NewResendOTPUseCase(userRepo, otpGen)
		loginUC := auth.NewLoginUseCase(userRepo, hasher, jwtService)
		socialLoginUC := auth.NewSocialLoginUseCase(userRepo, socialService, jwtService)
		refreshTokenUC := auth.NewRefreshTokenUseCase(userRepo, jwtService, jwtService)
		forgotPasswordUC := auth.NewForgotPasswordUseCase(userRepo, otpGen)
		resetPasswordUC := auth.NewResetPasswordUseCase(userRepo, hasher)
		logoutUC := auth.NewLogoutUseCase(blacklist)
		authHandler = infrahttp.NewAuthHandler(registerUC, verifyOTPUC, resendOTPUC, loginUC, socialLoginUC, refreshTokenUC, forgotPasswordUC, resetPasswordUC, logoutUC)
		authMiddleware = infrahttp.NewAuthMiddleware(jwtService, blacklist)

		getProfileUC := userapp.NewGetProfileUseCase(userRepo)
		updateProfileUC := userapp.NewUpdateProfileUseCase(userRepo)
		changePasswordUC := userapp.NewChangePasswordUseCase(userRepo, hasher)
		userHandler = infrahttp.NewUserHandler(getProfileUC, updateProfileUC, changePasswordUC)
	}

	// 4. Initialize HTTP Handler & Server
	handler := infrahttp.NewHandler(pool)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", handler.HealthCheck)
	mux.HandleFunc("/api/v1/health", handler.HealthCheck)

	if authHandler != nil {
		mux.HandleFunc("/api/v1/auth/register", authHandler.Register)
		mux.HandleFunc("/api/v1/auth/verify-otp", authHandler.VerifyOTP)
		mux.HandleFunc("/api/v1/auth/resend-otp", authHandler.ResendOTP)
		mux.HandleFunc("/api/v1/auth/login", authHandler.Login)
		mux.HandleFunc("/api/v1/auth/social-login", authHandler.SocialLogin)
		mux.HandleFunc("/api/v1/auth/refresh", authHandler.Refresh)
		mux.HandleFunc("/api/v1/auth/forgot-password", authHandler.ForgotPassword)
		mux.HandleFunc("/api/v1/auth/reset-password", authHandler.ResetPassword)
		mux.HandleFunc("/api/v1/auth/logout", authHandler.Logout)
	}

	if userHandler != nil && authMiddleware != nil {
		mux.HandleFunc("/api/v1/users/me", authMiddleware.RequireAuth(userHandler.HandleMe))
		mux.HandleFunc("/api/v1/users/me/password", authMiddleware.RequireAuth(userHandler.ChangePassword))
	}

	serverAddr := ":" + cfg.Port
	server := &http.Server{
		Addr:    serverAddr,
		Handler: mux,
	}

	// Graceful shutdown handling
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		fmt.Printf("Fundly server listening on %s (env: %s)...\n", serverAddr, cfg.Environment)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed to start: %v", err)
		}
	}()

	<-stop
	log.Println("Shutting down server gracefully...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Error shutting down HTTP server: %v", err)
	}
	log.Println("Fundly server stopped.")
}