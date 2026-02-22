package service

import (
	"context"
	"net/http"
	"strings"
	"time"

	"ava-sales/internal/models"
	"ava-sales/internal/repository"
	"github.com/google/uuid"
)

type TicketService struct {
	tx repository.TxManager
}

func NewTicketService(tx repository.TxManager) *TicketService {
	return &TicketService{tx: tx}
}

type CreateTicketInput struct {
	CustomerID   uuid.UUID
	DeviceSerial string
	Subject      string
	Description  string
	Category     *string
}

type AddAttachmentInput struct {
	TicketID   uuid.UUID
	CustomerID uuid.UUID
	FileType   string
	FileURL    string
}

type CreateFeedbackInput struct {
	TicketID   uuid.UUID
	CustomerID uuid.UUID
	Score      int
	Solved     bool
	Comment    *string
}

type PatchTicketInput struct {
	Status               *models.TicketStatus
	AssignedTechnicianID *uuid.UUID
	CloseReason          *string
	Actor                models.Actor
}

type AddCommentInput struct {
	TicketID uuid.UUID
	TechID   uuid.UUID
	Message  string
}

func (s *TicketService) ListCustomerTickets(ctx context.Context, customerID uuid.UUID, status *models.TicketStatus, page, pageSize int) ([]models.Ticket, models.Pagination, error) {
	filter := repository.CustomerTicketFilter{
		CustomerID: customerID,
		Status:     status,
		Page:       page,
		PageSize:   pageSize,
	}
	items, total, err := s.tx.Repo().ListCustomerTickets(ctx, filter)
	if err != nil {
		return nil, models.Pagination{}, err
	}
	return items, models.Pagination{Page: page, PageSize: pageSize, Total: total}, nil
}

func (s *TicketService) GetCustomerTicketDetail(ctx context.Context, customerID, ticketID uuid.UUID) (*models.TicketDetail, error) {
	ticket, err := s.tx.Repo().GetTicketByID(ctx, ticketID)
	if err != nil {
		return nil, err
	}
	if ticket.CustomerID != customerID {
		return nil, models.ErrForbidden
	}

	attachments, err := s.tx.Repo().ListAttachments(ctx, ticketID)
	if err != nil {
		return nil, err
	}
	events, err := s.tx.Repo().ListTicketEvents(ctx, ticketID)
	if err != nil {
		return nil, err
	}
	feedback, err := s.tx.Repo().GetFeedback(ctx, ticketID)
	if err != nil {
		return nil, err
	}

	return &models.TicketDetail{Ticket: *ticket, Attachments: attachments, Events: events, Feedback: feedback}, nil
}

