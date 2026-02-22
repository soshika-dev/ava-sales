package models

import (
	"time"

	"github.com/google/uuid"
)

type TicketStatus string

const (
	TicketStatusNew          TicketStatus = "NEW"
	TicketStatusInReview     TicketStatus = "IN_REVIEW"
	TicketStatusInProgress   TicketStatus = "IN_PROGRESS"
	TicketStatusNeedMoreInfo TicketStatus = "NEED_MORE_INFO"
	TicketStatusResolved     TicketStatus = "RESOLVED"
	TicketStatusRejected     TicketStatus = "REJECTED"
)

func (s TicketStatus) Valid() bool {
	switch s {
	case TicketStatusNew, TicketStatusInReview, TicketStatusInProgress, TicketStatusNeedMoreInfo, TicketStatusResolved, TicketStatusRejected:
		return true
	default:
		return false
	}
}

type Ticket struct {
	ID                   uuid.UUID  `json:"id"`
	TicketNumber         string     `json:"ticket_number"`
	CustomerID           uuid.UUID  `json:"customer_id"`
	DeviceSerial         string     `json:"device_serial"`
	Subject              string     `json:"subject"`
	Description          string     `json:"description"`
	Category             *string    `json:"category,omitempty"`
	Status               TicketStatus `json:"status"`
	AssignedTechnicianID *uuid.UUID `json:"assigned_technician_id,omitempty"`
	CloseReason          *string    `json:"close_reason,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
	TicketAttachmentsID  *uuid.UUID `json:"ticket_attachments_id,omitempty"`
	FeedbackScore        *int       `json:"feedback_score,omitempty"`
	ClosedAt             *time.Time `json:"closed_at,omitempty"`
}

type TicketEventType string

const (
	EventStatusChanged TicketEventType = "STATUS_CHANGED"
	EventComment       TicketEventType = "COMMENT"
	EventAssign        TicketEventType = "ASSIGN"
	EventClose         TicketEventType = "CLOSE"
)

type ActorRole string

const (
	RoleCustomer   ActorRole = "CUSTOMER"
	RoleTechnician ActorRole = "TECHNICIAN"
	RoleAdmin      ActorRole = "ADMIN"
)

type TicketEvent struct {
	ID         uuid.UUID    `json:"id"`
	TicketID   uuid.UUID    `json:"ticket_id"`
	EventType  TicketEventType `json:"event_type"`
	FromStatus *TicketStatus `json:"from_status,omitempty"`
	ToStatus   *TicketStatus `json:"to_status,omitempty"`
	Message    *string      `json:"message,omitempty"`
	ActorRole  ActorRole    `json:"actor_role"`
	ActorID    uuid.UUID    `json:"actor_id"`
	CreatedAt  time.Time    `json:"created_at"`
}

type TicketAttachment struct {
	ID         uuid.UUID `json:"id"`
	TicketID   uuid.UUID `json:"ticket_id"`
	FileType   string    `json:"file_type"`
	FileURL    string    `json:"file_url"`
	UploadedBy uuid.UUID `json:"uploaded_by"`
	CreatedAt  time.Time `json:"created_at"`
}

type TicketFeedback struct {
	TicketID   uuid.UUID `json:"ticket_id"`
	Score      int       `json:"score"`
	Solved     bool      `json:"solved"`
	Comment    *string   `json:"comment,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

type TicketDetail struct {
	Ticket
	Attachments []TicketAttachment `json:"attachments"`
	Events      []TicketEvent      `json:"events"`
	Feedback    *TicketFeedback    `json:"feedback,omitempty"`
}
