package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	categoryapp "github.com/Dodul01/fundly/internal/application/category"
	domainCategory "github.com/Dodul01/fundly/internal/domain/category"
)

type CategoryHandler struct {
	listCategoriesUC   *categoryapp.ListCategoriesUseCase
	createCategoryUC   *categoryapp.CreateCategoryUseCase
	updateCategoryUC   *categoryapp.UpdateCategoryUseCase
	deleteCategoryUC   *categoryapp.DeleteCategoryUseCase
}

func NewCategoryHandler(
	listCategoriesUC *categoryapp.ListCategoriesUseCase,
	createCategoryUC *categoryapp.CreateCategoryUseCase,
	updateCategoryUC *categoryapp.UpdateCategoryUseCase,
	deleteCategoryUC *categoryapp.DeleteCategoryUseCase,
) *CategoryHandler {
	return &CategoryHandler{
		listCategoriesUC: listCategoriesUC,
		createCategoryUC: createCategoryUC,
		updateCategoryUC: updateCategoryUC,
		deleteCategoryUC: deleteCategoryUC,
	}
}

// HandleCategories routes requests for /api/v1/categories based on HTTP method.
func (h *CategoryHandler) HandleCategories(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.ListCategories(w, r)
	case http.MethodPost:
		h.CreateCategory(w, r)
	default:
		JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
	}
}

// HandleCategoryByID routes requests for /api/v1/categories/{id} based on HTTP method.
func (h *CategoryHandler) HandleCategoryByID(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPatch:
		h.UpdateCategory(w, r)
	case http.MethodDelete:
		h.DeleteCategory(w, r)
	default:
		JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
	}
}

// ListCategories handles GET /api/v1/categories.
func (h *CategoryHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	userID := UserIDFromContext(r.Context())
	if userID == "" {
		JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing authenticated user context")
		return
	}

	typeFilter := r.URL.Query().Get("type")
	categories, err := h.listCategoriesUC.Execute(r.Context(), typeFilter)
	if err != nil {
		if errors.Is(err, domainCategory.ErrInvalidCategoryType) {
			JSONError(w, http.StatusBadRequest, "INVALID_TYPE", "Invalid category type filter: must be income or expense")
			return
		}
		JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve categories")
		return
	}

	JSON(w, http.StatusOK, categories, "")
}

type CreateCategoryRequest struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Icon  string `json:"icon"`
	Color string `json:"color"`
}

// CreateCategory handles POST /api/v1/categories.
func (h *CategoryHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	userID := UserIDFromContext(r.Context())
	if userID == "" {
		JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing authenticated user context")
		return
	}

	var req CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	res, err := h.createCategoryUC.Execute(r.Context(), categoryapp.CreateCategoryInput{
		Name:  req.Name,
		Type:  req.Type,
		Icon:  req.Icon,
		Color: req.Color,
	})
	if err != nil {
		switch {
		case errors.Is(err, domainCategory.ErrInvalidCategoryName):
			JSONError(w, http.StatusBadRequest, "INVALID_NAME", "Category name cannot be empty")
		case errors.Is(err, domainCategory.ErrInvalidCategoryType):
			JSONError(w, http.StatusBadRequest, "INVALID_TYPE", "Invalid category type: must be income or expense")
		default:
			JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create category")
		}
		return
	}

	JSON(w, http.StatusCreated, res, "Category created successfully")
}

type UpdateCategoryRequest struct {
	Name  *string `json:"name"`
	Color *string `json:"color"`
	Icon  *string `json:"icon"`
}

// UpdateCategory handles PATCH /api/v1/categories/{id}.
func (h *CategoryHandler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	userID := UserIDFromContext(r.Context())
	if userID == "" {
		JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing authenticated user context")
		return
	}

	categoryID := extractCategoryID(r)
	if categoryID == "" {
		JSONError(w, http.StatusBadRequest, "INVALID_CATEGORY_ID", "Category ID is required")
		return
	}

	var req UpdateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	res, err := h.updateCategoryUC.Execute(r.Context(), categoryapp.UpdateCategoryInput{
		ID:    categoryID,
		Name:  req.Name,
		Color: req.Color,
		Icon:  req.Icon,
	})
	if err != nil {
		switch {
		case errors.Is(err, domainCategory.ErrCategoryNotFound):
			JSONError(w, http.StatusNotFound, "CATEGORY_NOT_FOUND", "Category not found")
		case errors.Is(err, domainCategory.ErrInvalidCategoryName):
			JSONError(w, http.StatusBadRequest, "INVALID_NAME", "Category name cannot be empty")
		case errors.Is(err, domainCategory.ErrNoFieldsToUpdate):
			JSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "At least one field must be provided to update")
		default:
			JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update category")
		}
		return
	}

	JSON(w, http.StatusOK, res, "Category updated successfully")
}

// DeleteCategory handles DELETE /api/v1/categories/{id}.
func (h *CategoryHandler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	userID := UserIDFromContext(r.Context())
	if userID == "" {
		JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing authenticated user context")
		return
	}

	categoryID := extractCategoryID(r)
	if categoryID == "" {
		JSONError(w, http.StatusBadRequest, "INVALID_CATEGORY_ID", "Category ID is required")
		return
	}

	err := h.deleteCategoryUC.Execute(r.Context(), categoryID)
	if err != nil {
		if errors.Is(err, domainCategory.ErrCategoryNotFound) {
			JSONError(w, http.StatusNotFound, "CATEGORY_NOT_FOUND", "Category not found")
			return
		}
		JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete category")
		return
	}

	JSON(w, http.StatusOK, nil, "Category deleted successfully")
}

func extractCategoryID(r *http.Request) string {
	id := r.PathValue("id")
	if id != "" {
		return id
	}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) >= 4 && parts[0] == "api" && parts[1] == "v1" && parts[2] == "categories" {
		return parts[3]
	}
	return ""
}
