package models

import "time"

type TicketStatus string

const (
	TicketStatusNew          TicketStatus = "NEW"
	TicketStatusInReview     TicketStatus = "IN_REVIEW"
	TicketStatusInProgress   TicketStatus = "IN_PROGRESS"
	TicketStatusNeedMoreInfo TicketStatus = "NEED_MORE_INFO"
	TicketStatusResolved     TicketStatus = "RESOLVED"
	TicketStatusRejected     TicketStatus = "REJECTED"
)

type EventType string

const (
	EventTypeStatusChanged EventType = "STATUS_CHANGED"
	EventTypeComment       EventType = "COMMENT"
	EventTypeAssign        EventType = "ASSIGN"
	EventTypeClose         EventType = "CLOSE"
)

type ActorRole string

const (
	ActorRoleCustomer   ActorRole = "CUSTOMER"
	ActorRoleTechnician ActorRole = "TECHNICIAN"
	ActorRoleAdmin      ActorRole = "ADMIN"
)

type Ticket struct {
	ID                   string       `json:"id"`
	TicketNumber         string       `json:"ticket_number"`
	CustomerID           string       `json:"customer_id"`
	DeviceSerial         string       `json:"device_serial"`
	Subject              string       `json:"subject"`
	Description          string       `json:"description"`
	Category             *string      `json:"category"`
	Status               TicketStatus `json:"status"`
	AssignedTechnicianID *string      `json:"assigned_technician_id"`
	CloseReason          *string      `json:"close_reason"`
	CreatedAt            time.Time    `json:"created_at"`
	UpdatedAt            time.Time    `json:"updated_at"`
	ClosedAt             *time.Time   `json:"closed_at"`
}

type TicketEvent struct {
	ID         string        `json:"id"`
	TicketID   string        `json:"ticket_id"`
	EventType  EventType     `json:"event_type"`
	FromStatus *TicketStatus `json:"from_status"`
	ToStatus   *TicketStatus `json:"to_status"`
	Message    *string       `json:"message"`
	ActorRole  ActorRole     `json:"actor_role"`
	ActorID    string        `json:"actor_id"`
	CreatedAt  time.Time     `json:"created_at"`
}

type TicketAttachment struct {
	ID         string    `json:"id"`
	TicketID   string    `json:"ticket_id"`
	FileType   string    `json:"file_type"`
	FileURL    string    `json:"file_url"`
	UploadedBy string    `json:"uploaded_by"`
	CreatedAt  time.Time `json:"created_at"`
}

type TicketFeedback struct {
	TicketID  string    `json:"ticket_id"`
	Score     int       `json:"score"`
	Solved    bool      `json:"solved"`
	Comment   *string   `json:"comment"`
	CreatedAt time.Time `json:"created_at"`
}

type Agency struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Province   string    `json:"province"`
	City       string    `json:"city"`
	Address    string    `json:"address"`
	Phone      string    `json:"phone"`
	Latitude   *float64  `json:"latitude"`
	Longitude  *float64  `json:"longitude"`
	IsOfficial bool      `json:"is_official"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
