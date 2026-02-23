package repository

import (
	"context"

	"ava-sales/internal/models"
	"github.com/google/uuid"
)

type AgencyFilter struct {
	City     *string
	Province *string
	Query    *string
	Page     int
	PageSize int
}

type CustomerTicketFilter struct {
	CustomerID uuid.UUID
	Status     *models.TicketStatus
	Page       int
	PageSize   int
}

type TechTicketFilter struct {
	TechnicianID uuid.UUID
	Status       *models.TicketStatus
	AssignedOnly bool
	Page         int
	PageSize     int
}

type Repository interface {
	NextTicketNumber(ctx context.Context) (string, error)
	GetAppUserByID(ctx context.Context, userID uuid.UUID) (*models.AppUser, error)

	CreateTicket(ctx context.Context, ticket *models.Ticket) error
	GetTicketByID(ctx context.Context, ticketID uuid.UUID) (*models.Ticket, error)
	GetTicketByIDForUpdate(ctx context.Context, ticketID uuid.UUID) (*models.Ticket, error)
	ListCustomerTickets(ctx context.Context, filter CustomerTicketFilter) ([]models.Ticket, int, error)
	ListTechTickets(ctx context.Context, filter TechTicketFilter) ([]models.Ticket, int, error)
	UpdateTicket(ctx context.Context, ticket *models.Ticket) error

	CreateTicketEvent(ctx context.Context, event *models.TicketEvent) error
	ListTicketEvents(ctx context.Context, ticketID uuid.UUID) ([]models.TicketEvent, error)

	CreateAttachment(ctx context.Context, attachment *models.TicketAttachment) error
	ListAttachments(ctx context.Context, ticketID uuid.UUID) ([]models.TicketAttachment, error)

	CreateFeedback(ctx context.Context, feedback *models.TicketFeedback) error
	GetFeedback(ctx context.Context, ticketID uuid.UUID) (*models.TicketFeedback, error)
	FeedbackExists(ctx context.Context, ticketID uuid.UUID) (bool, error)

	ListAgencies(ctx context.Context, filter AgencyFilter) ([]models.Agency, int, error)
}

type TxManager interface {
	Repo() Repository
	WithTx(ctx context.Context, fn func(repo Repository) error) error
}
