package http

import (
	"errors"
	"net/http"
	"strings"

	attachmentapp "github.com/Dodul01/fundly/internal/application/attachment"
	domainAttachment "github.com/Dodul01/fundly/internal/domain/attachment"
)

type AttachmentHandler struct {
	uploadUC *attachmentapp.UploadAttachmentUseCase
	getUC    *attachmentapp.GetAttachmentUseCase
	deleteUC *attachmentapp.DeleteAttachmentUseCase
}

func NewAttachmentHandler(
	uploadUC *attachmentapp.UploadAttachmentUseCase,
	getUC *attachmentapp.GetAttachmentUseCase,
	deleteUC *attachmentapp.DeleteAttachmentUseCase,
) *AttachmentHandler {
	return &AttachmentHandler{uploadUC: uploadUC, getUC: getUC, deleteUC: deleteUC}
}

// HandleAttachments routes /api/v1/attachments.
func (h *AttachmentHandler) HandleAttachments(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		h.Upload(w, r)
		return
	}
	JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
}

// HandleAttachmentByID routes /api/v1/attachments/{id}.
func (h *AttachmentHandler) HandleAttachmentByID(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.Get(w, r)
	case http.MethodDelete:
		h.Delete(w, r)
	default:
		JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
	}
}

// Upload handles POST /api/v1/attachments.
func (h *AttachmentHandler) Upload(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	if userID == "" {
		JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing authenticated user context")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, attachmentapp.MaxAttachmentSize+1024)
	if err := r.ParseMultipartForm(attachmentapp.MaxAttachmentSize); err != nil {
		JSONError(w, http.StatusBadRequest, "FILE_TOO_LARGE", "Uploaded file exceeds maximum allowed limit of 10MB")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		JSONError(w, http.StatusBadRequest, "MISSING_FILE", "A file is required under the 'file' form field")
		return
	}
	defer file.Close()

	res, err := h.uploadUC.Execute(r.Context(), attachmentapp.UploadAttachmentInput{
		UserID:      userID,
		Filename:    header.Filename,
		ContentType: header.Header.Get("Content-Type"),
		Size:        header.Size,
		Reader:      file,
	})
	if err != nil {
		writeAttachmentError(w, err, "Failed to upload attachment")
		return
	}

	JSON(w, http.StatusCreated, res, "Attachment uploaded successfully")
}

// Get handles GET /api/v1/attachments/{id}.
func (h *AttachmentHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	if userID == "" {
		JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing authenticated user context")
		return
	}

	id := extractAttachmentID(r)
	if id == "" {
		JSONError(w, http.StatusBadRequest, "INVALID_ATTACHMENT_ID", "Attachment ID is required")
		return
	}

	res, err := h.getUC.Execute(r.Context(), userID, id)
	if err != nil {
		writeAttachmentError(w, err, "Failed to retrieve attachment")
		return
	}

	JSON(w, http.StatusOK, res, "")
}

// Delete handles DELETE /api/v1/attachments/{id}.
func (h *AttachmentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	if userID == "" {
		JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing authenticated user context")
		return
	}

	id := extractAttachmentID(r)
	if id == "" {
		JSONError(w, http.StatusBadRequest, "INVALID_ATTACHMENT_ID", "Attachment ID is required")
		return
	}

	if err := h.deleteUC.Execute(r.Context(), userID, id); err != nil {
		writeAttachmentError(w, err, "Failed to delete attachment")
		return
	}

	JSON(w, http.StatusOK, nil, "Attachment deleted successfully")
}

func writeAttachmentError(w http.ResponseWriter, err error, fallback string) {
	switch {
	case errors.Is(err, domainAttachment.ErrAttachmentNotFound):
		JSONError(w, http.StatusNotFound, "ATTACHMENT_NOT_FOUND", "Attachment not found")
	case errors.Is(err, domainAttachment.ErrFileTooLarge):
		JSONError(w, http.StatusBadRequest, "FILE_TOO_LARGE", "File exceeds the maximum allowed size of 10MB")
	case errors.Is(err, domainAttachment.ErrUnsupportedType):
		JSONError(w, http.StatusBadRequest, "UNSUPPORTED_TYPE", "Only images and PDF files are allowed")
	case errors.Is(err, domainAttachment.ErrInvalidFile):
		JSONError(w, http.StatusBadRequest, "INVALID_FILE", "Invalid file payload")
	default:
		JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", fallback)
	}
}

func extractAttachmentID(r *http.Request) string {
	id := r.PathValue("id")
	if id != "" {
		return id
	}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) >= 4 && parts[0] == "api" && parts[1] == "v1" && parts[2] == "attachments" {
		return parts[3]
	}
	return ""
}
