package service

import (
	"context"
	"errors"
	"testing"
	"time"

	appauth "ava-sales/internal/auth"
	"ava-sales/internal/models"
	"ava-sales/internal/repository"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type mockTxManager struct {
	repo *mockRepo
}

func (m *mockTxManager) Repo() repository.Repository { return m.repo }
func (m *mockTxManager) WithTx(ctx context.Context, fn func(repo repository.Repository) error) error {
	return fn(m.repo)
}

type mockRepo struct {
	usersByID       map[uuid.UUID]*models.AppUser
	usersByIdentity map[string]*models.AppUser
	refreshByHash   map[string]*models.RefreshToken
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		usersByID:       map[uuid.UUID]*models.AppUser{},
		usersByIdentity: map[string]*models.AppUser{},
		refreshByHash:   map[string]*models.RefreshToken{},
	}
}

func (m *mockRepo) NextTicketNumber(ctx context.Context) (string, error) { return "", errors.New("not implemented") }
func (m *mockRepo) CreateTicket(ctx context.Context, ticket *models.Ticket) error { return errors.New("not implemented") }
func (m *mockRepo) GetTicketByID(ctx context.Context, ticketID uuid.UUID) (*models.Ticket, error) {
	return nil, errors.New("not implemented")
}
func (m *mockRepo) GetTicketByIDForUpdate(ctx context.Context, ticketID uuid.UUID) (*models.Ticket, error) {
	return nil, errors.New("not implemented")
}
func (m *mockRepo) ListCustomerTickets(ctx context.Context, filter repository.CustomerTicketFilter) ([]models.Ticket, int, error) {
	return nil, 0, errors.New("not implemented")
}
func (m *mockRepo) ListTechTickets(ctx context.Context, filter repository.TechTicketFilter) ([]models.Ticket, int, error) {
	return nil, 0, errors.New("not implemented")
}
func (m *mockRepo) UpdateTicket(ctx context.Context, ticket *models.Ticket) error { return errors.New("not implemented") }
func (m *mockRepo) CreateTicketEvent(ctx context.Context, event *models.TicketEvent) error { return errors.New("not implemented") }
func (m *mockRepo) ListTicketEvents(ctx context.Context, ticketID uuid.UUID) ([]models.TicketEvent, error) {
	return nil, errors.New("not implemented")
}
func (m *mockRepo) CreateAttachment(ctx context.Context, attachment *models.TicketAttachment) error {
	return errors.New("not implemented")
}
func (m *mockRepo) ListAttachments(ctx context.Context, ticketID uuid.UUID) ([]models.TicketAttachment, error) {
	return nil, errors.New("not implemented")
}
func (m *mockRepo) CreateFeedback(ctx context.Context, feedback *models.TicketFeedback) error {
	return errors.New("not implemented")
}
func (m *mockRepo) GetFeedback(ctx context.Context, ticketID uuid.UUID) (*models.TicketFeedback, error) {
	return nil, errors.New("not implemented")
}
func (m *mockRepo) FeedbackExists(ctx context.Context, ticketID uuid.UUID) (bool, error) {
	return false, errors.New("not implemented")
}
func (m *mockRepo) ListAgencies(ctx context.Context, filter repository.AgencyFilter) ([]models.Agency, int, error) {
	return nil, 0, errors.New("not implemented")
}

func (m *mockRepo) GetAppUserByID(ctx context.Context, userID uuid.UUID) (*models.AppUser, error) {
	u, ok := m.usersByID[userID]
	if !ok {
		return nil, models.ErrNotFound
	}
	return u, nil
}

func (m *mockRepo) GetAppUserByEmailOrUsername(ctx context.Context, identity string) (*models.AppUser, error) {
	u, ok := m.usersByIdentity[identity]
	if !ok {
		return nil, models.ErrNotFound
	}
	return u, nil
}

func (m *mockRepo) CreateRefreshToken(ctx context.Context, token *models.RefreshToken) error {
	m.refreshByHash[token.TokenHash] = token
	return nil
}

