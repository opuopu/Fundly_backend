package http

import (
	"log"
	"net/http"
	"os"
	"time"

	accountapp "github.com/Dodul01/fundly/internal/application/account"
	attachmentapp "github.com/Dodul01/fundly/internal/application/attachment"
	"github.com/Dodul01/fundly/internal/application/auth"
	budgetapp "github.com/Dodul01/fundly/internal/application/budget"
	categoryapp "github.com/Dodul01/fundly/internal/application/category"
	configapp "github.com/Dodul01/fundly/internal/application/config"
	feedbackapp "github.com/Dodul01/fundly/internal/application/feedback"
	reportapp "github.com/Dodul01/fundly/internal/application/report"
	statapp "github.com/Dodul01/fundly/internal/application/stat"
	termsapp "github.com/Dodul01/fundly/internal/application/terms"
	transactionapp "github.com/Dodul01/fundly/internal/application/transaction"
	"github.com/Dodul01/fundly/internal/application/transfer"
	userapp "github.com/Dodul01/fundly/internal/application/user"
	"github.com/Dodul01/fundly/internal/config"
	"github.com/Dodul01/fundly/internal/infrastructure/email"
	"github.com/Dodul01/fundly/internal/infrastructure/persistence/postgres"
	"github.com/Dodul01/fundly/internal/infrastructure/security"
	"github.com/Dodul01/fundly/internal/infrastructure/storage"
	"github.com/jackc/pgx/v5/pgxpool"
)

