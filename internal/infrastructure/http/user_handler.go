package http

import (
	"encoding/json"
	"errors"
	"net/http"

	userapp "github.com/Dodul01/fundly/internal/application/user"
	domainUser "github.com/Dodul01/fundly/internal/domain/user"
)

type UserHandler struct {
	getProfileUC     *userapp.GetProfileUseCase
	updateProfileUC  *userapp.UpdateProfileUseCase
	changePasswordUC *userapp.ChangePasswordUseCase
	uploadAvatarUC   *userapp.UploadAvatarUseCase
}

func NewUserHandler(
	getProfileUC *userapp.GetProfileUseCase,
	updateProfileUC *userapp.UpdateProfileUseCase,
	changePasswordUC *userapp.ChangePasswordUseCase,
	uploadAvatarUC *userapp.UploadAvatarUseCase,
) *UserHandler {
	return &UserHandler{
		getProfileUC:     getProfileUC,
		updateProfileUC:  updateProfileUC,
		changePasswordUC: changePasswordUC,
		uploadAvatarUC:   uploadAvatarUC,
	}
}

// HandleMe routes requests for /api/v1/users/me based on HTTP method.
func (h *UserHandler) HandleMe(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.GetProfile(w, r)
	case http.MethodPatch:
		h.UpdateProfile(w, r)
	default:
		JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
	}
}

// GetProfile handles GET /api/v1/users/me.
func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	userID := UserIDFromContext(r.Context())
	if userID == "" {
		JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing authenticated user context")
		return
	}

	profile, err := h.getProfileUC.Execute(r.Context(), userID)
	if err != nil {
		if errors.Is(err, domainUser.ErrUserNotFound) {
			JSONError(w, http.StatusNotFound, "NOT_FOUND", "User profile not found")
			return
		}
		JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve user profile")
		return
	}

	JSON(w, http.StatusOK, profile, "")
}

// UpdateProfile handles PATCH /api/v1/users/me.
func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	userID := UserIDFromContext(r.Context())
	if userID == "" {
		JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing authenticated user context")
		return
	}

	var input userapp.UpdateProfileInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		JSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid JSON payload")
		return
	}

	profile, err := h.updateProfileUC.Execute(r.Context(), userID, input)
	if err != nil {
		if errors.Is(err, domainUser.ErrInvalidUserData) {
			JSONError(w, http.StatusBadRequest, "INVALID_INPUT", "At least one valid field (full_name or email) must be provided")
			return
		}
		if errors.Is(err, domainUser.ErrInvalidEmail) {
			JSONError(w, http.StatusBadRequest, "INVALID_EMAIL", "The provided email format is invalid")
			return
		}
		if errors.Is(err, domainUser.ErrEmailAlreadyExists) {
			JSONError(w, http.StatusConflict, "EMAIL_EXISTS", "This email address is already registered to another account")
			return
		}
		if errors.Is(err, domainUser.ErrUserNotFound) {
			JSONError(w, http.StatusNotFound, "NOT_FOUND", "User profile not found")
			return
		}
		JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update user profile")
		return
	}

	JSON(w, http.StatusOK, profile, "Profile updated successfully")
}

// ChangePassword handles PUT /api/v1/users/me/password.
func (h *UserHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	userID := UserIDFromContext(r.Context())
	if userID == "" {
		JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing authenticated user context")
		return
	}

	var input userapp.ChangePasswordInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		JSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid JSON payload")
		return
	}

	if err := h.changePasswordUC.Execute(r.Context(), userID, input); err != nil {
		if errors.Is(err, domainUser.ErrInvalidUserData) {
			JSONError(w, http.StatusBadRequest, "INVALID_INPUT", "Current password and new password are required")
			return
		}
		if errors.Is(err, domainUser.ErrPasswordMismatch) {
			JSONError(w, http.StatusBadRequest, "PASSWORD_MISMATCH", "New password and confirmation password do not match")
			return
		}
		if errors.Is(err, domainUser.ErrPasswordTooWeak) {
			JSONError(w, http.StatusBadRequest, "WEAK_PASSWORD", "Password must be at least 8 characters long")
			return
		}
		if errors.Is(err, domainUser.ErrInvalidCurrentPassword) {
			JSONError(w, http.StatusBadRequest, "INVALID_CREDENTIALS", "Current password is incorrect")
			return
		}
		if errors.Is(err, domainUser.ErrUserNotFound) {
			JSONError(w, http.StatusNotFound, "NOT_FOUND", "User profile not found")
			return
		}
		JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to change password")
		return
	}

	JSON(w, http.StatusOK, nil, "Password changed successfully")
}

// UploadAvatar handles POST /api/v1/users/me/avatar.
func (h *UserHandler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	userID := UserIDFromContext(r.Context())
	if userID == "" {
		JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing authenticated user context")
		return
	}

	// Limit request body to 5MB + 1KB overhead
	r.Body = http.MaxBytesReader(w, r.Body, userapp.MaxAvatarSize+1024)
	if err := r.ParseMultipartForm(userapp.MaxAvatarSize); err != nil {
		JSONError(w, http.StatusBadRequest, "FILE_TOO_LARGE", "Uploaded file exceeds maximum allowed limit of 5MB")
		return
	}

	file, header, err := r.FormFile("avatar")
	if err != nil {
		JSONError(w, http.StatusBadRequest, "MISSING_FILE", "Image file is required under the 'avatar' form field")
		return
	}
	defer file.Close()

	res, err := h.uploadAvatarUC.Execute(r.Context(), userID, userapp.UploadAvatarInput{
		Filename:    header.Filename,
		ContentType: header.Header.Get("Content-Type"),
		Size:        header.Size,
		Reader:      file,
	})
	if err != nil {
		if errors.Is(err, domainUser.ErrImageTooLarge) {
			JSONError(w, http.StatusBadRequest, "FILE_TOO_LARGE", "Image size exceeds maximum allowed limit of 5MB")
			return
		}
		if errors.Is(err, domainUser.ErrInvalidImageFormat) {
			JSONError(w, http.StatusBadRequest, "INVALID_FORMAT", "Only jpeg, png, and webp images are allowed")
			return
		}
		if errors.Is(err, domainUser.ErrInvalidUserData) {
			JSONError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid image payload")
			return
		}
		if errors.Is(err, domainUser.ErrUserNotFound) {
			JSONError(w, http.StatusNotFound, "NOT_FOUND", "User profile not found")
			return
		}
		JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to upload avatar")
		return
	}

	JSON(w, http.StatusOK, res, "Avatar uploaded successfully")
}
