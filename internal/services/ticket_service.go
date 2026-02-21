package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"ava-sales/internal/models"
	"ava-sales/internal/repositories"
)

type TicketService struct {
	repo *repositories.TicketRepository
}

func NewTicketService(repo *repositories.TicketRepository) *TicketService {
	return &TicketService{repo: repo}
}

func (s *TicketService) CreateTicket(ctx context.Context, t *models.Ticket) (*models.Ticket, error) {
	t.Status = models.TicketStatusNew
	ticketNumber, err := generateTicketNumber()
	if err != nil {
		return nil, err
	}
	t.TicketNumber = ticketNumber
	return s.repo.CreateTicket(ctx, t, t.CustomerID)
}

func (s *TicketService) ValidatePatch(current *models.Ticket, patch repositories.TicketPatch) error {
	if patch.Status != nil {
		if !isValidStatus(*patch.Status) {
			return fmt.Errorf("invalid status")
		}
		next := *patch.Status
		if (next == models.TicketStatusResolved || next == models.TicketStatusRejected) && (patch.CloseReason == nil || strings.TrimSpace(*patch.CloseReason) == "") {
			return fmt.Errorf("close_reason is required when status is RESOLVED or REJECTED")
		}
		if (current.Status == models.TicketStatusResolved || current.Status == models.TicketStatusRejected) && next != current.Status {
			return fmt.Errorf("cannot transition ticket away from terminal status")
		}
	}
	return nil
}

func isValidStatus(s models.TicketStatus) bool {
	switch s {
	case models.TicketStatusNew, models.TicketStatusInReview, models.TicketStatusInProgress, models.TicketStatusNeedMoreInfo, models.TicketStatusResolved, models.TicketStatusRejected:
		return true
	default:
		return false
	}
}

func generateTicketNumber() (string, error) {
	return fmt.Sprintf("TK-%d", time.Now().UnixNano()), nil
}