// BuildRouter constructs and registers all 46 API routes with their dependencies.
func BuildRouter(pool *pgxpool.Pool, cfg *config.Config) http.Handler {
	var authHandler *AuthHandler
	var userHandler *UserHandler
	var accountHandler *AccountHandler
	var categoryHandler *CategoryHandler
	var transactionHandler *TransactionHandler
	var budgetHandler *BudgetHandler
	var statHandler *StatHandler
	var reportHandler *ReportHandler
	var attachmentHandler *AttachmentHandler
	var configHandler *ConfigHandler
	var feedbackHandler *FeedbackHandler
	var termsHandler *TermsHandler
	var authMiddleware *AuthMiddleware

	if pool != nil {
		userRepo := postgres.NewUserRepository(pool)
		accRepo := postgres.NewAccountRepository(pool)
		txRepo := postgres.NewTransactionRepository(pool)
		txManager := postgres.NewTxManager(pool)
		transferService := transfer.NewService(accRepo, txRepo, txManager)

		hasher := security.NewBcryptHasher()
		otpGen := security.NewCryptoOTPGenerator()
		jwtSecret := os.Getenv("JWT_SECRET")
		if jwtSecret == "" {
			jwtSecret = "fundly-default-development-secret-key-32chars"
		}
		jwtService := security.NewJWTService(jwtSecret, 24*time.Hour, 7*24*time.Hour)
		socialService := security.NewSocialAuthService()
		blacklist := security.NewMemoryTokenBlacklist()

		var mailer auth.Mailer
		if cfg.SMTPConfigured() {
			mailer = email.NewSMTPMailer(email.SMTPConfig{
				Host:     cfg.SMTPHost,
				Port:     cfg.SMTPPort,
				Username: cfg.SMTPUsername,
				Password: cfg.SMTPPassword,
				From:     cfg.SMTPFrom,
				FromName: cfg.SMTPFromName,
			})
			log.Printf("✉️  SMTP mailer enabled (%s)", cfg.SMTPHost)
		} else {
			mailer = email.NewLogMailer()
			log.Println("✉️  SMTP not configured; using dev log mailer (codes printed to server log)")
		}

		registerUC := auth.NewRegisterUseCase(userRepo, hasher, otpGen, mailer)
		verifyOTPUC := auth.NewVerifyOTPUseCase(userRepo, jwtService)
		resendOTPUC := auth.NewResendOTPUseCase(userRepo, otpGen, mailer)
		loginUC := auth.NewLoginUseCase(userRepo, hasher, jwtService)
		socialLoginUC := auth.NewSocialLoginUseCase(userRepo, socialService, jwtService)
		refreshTokenUC := auth.NewRefreshTokenUseCase(userRepo, jwtService, jwtService)
		forgotPasswordUC := auth.NewForgotPasswordUseCase(userRepo, otpGen, mailer)
		resetPasswordUC := auth.NewResetPasswordUseCase(userRepo, hasher)
		logoutUC := auth.NewLogoutUseCase(blacklist)
		authHandler = NewAuthHandler(registerUC, verifyOTPUC, resendOTPUC, loginUC, socialLoginUC, refreshTokenUC, forgotPasswordUC, resetPasswordUC, logoutUC)
		authMiddleware = NewAuthMiddleware(jwtService, blacklist)

		getProfileUC := userapp.NewGetProfileUseCase(userRepo)
		updateProfileUC := userapp.NewUpdateProfileUseCase(userRepo)
		changePasswordUC := userapp.NewChangePasswordUseCase(userRepo, hasher)
		fileStorage := storage.NewLocalFileStorage(cfg.UploadsDir, "/uploads")
		uploadAvatarUC := userapp.NewUploadAvatarUseCase(userRepo, fileStorage)
		userHandler = NewUserHandler(getProfileUC, updateProfileUC, changePasswordUC, uploadAvatarUC)

		listAccountsUC := accountapp.NewListAccountsUseCase(accRepo)
		createAccountUC := accountapp.NewCreateAccountUseCase(accRepo)
		getAccountUC := accountapp.NewGetAccountUseCase(accRepo)
		updateAccountUC := accountapp.NewUpdateAccountUseCase(accRepo)
		deleteAccountUC := accountapp.NewDeleteAccountUseCase(accRepo)
		getSummaryUC := accountapp.NewGetAccountsSummaryUseCase(accRepo)
		accountHandler = NewAccountHandler(listAccountsUC, createAccountUC, getAccountUC, updateAccountUC, deleteAccountUC, getSummaryUC)

		categoryRepo := postgres.NewCategoryRepository(pool)
		listCategoriesUC := categoryapp.NewListCategoriesUseCase(categoryRepo)
		createCategoryUC := categoryapp.NewCreateCategoryUseCase(categoryRepo)
		updateCategoryUC := categoryapp.NewUpdateCategoryUseCase(categoryRepo)
		deleteCategoryUC := categoryapp.NewDeleteCategoryUseCase(categoryRepo)
		categoryHandler = NewCategoryHandler(listCategoriesUC, createCategoryUC, updateCategoryUC, deleteCategoryUC)

		configRepo := postgres.NewConfigRepository(pool)

		attachmentRepo := postgres.NewAttachmentRepository(pool)
		createTxUC := transactionapp.NewCreateTransactionUseCase(txRepo, accRepo, categoryRepo, attachmentRepo, txManager)
		listTxUC := transactionapp.NewListTransactionsUseCase(txRepo, accRepo, categoryRepo, attachmentRepo)
		getTxUC := transactionapp.NewGetTransactionUseCase(txRepo, accRepo, categoryRepo, attachmentRepo)
		updateTxUC := transactionapp.NewUpdateTransactionUseCase(txRepo, accRepo, categoryRepo, attachmentRepo, txManager)
		deleteTxUC := transactionapp.NewDeleteTransactionUseCase(txRepo, accRepo, txManager)
		searchTxUC := transactionapp.NewSearchTransactionsUseCase(txRepo, accRepo, categoryRepo, attachmentRepo)
		transferTxUC := transactionapp.NewTransferUseCase(transferService, accRepo, categoryRepo)
		transactionHandler = NewTransactionHandler(listTxUC, createTxUC, getTxUC, updateTxUC, deleteTxUC, searchTxUC, transferTxUC)

		budgetRepo := postgres.NewBudgetRepository(pool)
		createBudgetUC := budgetapp.NewCreateBudgetUseCase(budgetRepo, categoryRepo, txRepo)
		listBudgetsUC := budgetapp.NewListBudgetsUseCase(budgetRepo, categoryRepo, txRepo)
		getBudgetUC := budgetapp.NewGetBudgetUseCase(budgetRepo, categoryRepo, txRepo)
		updateBudgetUC := budgetapp.NewUpdateBudgetUseCase(budgetRepo, categoryRepo, txRepo)
		deleteBudgetUC := budgetapp.NewDeleteBudgetUseCase(budgetRepo)
		budgetHandler = NewBudgetHandler(listBudgetsUC, createBudgetUC, getBudgetUC, updateBudgetUC, deleteBudgetUC)

		statRepo := postgres.NewStatRepository(pool)
		breakdownUC := statapp.NewBreakdownUseCase(statRepo, categoryRepo, configRepo)
		overviewUC := statapp.NewOverviewUseCase(statRepo, budgetRepo, configRepo)
		statHandler = NewStatHandler(breakdownUC, overviewUC)

		compareUC := reportapp.NewCompareUseCase(statRepo, categoryRepo, configRepo)
		reportHandler = NewReportHandler(compareUC)

		fileStorageForAttachments := storage.NewLocalFileStorage(cfg.UploadsDir, "/uploads")
		uploadAttachmentUC := attachmentapp.NewUploadAttachmentUseCase(attachmentRepo, fileStorageForAttachments)
		getAttachmentUC := attachmentapp.NewGetAttachmentUseCase(attachmentRepo)
		deleteAttachmentUC := attachmentapp.NewDeleteAttachmentUseCase(attachmentRepo)
		attachmentHandler = NewAttachmentHandler(uploadAttachmentUC, getAttachmentUC, deleteAttachmentUC)

		getConfigUC := configapp.NewGetConfigUseCase(configRepo)
		updateConfigUC := configapp.NewUpdateConfigUseCase(configRepo)
		configHandler = NewConfigHandler(getConfigUC, updateConfigUC)

		feedbackRepo := postgres.NewFeedbackRepository(pool)
		submitFeedbackUC := feedbackapp.NewSubmitFeedbackUseCase(feedbackRepo)
		feedbackHandler = NewFeedbackHandler(submitFeedbackUC)

		termsRepo := postgres.NewTermsRepository(pool)
		getLatestTermsUC := termsapp.NewGetLatestUseCase(termsRepo)
		acknowledgeTermsUC := termsapp.NewAcknowledgeUseCase(termsRepo)
		termsHandler = NewTermsHandler(getLatestTermsUC, acknowledgeTermsUC)
	}

	handler := NewHandler(pool)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", handler.HealthCheck)
	mux.HandleFunc("/api/v1/health", handler.HealthCheck)

	// Serve static uploads if directory exists
	if cfg.UploadsDir != "" {
		mux.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir(cfg.UploadsDir))))
	}

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
		mux.HandleFunc("/api/v1/users/me/avatar", authMiddleware.RequireAuth(userHandler.UploadAvatar))
	}

	if accountHandler != nil && authMiddleware != nil {
		mux.HandleFunc("/api/v1/accounts", authMiddleware.RequireAuth(accountHandler.HandleAccounts))
		mux.HandleFunc("/api/v1/accounts/summary", authMiddleware.RequireAuth(accountHandler.GetSummary))
		mux.HandleFunc("/api/v1/accounts/{id}", authMiddleware.RequireAuth(accountHandler.HandleAccountByID))
	}

	if categoryHandler != nil && authMiddleware != nil {
		mux.HandleFunc("/api/v1/categories", authMiddleware.RequireAuth(categoryHandler.HandleCategories))
		mux.HandleFunc("/api/v1/categories/{id}", authMiddleware.RequireAuth(categoryHandler.HandleCategoryByID))
	}

	if transactionHandler != nil && authMiddleware != nil {
		mux.HandleFunc("/api/v1/transactions/search", authMiddleware.RequireAuth(transactionHandler.Search))
		mux.HandleFunc("/api/v1/transactions/transfer", authMiddleware.RequireAuth(transactionHandler.Transfer))
		mux.HandleFunc("/api/v1/transactions", authMiddleware.RequireAuth(transactionHandler.HandleTransactions))
		mux.HandleFunc("/api/v1/transactions/{id}", authMiddleware.RequireAuth(transactionHandler.HandleTransactionByID))
	}

	if budgetHandler != nil && authMiddleware != nil {
		mux.HandleFunc("/api/v1/budgets", authMiddleware.RequireAuth(budgetHandler.HandleBudgets))
		mux.HandleFunc("/api/v1/budgets/{id}", authMiddleware.RequireAuth(budgetHandler.HandleBudgetByID))
	}

	if statHandler != nil && authMiddleware != nil {
		mux.HandleFunc("/api/v1/stats/breakdown", authMiddleware.RequireAuth(statHandler.Breakdown))
		mux.HandleFunc("/api/v1/stats/overview", authMiddleware.RequireAuth(statHandler.Overview))
	}

	if reportHandler != nil && authMiddleware != nil {
		mux.HandleFunc("/api/v1/reports/compare", authMiddleware.RequireAuth(reportHandler.Compare))
	}

	if attachmentHandler != nil && authMiddleware != nil {
		mux.HandleFunc("/api/v1/attachments", authMiddleware.RequireAuth(attachmentHandler.HandleAttachments))
		mux.HandleFunc("/api/v1/attachments/{id}", authMiddleware.RequireAuth(attachmentHandler.HandleAttachmentByID))
	}

	if configHandler != nil && authMiddleware != nil {
		mux.HandleFunc("/api/v1/configuration", authMiddleware.RequireAuth(configHandler.HandleConfiguration))
	}

	if feedbackHandler != nil && authMiddleware != nil {
		mux.HandleFunc("/api/v1/feedback", authMiddleware.RequireAuth(feedbackHandler.SubmitFeedback))
	}

	if termsHandler != nil && authMiddleware != nil {
		mux.HandleFunc("/api/v1/terms/latest", termsHandler.GetLatest)
		mux.HandleFunc("/api/v1/terms/acknowledge", authMiddleware.RequireAuth(termsHandler.Acknowledge))
	}

	return mux
}