func (m *mockRepo) GetRefreshTokenByHash(ctx context.Context, tokenHash string) (*models.RefreshToken, error) {
	rt, ok := m.refreshByHash[tokenHash]
	if !ok {
		return nil, models.ErrNotFound
	}
	return rt, nil
}

func (m *mockRepo) RevokeRefreshToken(ctx context.Context, tokenID uuid.UUID) error {
	for _, rt := range m.refreshByHash {
		if rt.ID == tokenID {
			now := time.Now().UTC()
			rt.RevokedAt = &now
			return nil
		}
	}
	return models.ErrNotFound
}

func TestAuthService_LoginSuccess(t *testing.T) {
	repo := newMockRepo()
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)
	userID := uuid.New()
	user := &models.AppUser{ID: userID, Email: "user@example.com", PasswordHash: string(passwordHash), Role: models.RoleCustomer, IsActive: true}
	repo.usersByID[userID] = user
	repo.usersByIdentity["user@example.com"] = user

	svc := NewAuthService(&mockTxManager{repo: repo}, "secret-key", 15, 7)
	resp, err := svc.Login(context.Background(), "user@example.com", "secret")
	if err != nil {
		t.Fatalf("login error: %v", err)
	}
	if resp.AccessToken == "" || resp.RefreshToken == "" {
		t.Fatalf("expected tokens in response")
	}
	if resp.ExpiresIn != 900 {
		t.Fatalf("expected expires_in 900, got %d", resp.ExpiresIn)
	}
}

func TestAuthService_LoginWrongPassword(t *testing.T) {
	repo := newMockRepo()
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)
	userID := uuid.New()
	user := &models.AppUser{ID: userID, Email: "user@example.com", PasswordHash: string(passwordHash), Role: models.RoleCustomer, IsActive: true}
	repo.usersByID[userID] = user
	repo.usersByIdentity["user@example.com"] = user

	svc := NewAuthService(&mockTxManager{repo: repo}, "secret-key", 15, 7)
	_, err := svc.Login(context.Background(), "user@example.com", "wrong")
	if err == nil {
		t.Fatal("expected wrong password login to fail")
	}
}

func TestAuthService_RefreshFlow(t *testing.T) {
	repo := newMockRepo()
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)
	userID := uuid.New()
	user := &models.AppUser{ID: userID, Email: "user@example.com", PasswordHash: string(passwordHash), Role: models.RoleCustomer, IsActive: true}
	repo.usersByID[userID] = user
	repo.usersByIdentity["user@example.com"] = user

	svc := NewAuthService(&mockTxManager{repo: repo}, "secret-key", 15, 7)
	loginResp, err := svc.Login(context.Background(), "user@example.com", "secret")
	if err != nil {
		t.Fatalf("login error: %v", err)
	}

	refreshResp, err := svc.Refresh(context.Background(), loginResp.RefreshToken)
	if err != nil {
		t.Fatalf("refresh error: %v", err)
	}
	if refreshResp.AccessToken == "" || refreshResp.RefreshToken == "" {
		t.Fatalf("expected rotated tokens")
	}

	oldHash := appauth.HashToken(loginResp.RefreshToken)
	oldToken, ok := repo.refreshByHash[oldHash]
	if !ok || oldToken.RevokedAt == nil {
		t.Fatalf("expected old refresh token to be revoked")
	}
}

func TestAuthService_RefreshExpiredRejected(t *testing.T) {
	repo := newMockRepo()
	userID := uuid.New()
	user := &models.AppUser{ID: userID, Email: "user@example.com", PasswordHash: "x", Role: models.RoleCustomer, IsActive: true}
	repo.usersByID[userID] = user

	raw := "expired-token"
	hash := appauth.HashToken(raw)
	repo.refreshByHash[hash] = &models.RefreshToken{ID: uuid.New(), UserID: userID, TokenHash: hash, ExpiresAt: time.Now().UTC().Add(-time.Minute)}

	svc := NewAuthService(&mockTxManager{repo: repo}, "secret-key", 15, 7)
	_, err := svc.Refresh(context.Background(), raw)
	if err == nil {
		t.Fatal("expected expired refresh token to fail")
	}
}
