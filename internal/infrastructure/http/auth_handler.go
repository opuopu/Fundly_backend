package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/Dodul01/fundly/internal/application/auth"
	"github.com/Dodul01/fundly/internal/domain/user"
)

type AuthHandler struct {
	registerUC       *auth.RegisterUseCase
	verifyOTPUC      *auth.VerifyOTPUseCase
	resendOTPUC      *auth.ResendOTPUseCase
	loginUC          *auth.LoginUseCase
	socialLoginUC    *auth.SocialLoginUseCase
	refreshTokenUC   *auth.RefreshTokenUseCase
	forgotPasswordUC *auth.ForgotPasswordUseCase
	resetPasswordUC  *auth.ResetPasswordUseCase
	logoutUC         *auth.LogoutUseCase
}

func NewAuthHandler(
	registerUC *auth.RegisterUseCase,
	verifyOTPUC *auth.VerifyOTPUseCase,
	resendOTPUC *auth.ResendOTPUseCase,
	loginUC *auth.LoginUseCase,
	socialLoginUC *auth.SocialLoginUseCase,
	refreshTokenUC *auth.RefreshTokenUseCase,
	forgotPasswordUC *auth.ForgotPasswordUseCase,
	resetPasswordUC *auth.ResetPasswordUseCase,
	logoutUC *auth.LogoutUseCase,
) *AuthHandler {
	return &AuthHandler{
		registerUC:       registerUC,
		verifyOTPUC:      verifyOTPUC,
		resendOTPUC:      resendOTPUC,
		loginUC:          loginUC,
		socialLoginUC:    socialLoginUC,
		refreshTokenUC:   refreshTokenUC,
		forgotPasswordUC: forgotPasswordUC,
		resetPasswordUC:  resetPasswordUC,
		logoutUC:         logoutUC,
	}
}

type RegisterRequest struct {
	Email           string `json:"email"`
	FullName        string `json:"full_name,omitempty"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
}

type VerifyOTPRequest struct {
	Email string `json:"email"`
	OTP   string `json:"otp"`
}

type ResendOTPRequest struct {
	Email string `json:"email"`
}

type LoginRequest struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	RememberMe bool   `json:"remember_me"`
}

type SocialLoginRequest struct {
	Provider string `json:"provider"`
	IDToken  string `json:"id_token"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

type ResetPasswordRequest struct {
	Email       string `json:"email"`
	NewPassword string `json:"new_password"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token,omitempty"`
}

// Register handles POST /api/v1/auth/register.
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, http.StatusBadRequest, "INVALID_JSON", "Request body contains invalid JSON")
		return
	}

	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" {
		JSONError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Email is required", FieldIssue{Field: "email", Issue: "cannot be empty"})
		return
	}
	if req.Password == "" {
		JSONError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Password is required", FieldIssue{Field: "password", Issue: "cannot be empty"})
		return
	}

	cmd := auth.RegisterCommand{
		Email:           req.Email,
		FullName:        req.FullName,
		Password:        req.Password,
		ConfirmPassword: req.ConfirmPassword,
	}

	res, err := h.registerUC.Execute(r.Context(), cmd)
	if err != nil {
		if errors.Is(err, user.ErrEmailAlreadyExists) {
			JSONError(w, http.StatusConflict, "EMAIL_EXISTS", "A user with this email already exists")
			return
		}
		if errors.Is(err, user.ErrPasswordTooWeak) {
			JSONError(w, http.StatusUnprocessableEntity, "WEAK_PASSWORD", "Password must be at least 8 characters", FieldIssue{Field: "password", Issue: "must be at least 8 characters"})
			return
		}
		if err.Error() == "passwords do not match" {
			JSONError(w, http.StatusBadRequest, "PASSWORD_MISMATCH", "Passwords do not match", FieldIssue{Field: "confirm_password", Issue: "does not match password"})
			return
		}

		JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An error occurred during registration")
		return
	}

	message := "Verification OTP sent to your email"
	if !res.OTPSent {
		message = "Account created, but we couldn't send the verification email. Please request a new OTP."
	}
	JSON(w, http.StatusCreated, res, message)
}