func (s *TicketService) CreateCustomerTicket(ctx context.Context, in CreateTicketInput) (*models.Ticket, error) {
	if strings.TrimSpace(in.DeviceSerial) == "" || strings.TrimSpace(in.Subject) == "" || strings.TrimSpace(in.Description) == "" {
		return nil, models.NewAPIError("VALIDATION_ERROR", "device_serial, subject, and description are required", http.StatusBadRequest, nil)
	}

	var out *models.Ticket
	err := s.tx.WithTx(ctx, func(repo repository.Repository) error {
		ticketNumber, err := repo.NextTicketNumber(ctx)
		if err != nil {
			return err
		}

		ticket := &models.Ticket{
			ID:           uuid.New(),
			TicketNumber: ticketNumber,
			CustomerID:   in.CustomerID,
			DeviceSerial: strings.TrimSpace(in.DeviceSerial),
			Subject:      strings.TrimSpace(in.Subject),
			Description:  strings.TrimSpace(in.Description),
			Category:     trimStringPtr(in.Category),
			Status:       models.TicketStatusNew,
		}

		if err := repo.CreateTicket(ctx, ticket); err != nil {
			return err
		}

		to := models.TicketStatusNew
		event := &models.TicketEvent{
			ID:        uuid.New(),
			TicketID:  ticket.ID,
			EventType: models.EventStatusChanged,
			ToStatus:  &to,
			ActorRole: models.RoleCustomer,
			ActorID:   in.CustomerID,
		}
		if err := repo.CreateTicketEvent(ctx, event); err != nil {
			return err
		}

		stored, err := repo.GetTicketByID(ctx, ticket.ID)
		if err != nil {
			return err
		}
		out = stored
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (s *TicketService) AddCustomerAttachment(ctx context.Context, in AddAttachmentInput) (*models.TicketAttachment, error) {
	fileType := strings.TrimSpace(strings.ToUpper(in.FileType))
	if fileType != "IMAGE" && fileType != "VIDEO" {
		return nil, models.NewAPIError("VALIDATION_ERROR", "file_type must be IMAGE or VIDEO", http.StatusBadRequest, nil)
	}
	if strings.TrimSpace(in.FileURL) == "" {
		return nil, models.NewAPIError("VALIDATION_ERROR", "file_url is required", http.StatusBadRequest, nil)
	}

	var out *models.TicketAttachment
	err := s.tx.WithTx(ctx, func(repo repository.Repository) error {
		ticket, err := repo.GetTicketByID(ctx, in.TicketID)
		if err != nil {
			return err
		}
		if ticket.CustomerID != in.CustomerID {
			return models.ErrForbidden
		}

		attachment := &models.TicketAttachment{
			ID:         uuid.New(),
			TicketID:   in.TicketID,
			FileType:   fileType,
			FileURL:    strings.TrimSpace(in.FileURL),
			UploadedBy: in.CustomerID,
		}
		if err := repo.CreateAttachment(ctx, attachment); err != nil {
			return err
		}
		attachments, err := repo.ListAttachments(ctx, in.TicketID)
		if err != nil {
			return err
		}
		if len(attachments) > 0 {
			lastID := attachments[len(attachments)-1].ID
			ticket.TicketAttachmentsID = &lastID
			if err := repo.UpdateTicket(ctx, ticket); err != nil {
				return err
			}
		}
		out = attachment
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (s *TicketService) CreateCustomerFeedback(ctx context.Context, in CreateFeedbackInput) (*models.TicketFeedback, error) {
	if in.Score < 1 || in.Score > 5 {
		return nil, models.NewAPIError("VALIDATION_ERROR", "score must be between 1 and 5", http.StatusBadRequest, nil)
	}

	var out *models.TicketFeedback
	err := s.tx.WithTx(ctx, func(repo repository.Repository) error {
		ticket, err := repo.GetTicketByIDForUpdate(ctx, in.TicketID)
		if err != nil {
			return err
		}
		if ticket.CustomerID != in.CustomerID {
			return models.ErrForbidden
		}
		if ticket.Status != models.TicketStatusResolved && ticket.Status != models.TicketStatusRejected {
			return models.NewAPIError("FEEDBACK_NOT_ALLOWED", "feedback is allowed only for RESOLVED or REJECTED tickets", http.StatusBadRequest, nil)
		}
		exists, err := repo.FeedbackExists(ctx, in.TicketID)
		if err != nil {
			return err
		}
		if exists {
			return models.NewAPIError("CONFLICT", "feedback already exists for this ticket", http.StatusConflict, nil)
		}

		feedback := &models.TicketFeedback{
			TicketID: in.TicketID,
			Score:    in.Score,
			Solved:   in.Solved,
			Comment:  trimStringPtr(in.Comment),
		}
		if err := repo.CreateFeedback(ctx, feedback); err != nil {
			return err
		}

		ticket.FeedbackScore = &in.Score
		if err := repo.UpdateTicket(ctx, ticket); err != nil {
			return err
		}

		stored, err := repo.GetFeedback(ctx, in.TicketID)
		if err != nil {
			return err
		}
		out = stored
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (s *TicketService) ListTechTickets(ctx context.Context, techID uuid.UUID, status *models.TicketStatus, assigned bool, page, pageSize int) ([]models.Ticket, models.Pagination, error) {
	items, total, err := s.tx.Repo().ListTechTickets(ctx, repository.TechTicketFilter{
		TechnicianID: techID,
		Status:       status,
		AssignedOnly: assigned,
		Page:         page,
		PageSize:     pageSize,
	})
	if err != nil {
		return nil, models.Pagination{}, err
	}
	return items, models.Pagination{Page: page, PageSize: pageSize, Total: total}, nil
}

func (s *TicketService) PatchTechTicket(ctx context.Context, ticketID uuid.UUID, in PatchTicketInput) (*models.Ticket, error) {
	if in.Status == nil && in.AssignedTechnicianID == nil && in.CloseReason == nil {
		return nil, models.NewAPIError("VALIDATION_ERROR", "at least one field must be provided", http.StatusBadRequest, nil)
	}

	var out *models.Ticket
	err := s.tx.WithTx(ctx, func(repo repository.Repository) error {
		ticket, err := repo.GetTicketByIDForUpdate(ctx, ticketID)
		if err != nil {
			return err
		}

		if in.AssignedTechnicianID != nil {
			oldAssigned := ticket.AssignedTechnicianID
			if !sameUUIDPtr(oldAssigned, in.AssignedTechnicianID) {
				ticket.AssignedTechnicianID = in.AssignedTechnicianID
				msg := "assigned technician updated"
				if err := repo.CreateTicketEvent(ctx, &models.TicketEvent{
					ID:        uuid.New(),
					TicketID:  ticket.ID,
					EventType: models.EventAssign,
					Message:   &msg,
					ActorRole: in.Actor.Role,
					ActorID:   in.Actor.ActorID,
				}); err != nil {
					return err
				}
			}
		}

		if in.Status != nil {
			if !in.Status.Valid() {
				return models.NewAPIError("VALIDATION_ERROR", "invalid status", http.StatusBadRequest, nil)
			}
			if !isValidTransition(ticket.Status, *in.Status) {
				return models.NewAPIError("INVALID_STATUS_TRANSITION", "invalid status transition", http.StatusBadRequest, map[string]any{"from": ticket.Status, "to": *in.Status})
			}
			if ticket.Status != *in.Status {
				from := ticket.Status
				to := *in.Status
				ticket.Status = to

				if err := repo.CreateTicketEvent(ctx, &models.TicketEvent{
					ID:         uuid.New(),
					TicketID:   ticket.ID,
					EventType:  models.EventStatusChanged,
					FromStatus: &from,
					ToStatus:   &to,
					ActorRole:  in.Actor.Role,
					ActorID:    in.Actor.ActorID,
				}); err != nil {
					return err
				}

				if to == models.TicketStatusResolved || to == models.TicketStatusRejected {
					reason := trimStringPtr(in.CloseReason)
					if reason == nil || *reason == "" {
						return models.NewAPIError("VALIDATION_ERROR", "close_reason is required when closing ticket", http.StatusBadRequest, nil)
					}
					now := time.Now().UTC()
					ticket.ClosedAt = &now
					ticket.CloseReason = reason
					if err := repo.CreateTicketEvent(ctx, &models.TicketEvent{
						ID:         uuid.New(),
						TicketID:   ticket.ID,
						EventType:  models.EventClose,
						FromStatus: &from,
						ToStatus:   &to,
						Message:    reason,
						ActorRole:  in.Actor.Role,
						ActorID:    in.Actor.ActorID,
					}); err != nil {
						return err
					}
				} else {
					ticket.ClosedAt = nil
					ticket.CloseReason = nil
				}
			}
		}

		if in.CloseReason != nil && in.Status == nil {
			if ticket.Status != models.TicketStatusResolved && ticket.Status != models.TicketStatusRejected {
				return models.NewAPIError("VALIDATION_ERROR", "close_reason can only be set for RESOLVED or REJECTED tickets", http.StatusBadRequest, nil)
			}
			reason := trimStringPtr(in.CloseReason)
			if reason == nil {
				return models.NewAPIError("VALIDATION_ERROR", "close_reason must not be empty", http.StatusBadRequest, nil)
			}
			ticket.CloseReason = reason
		}

		if err := repo.UpdateTicket(ctx, ticket); err != nil {
			return err
		}
		updated, err := repo.GetTicketByID(ctx, ticket.ID)
		if err != nil {
			return err
		}
		out = updated
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (s *TicketService) AddTechComment(ctx context.Context, in AddCommentInput) (*models.TicketEvent, error) {
	if strings.TrimSpace(in.Message) == "" {
		return nil, models.NewAPIError("VALIDATION_ERROR", "message is required", http.StatusBadRequest, nil)
	}

	var out *models.TicketEvent
	err := s.tx.WithTx(ctx, func(repo repository.Repository) error {
		_, err := repo.GetTicketByID(ctx, in.TicketID)
		if err != nil {
			return err
		}
		message := strings.TrimSpace(in.Message)
		event := &models.TicketEvent{
			ID:        uuid.New(),
			TicketID:  in.TicketID,
			EventType: models.EventComment,
			Message:   &message,
			ActorRole: models.RoleTechnician,
			ActorID:   in.TechID,
		}
		if err := repo.CreateTicketEvent(ctx, event); err != nil {
			return err
		}
		out = event
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

func sameUUIDPtr(a, b *uuid.UUID) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func trimStringPtr(v *string) *string {
	if v == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*v)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func isValidTransition(from, to models.TicketStatus) bool {
	if from == to {
		return true
	}
	allowed := map[models.TicketStatus]map[models.TicketStatus]bool{
		models.TicketStatusNew: {
			models.TicketStatusInReview: true,
		},
		models.TicketStatusInReview: {
			models.TicketStatusInProgress:   true,
			models.TicketStatusNeedMoreInfo: true,
			models.TicketStatusRejected:     true,
		},
		models.TicketStatusNeedMoreInfo: {
			models.TicketStatusInReview: true,
			models.TicketStatusRejected: true,
		},
		models.TicketStatusInProgress: {
			models.TicketStatusResolved:     true,
			models.TicketStatusNeedMoreInfo: true,
			models.TicketStatusRejected:     true,
		},
		models.TicketStatusResolved: {},
		models.TicketStatusRejected: {},
	}
	return allowed[from][to]
}
