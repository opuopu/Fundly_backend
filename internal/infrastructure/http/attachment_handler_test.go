package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	attachmentapp "github.com/Dodul01/fundly/internal/application/attachment"
	domainAttachment "github.com/Dodul01/fundly/internal/domain/attachment"
	infrahttp "github.com/Dodul01/fundly/internal/infrastructure/http"
)

type attachmentRepoStub struct {
	items map[string]*domainAttachment.Attachment
}

func (m *attachmentRepoStub) Create(_ context.Context, a *domainAttachment.Attachment) error {
	m.items[a.ID] = a
	return nil
}
func (m *attachmentRepoStub) GetByID(_ context.Context, id string) (*domainAttachment.Attachment, error) {
	a, ok := m.items[id]
	if !ok {
		return nil, domainAttachment.ErrAttachmentNotFound
	}
	return a, nil
}
func (m *attachmentRepoStub) Delete(_ context.Context, id string) error {
	delete(m.items, id)
	return nil
}
func (m *attachmentRepoStub) ListByTransactionID(_ context.Context, _ string) ([]*domainAttachment.Attachment, error) {
	return nil, nil
}
func (m *attachmentRepoStub) CountByTransactionID(_ context.Context, _ string) (int, error) {
	return 0, nil
}
func (m *attachmentRepoStub) LinkToTransaction(_ context.Context, _, _ string, _ []string) error {
	return nil
}

type storageStub struct{}

func (storageStub) Save(_ context.Context, subDir, filename string, reader io.Reader) (string, error) {
	_, _ = io.Copy(io.Discard, reader)
	return "/uploads/" + subDir + "/" + filename, nil
}

func setupAttachmentHandler() (*infrahttp.AttachmentHandler, *attachmentRepoStub) {
	repo := &attachmentRepoStub{items: make(map[string]*domainAttachment.Attachment)}
	return infrahttp.NewAttachmentHandler(
		attachmentapp.NewUploadAttachmentUseCase(repo, storageStub{}),
		attachmentapp.NewGetAttachmentUseCase(repo),
		attachmentapp.NewDeleteAttachmentUseCase(repo),
	), repo
}

func multipartPNG(t *testing.T, field, filename string) (*bytes.Buffer, string) {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	part, err := w.CreateFormFile(field, filename)
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	png := append([]byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}, make([]byte, 64)...)
	_, _ = part.Write(png)
	_ = w.Close()
	return &buf, w.FormDataContentType()
}

func TestAttachmentHandler_Upload(t *testing.T) {
	handler, repo := setupAttachmentHandler()
	body, contentType := multipartPNG(t, "file", "receipt.png")

	req := authed(httptest.NewRequest(http.MethodPost, "/api/v1/attachments", body), "usr_1")
	req.Header.Set("Content-Type", contentType)
	rec := httptest.NewRecorder()

	handler.HandleAttachments(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d. body: %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Data attachmentapp.AttachmentResponse `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if resp.Data.ContentType != "image/png" {
		t.Errorf("expected image/png, got %s", resp.Data.ContentType)
	}
	if len(repo.items) != 1 {
		t.Errorf("expected 1 persisted attachment, got %d", len(repo.items))
	}
}

func TestAttachmentHandler_Upload_MissingFile(t *testing.T) {
	handler, _ := setupAttachmentHandler()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	_ = w.WriteField("other", "x")
	_ = w.Close()

	req := authed(httptest.NewRequest(http.MethodPost, "/api/v1/attachments", &buf), "usr_1")
	req.Header.Set("Content-Type", w.FormDataContentType())
	rec := httptest.NewRecorder()

	handler.HandleAttachments(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestAttachmentHandler_Get_NotFound(t *testing.T) {
	handler, _ := setupAttachmentHandler()
	req := authed(httptest.NewRequest(http.MethodGet, "/api/v1/attachments/att_missing", nil), "usr_1")
	req.SetPathValue("id", "att_missing")
	rec := httptest.NewRecorder()

	handler.HandleAttachmentByID(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestAttachmentHandler_DeleteOwned(t *testing.T) {
	handler, repo := setupAttachmentHandler()
	repo.items["att_1"] = domainAttachment.New("att_1", "usr_1", "r.png", "/u/r.png", "image/png", 100)

	req := authed(httptest.NewRequest(http.MethodDelete, "/api/v1/attachments/att_1", nil), "usr_1")
	req.SetPathValue("id", "att_1")
	rec := httptest.NewRecorder()

	handler.HandleAttachmentByID(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. body: %s", rec.Code, rec.Body.String())
	}
	if len(repo.items) != 0 {
		t.Errorf("expected attachment deleted")
	}
}

func TestAttachmentHandler_Unauthorized(t *testing.T) {
	handler, _ := setupAttachmentHandler()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/attachments", nil)
	rec := httptest.NewRecorder()

	handler.HandleAttachments(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}