// VerifyOTP handles POST /api/v1/auth/verify-otp.
func (h *AuthHandler) VerifyOTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	var req VerifyOTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, http.StatusBadRequest, "INVALID_JSON", "Request body contains invalid JSON")
		return
	}

	req.Email = strings.TrimSpace(req.Email)
	req.OTP = strings.TrimSpace(req.OTP)
	if req.Email == "" {
		JSONError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Email is required", FieldIssue{Field: "email", Issue: "cannot be empty"})
		return
	}
	if req.OTP == "" {
		JSONError(w, http.StatusBadRequest, "VALIDATION_ERROR", "OTP code is required", FieldIssue{Field: "otp", Issue: "cannot be empty"})
		return
	}

	cmd := auth.VerifyOTPCommand{
		Email: req.Email,
		OTP:   req.OTP,
	}

	res, err := h.verifyOTPUC.Execute(r.Context(), cmd)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			JSONError(w, http.StatusNotFound, "USER_NOT_FOUND", "No user found with the provided email")
			return
		}
		if errors.Is(err, user.ErrInvalidOTP) {
			JSONError(w, http.StatusBadRequest, "INVALID_OTP", "The verification code is incorrect")
			return
		}
		if errors.Is(err, user.ErrOTPExpired) {
			JSONError(w, http.StatusBadRequest, "OTP_EXPIRED", "The verification code has expired, please request a new one")
			return
		}

		JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An error occurred during OTP verification")
		return
	}

	JSON(w, http.StatusOK, res, "OTP Verified")
}

// ResendOTP handles POST /api/v1/auth/resend-otp.
func (h *AuthHandler) ResendOTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	var req ResendOTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, http.StatusBadRequest, "INVALID_JSON", "Request body contains invalid JSON")
		return
	}

	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" {
		JSONError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Email is required", FieldIssue{Field: "email", Issue: "cannot be empty"})
		return
	}

	cmd := auth.ResendOTPCommand{
		Email: req.Email,
	}

	if err := h.resendOTPUC.Execute(r.Context(), cmd); err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			JSONError(w, http.StatusNotFound, "USER_NOT_FOUND", "No user found with the provided email")
			return
		}

		JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to dispatch verification code")
		return
	}

	JSON(w, http.StatusOK, nil, "New verification code has been dispatched")
}

// Login handles POST /api/v1/auth/login.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, http.StatusBadRequest, "INVALID_JSON", "Request body contains invalid JSON")
		return
	}

	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" {
		JSONError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Email is required", FieldIssue{Field: "email", Issue: "cannot be empty"})
		return
	}
	if req.Password == "" {
		JSONError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Password is required", FieldIssue{Field: "password", Issue: "cannot be empty"})
		return
	}

	cmd := auth.LoginCommand{
		Email:      req.Email,
		Password:   req.Password,
		RememberMe: req.RememberMe,
	}

	res, err := h.loginUC.Execute(r.Context(), cmd)
	if err != nil {
		if errors.Is(err, user.ErrInvalidCredentials) {
			JSONError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "The email or password you entered is incorrect")
			return
		}
		if errors.Is(err, user.ErrAccountNotVerified) {
			JSONError(w, http.StatusForbidden, "ACCOUNT_UNVERIFIED", "Your account is not verified yet. Please verify your OTP")
			return
		}

		JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An error occurred during authentication")
		return
	}

	JSON(w, http.StatusOK, res, "Login successful")
}

// SocialLogin handles POST /api/v1/auth/social-login.
func (h *AuthHandler) SocialLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	var req SocialLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, http.StatusBadRequest, "INVALID_JSON", "Request body contains invalid JSON")
		return
	}

	req.Provider = strings.TrimSpace(req.Provider)
	req.IDToken = strings.TrimSpace(req.IDToken)

	if req.Provider == "" {
		JSONError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Provider is required", FieldIssue{Field: "provider", Issue: "cannot be empty"})
		return
	}
	if req.IDToken == "" {
		JSONError(w, http.StatusBadRequest, "VALIDATION_ERROR", "ID token is required", FieldIssue{Field: "id_token", Issue: "cannot be empty"})
		return
	}

	cmd := auth.SocialLoginCommand{
		Provider: req.Provider,
		IDToken:  req.IDToken,
	}

	res, err := h.socialLoginUC.Execute(r.Context(), cmd)
	if err != nil {
		if errors.Is(err, user.ErrUnsupportedSocialProvider) {
			JSONError(w, http.StatusBadRequest, "UNSUPPORTED_PROVIDER", "Supported providers are google, apple, and facebook")
			return
		}
		if errors.Is(err, user.ErrInvalidSocialToken) {
			JSONError(w, http.StatusUnauthorized, "INVALID_TOKEN", "The provided social authentication token is invalid or expired")
			return
		}

		JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An error occurred during social authentication")
		return
	}

	JSON(w, http.StatusOK, res, "Social authentication successful")
}

