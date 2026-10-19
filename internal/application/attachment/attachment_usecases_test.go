package attachment_test

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	attachmentapp "github.com/Dodul01/fundly/internal/application/attachment"
	domainAttachment "github.com/Dodul01/fundly/internal/domain/attachment"
)

type mockAttachmentRepo struct {
	items map[string]*domainAttachment.Attachment
}

func newMockRepo() *mockAttachmentRepo {
	return &mockAttachmentRepo{items: make(map[string]*domainAttachment.Attachment)}
}

func (m *mockAttachmentRepo) Create(_ context.Context, a *domainAttachment.Attachment) error {
	m.items[a.ID] = a
	return nil
}
func (m *mockAttachmentRepo) GetByID(_ context.Context, id string) (*domainAttachment.Attachment, error) {
	a, ok := m.items[id]
	if !ok {
		return nil, domainAttachment.ErrAttachmentNotFound
	}
	return a, nil
}
func (m *mockAttachmentRepo) Delete(_ context.Context, id string) error {
	if _, ok := m.items[id]; !ok {
		return domainAttachment.ErrAttachmentNotFound
	}
	delete(m.items, id)
	return nil
}
func (m *mockAttachmentRepo) ListByTransactionID(_ context.Context, _ string) ([]*domainAttachment.Attachment, error) {
	return nil, nil
}
func (m *mockAttachmentRepo) CountByTransactionID(_ context.Context, _ string) (int, error) {
	return 0, nil
}
func (m *mockAttachmentRepo) LinkToTransaction(_ context.Context, _, _ string, _ []string) error {
	return nil
}

// mockStorage records the saved file and returns a deterministic URL.
type mockStorage struct {
	savedBytes int
}

func (m *mockStorage) Save(_ context.Context, subDir, filename string, reader io.Reader) (string, error) {
	b, _ := io.ReadAll(reader)
	m.savedBytes = len(b)
	return "/uploads/" + subDir + "/" + filename, nil
}

// a minimal PNG header so http.DetectContentType yields image/png.
var pngData = append([]byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}, make([]byte, 64)...)

func TestUploadAttachment_Success(t *testing.T) {
	repo := newMockRepo()
	store := &mockStorage{}
	uc := attachmentapp.NewUploadAttachmentUseCase(repo, store)

	res, err := uc.Execute(context.Background(), attachmentapp.UploadAttachmentInput{
		UserID:      "usr_1",
		Filename:    "receipt.png",
		ContentType: "image/png",
		Size:        int64(len(pngData)),
		Reader:      strings.NewReader(string(pngData)),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.ID == "" || !strings.HasPrefix(res.ID, "att_") {
		t.Errorf("expected generated att_ id, got %q", res.ID)
	}
	if res.ContentType != "image/png" {
		t.Errorf("expected image/png, got %s", res.ContentType)
	}
	if res.FileName != "receipt.png" {
		t.Errorf("expected original filename preserved, got %s", res.FileName)
	}
	if len(repo.items) != 1 {
		t.Errorf("expected attachment persisted, got %d", len(repo.items))
	}
	if store.savedBytes == 0 {
		t.Error("expected storage to receive file bytes")
	}
}

func TestUploadAttachment_TooLarge(t *testing.T) {
	uc := attachmentapp.NewUploadAttachmentUseCase(newMockRepo(), &mockStorage{})
	_, err := uc.Execute(context.Background(), attachmentapp.UploadAttachmentInput{
		UserID:      "usr_1",
		Filename:    "big.png",
		ContentType: "image/png",
		Size:        attachmentapp.MaxAttachmentSize + 1,
		Reader:      strings.NewReader("x"),
	})
	if !errors.Is(err, domainAttachment.ErrFileTooLarge) {
		t.Fatalf("expected ErrFileTooLarge, got %v", err)
	}
}

func TestUploadAttachment_UnsupportedType(t *testing.T) {
	uc := attachmentapp.NewUploadAttachmentUseCase(newMockRepo(), &mockStorage{})
	data := "this is plain text, not an image or pdf"
	_, err := uc.Execute(context.Background(), attachmentapp.UploadAttachmentInput{
		UserID:      "usr_1",
		Filename:    "note.txt",
		ContentType: "text/plain",
		Size:        int64(len(data)),
		Reader:      strings.NewReader(data),
	})
	if !errors.Is(err, domainAttachment.ErrUnsupportedType) {
		t.Fatalf("expected ErrUnsupportedType, got %v", err)
	}
}

func TestUploadAttachment_SpoofedContentTypeRejected(t *testing.T) {
	uc := attachmentapp.NewUploadAttachmentUseCase(newMockRepo(), &mockStorage{})
	// Arbitrary non-image bytes, but the client lies and declares image/png.
	data := "GET / HTTP/1.1\r\nHost: evil\r\n\r\n<script>alert(1)</script>"
	_, err := uc.Execute(context.Background(), attachmentapp.UploadAttachmentInput{
		UserID:      "usr_1",
		Filename:    "payload.png",
		ContentType: "image/png",
		Size:        int64(len(data)),
		Reader:      strings.NewReader(data),
	})
	if !errors.Is(err, domainAttachment.ErrUnsupportedType) {
		t.Fatalf("expected spoofed content type to be rejected, got %v", err)
	}
}

func TestGetAttachment_OwnershipGuard(t *testing.T) {
	repo := newMockRepo()
	repo.items["att_1"] = domainAttachment.New("att_1", "owner", "r.png", "/u/r.png", "image/png", 100)

	uc := attachmentapp.NewGetAttachmentUseCase(repo)
	if _, err := uc.Execute(context.Background(), "owner", "att_1"); err != nil {
		t.Fatalf("owner should fetch: %v", err)
	}
	_, err := uc.Execute(context.Background(), "intruder", "att_1")
	if !errors.Is(err, domainAttachment.ErrAttachmentNotFound) {
		t.Fatalf("expected not-found for non-owner, got %v", err)
	}
}

func TestDeleteAttachment_OwnershipGuard(t *testing.T) {
	repo := newMockRepo()
	repo.items["att_1"] = domainAttachment.New("att_1", "owner", "r.png", "/u/r.png", "image/png", 100)

	uc := attachmentapp.NewDeleteAttachmentUseCase(repo)
	if err := uc.Execute(context.Background(), "intruder", "att_1"); !errors.Is(err, domainAttachment.ErrAttachmentNotFound) {
		t.Fatalf("expected not-found for non-owner, got %v", err)
	}
	if err := uc.Execute(context.Background(), "owner", "att_1"); err != nil {
		t.Fatalf("owner delete failed: %v", err)
	}
	if len(repo.items) != 0 {
		t.Errorf("expected attachment removed")
	}
}