// Refresh handles POST /api/v1/auth/refresh.
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	var req RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, http.StatusBadRequest, "INVALID_JSON", "Request body contains invalid JSON")
		return
	}

	req.RefreshToken = strings.TrimSpace(req.RefreshToken)
	if req.RefreshToken == "" {
		JSONError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Refresh token is required", FieldIssue{Field: "refresh_token", Issue: "cannot be empty"})
		return
	}

	cmd := auth.RefreshTokenCommand{
		RefreshToken: req.RefreshToken,
	}

	res, err := h.refreshTokenUC.Execute(r.Context(), cmd)
	if err != nil {
		if errors.Is(err, user.ErrInvalidRefreshToken) {
			JSONError(w, http.StatusUnauthorized, "INVALID_REFRESH_TOKEN", "The provided refresh token is invalid or expired")
			return
		}
		if errors.Is(err, user.ErrUserNotFound) {
			JSONError(w, http.StatusUnauthorized, "USER_NOT_FOUND", "The user associated with this token no longer exists")
			return
		}
		if errors.Is(err, user.ErrAccountNotVerified) {
			JSONError(w, http.StatusForbidden, "ACCOUNT_UNVERIFIED", "Account is not verified")
			return
		}

		JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An error occurred while refreshing token")
		return
	}

	JSON(w, http.StatusOK, res, "Token refreshed successfully")
}

// ForgotPassword handles POST /api/v1/auth/forgot-password.
func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	var req ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, http.StatusBadRequest, "INVALID_JSON", "Request body contains invalid JSON")
		return
	}

	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" {
		JSONError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Email is required", FieldIssue{Field: "email", Issue: "cannot be empty"})
		return
	}

	cmd := auth.ForgotPasswordCommand{Email: req.Email}
	if err := h.forgotPasswordUC.Execute(r.Context(), cmd); err != nil {
		if errors.Is(err, user.ErrInvalidEmail) {
			JSONError(w, http.StatusBadRequest, "INVALID_EMAIL", "Invalid email format")
			return
		}
		JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An error occurred while processing password reset")
		return
	}

	JSON(w, http.StatusOK, nil, "Password reset code sent to your email")
}

// ResetPassword handles POST /api/v1/auth/reset-password.
func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	var req ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, http.StatusBadRequest, "INVALID_JSON", "Request body contains invalid JSON")
		return
	}

	req.Email = strings.TrimSpace(req.Email)

	if req.Email == "" {
		JSONError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Email is required", FieldIssue{Field: "email", Issue: "cannot be empty"})
		return
	}
	if req.NewPassword == "" {
		JSONError(w, http.StatusBadRequest, "VALIDATION_ERROR", "New password is required", FieldIssue{Field: "new_password", Issue: "cannot be empty"})
		return
	}

	cmd := auth.ResetPasswordCommand{
		Email:       req.Email,
		NewPassword: req.NewPassword,
	}

	if err := h.resetPasswordUC.Execute(r.Context(), cmd); err != nil {
		if errors.Is(err, user.ErrPasswordTooWeak) {
			JSONError(w, http.StatusBadRequest, "PASSWORD_TOO_WEAK", "Password must be at least 8 characters long")
			return
		}
		if errors.Is(err, user.ErrUserNotFound) {
			JSONError(w, http.StatusNotFound, "USER_NOT_FOUND", "User not found")
			return
		}

		JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An error occurred while resetting password")
		return
	}

	JSON(w, http.StatusOK, nil, "Password updated successfully")
}

// Logout handles POST /api/v1/auth/logout.
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	var req LogoutRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	var accessToken string
	authHeader := r.Header.Get("Authorization")
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		accessToken = strings.TrimSpace(parts[1])
	}

	cmd := auth.LogoutCommand{
		AccessToken:  accessToken,
		RefreshToken: strings.TrimSpace(req.RefreshToken),
	}

	_ = h.logoutUC.Execute(r.Context(), cmd)

	JSON(w, http.StatusOK, nil, "Logged out successfully")
}





