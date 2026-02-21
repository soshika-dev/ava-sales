package repositories

import (
	"context"
	"fmt"

	"ava-sales/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TicketFilter struct {
	CustomerID           string
	Status               string
	AssignedTechnicianID string
}

type TicketPatch struct {
	Status               *models.TicketStatus
	AssignedTechnicianID *string
	CloseReason          *string
}

type TicketRepository struct {
	db *pgxpool.Pool
}

func NewTicketRepository(db *pgxpool.Pool) *TicketRepository { return &TicketRepository{db: db} }

func (r *TicketRepository) ListCustomerTickets(ctx context.Context, customerID string) ([]models.Ticket, error) {
	query := `SELECT id, ticket_number, customer_id, device_serial, subject, description, category, status, assigned_technician_id, close_reason, created_at, updated_at, closed_at
		FROM tickets WHERE customer_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTickets(rows)
}

func (r *TicketRepository) ListTechTickets(ctx context.Context, f TicketFilter) ([]models.Ticket, error) {
	args := []any{}
	where := "WHERE 1=1"
	i := 1
	if f.Status != "" {
		where += fmt.Sprintf(" AND status = $%d", i)
		args = append(args, f.Status)
		i++
	}
	if f.AssignedTechnicianID != "" {
		where += fmt.Sprintf(" AND assigned_technician_id = $%d", i)
		args = append(args, f.AssignedTechnicianID)
		i++
	}
	query := `SELECT id, ticket_number, customer_id, device_serial, subject, description, category, status, assigned_technician_id, close_reason, created_at, updated_at, closed_at
		FROM tickets ` + where + ` ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTickets(rows)
}

func (r *TicketRepository) GetTicket(ctx context.Context, id string) (*models.Ticket, error) {
	query := `SELECT id, ticket_number, customer_id, device_serial, subject, description, category, status, assigned_technician_id, close_reason, created_at, updated_at, closed_at
		FROM tickets WHERE id = $1`
	var t models.Ticket
	if err := r.db.QueryRow(ctx, query, id).Scan(&t.ID, &t.TicketNumber, &t.CustomerID, &t.DeviceSerial, &t.Subject, &t.Description, &t.Category, &t.Status, &t.AssignedTechnicianID, &t.CloseReason, &t.CreatedAt, &t.UpdatedAt, &t.ClosedAt); err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TicketRepository) CreateTicket(ctx context.Context, t *models.Ticket, actorID string) (*models.Ticket, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	if err := tx.QueryRow(ctx, `INSERT INTO tickets (ticket_number, customer_id, device_serial, subject, description, category, status)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		RETURNING id, created_at, updated_at`,
		t.TicketNumber, t.CustomerID, t.DeviceSerial, t.Subject, t.Description, t.Category, t.Status,
	).Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt); err != nil {
		return nil, err
	}

	if _, err := tx.Exec(ctx, `INSERT INTO ticket_events (ticket_id,event_type,from_status,to_status,actor_role,actor_id)
		VALUES ($1,$2,$3,$4,$5,$6)`, t.ID, models.EventTypeStatusChanged, nil, t.Status, models.ActorRoleCustomer, actorID); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return t, nil
}

func (r *TicketRepository) AddAttachment(ctx context.Context, a *models.TicketAttachment) error {
	return r.db.QueryRow(ctx, `INSERT INTO ticket_attachments (ticket_id,file_type,file_url,uploaded_by)
		VALUES ($1,$2,$3,$4) RETURNING id,created_at`, a.TicketID, a.FileType, a.FileURL, a.UploadedBy).Scan(&a.ID, &a.CreatedAt)
}

func (r *TicketRepository) AddFeedback(ctx context.Context, f *models.TicketFeedback) error {
	return r.db.QueryRow(ctx, `INSERT INTO ticket_feedback (ticket_id,score,solved,comment)
		VALUES ($1,$2,$3,$4) RETURNING created_at`, f.TicketID, f.Score, f.Solved, f.Comment).Scan(&f.CreatedAt)
}

func (r *TicketRepository) ListEvents(ctx context.Context, ticketID string) ([]models.TicketEvent, error) {
	rows, err := r.db.Query(ctx, `SELECT id,ticket_id,event_type,from_status,to_status,message,actor_role,actor_id,created_at
		FROM ticket_events WHERE ticket_id = $1 ORDER BY created_at ASC`, ticketID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	res := make([]models.TicketEvent, 0)
	for rows.Next() {
		var e models.TicketEvent
		if err := rows.Scan(&e.ID, &e.TicketID, &e.EventType, &e.FromStatus, &e.ToStatus, &e.Message, &e.ActorRole, &e.ActorID, &e.CreatedAt); err != nil {
			return nil, err
		}
		res = append(res, e)
	}
	return res, rows.Err()
}

func (r *TicketRepository) UpdateTicketByTechnician(ctx context.Context, ticketID, actorID string, patch TicketPatch) (*models.Ticket, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	current, err := r.getTicketForUpdate(ctx, tx, ticketID)
	if err != nil {
		return nil, err
	}

	nextStatus := current.Status
	if patch.Status != nil {
		nextStatus = *patch.Status
	}
	nextAssigned := current.AssignedTechnicianID
	if patch.AssignedTechnicianID != nil {
		nextAssigned = patch.AssignedTechnicianID
	}
	nextCloseReason := current.CloseReason
	var nextClosedAt *string
	if patch.Status != nil {
		if nextStatus == models.TicketStatusResolved || nextStatus == models.TicketStatusRejected {
			nextCloseReason = patch.CloseReason
			now := "now"
			nextClosedAt = &now
		} else {
			nextCloseReason = nil
			nextClosedAt = nil
		}
	}

	updateQuery := `UPDATE tickets SET status=$1, assigned_technician_id=$2, close_reason=$3, closed_at=`
	args := []any{nextStatus, nextAssigned, nextCloseReason, ticketID}
	if nextClosedAt != nil {
		updateQuery += `NOW()`
	} else {
		updateQuery += `NULL`
	}
	updateQuery += ` WHERE id=$4 RETURNING id,ticket_number,customer_id,device_serial,subject,description,category,status,assigned_technician_id,close_reason,created_at,updated_at,closed_at`
	var updated models.Ticket
	if err := tx.QueryRow(ctx, updateQuery, args...).Scan(&updated.ID, &updated.TicketNumber, &updated.CustomerID, &updated.DeviceSerial, &updated.Subject, &updated.Description, &updated.Category, &updated.Status, &updated.AssignedTechnicianID, &updated.CloseReason, &updated.CreatedAt, &updated.UpdatedAt, &updated.ClosedAt); err != nil {
		return nil, err
	}

	if patch.Status != nil && current.Status != *patch.Status {
		if _, err := tx.Exec(ctx, `INSERT INTO ticket_events (ticket_id,event_type,from_status,to_status,actor_role,actor_id)
			VALUES ($1,$2,$3,$4,$5,$6)`, ticketID, models.EventTypeStatusChanged, current.Status, *patch.Status, models.ActorRoleTechnician, actorID); err != nil {
			return nil, err
		}
		if *patch.Status == models.TicketStatusResolved || *patch.Status == models.TicketStatusRejected {
			msg := patch.CloseReason
			if _, err := tx.Exec(ctx, `INSERT INTO ticket_events (ticket_id,event_type,message,actor_role,actor_id)
				VALUES ($1,$2,$3,$4,$5)`, ticketID, models.EventTypeClose, msg, models.ActorRoleTechnician, actorID); err != nil {
				return nil, err
			}
		}
	}

	if patch.AssignedTechnicianID != nil && (current.AssignedTechnicianID == nil || *current.AssignedTechnicianID != *patch.AssignedTechnicianID) {
		msg := fmt.Sprintf("Assigned to technician %s", *patch.AssignedTechnicianID)
		if _, err := tx.Exec(ctx, `INSERT INTO ticket_events (ticket_id,event_type,message,actor_role,actor_id)
			VALUES ($1,$2,$3,$4,$5)`, ticketID, models.EventTypeAssign, msg, models.ActorRoleTechnician, actorID); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &updated, nil
}

func (r *TicketRepository) AddComment(ctx context.Context, ticketID, actorID, message string) error {
	_, err := r.db.Exec(ctx, `INSERT INTO ticket_events (ticket_id,event_type,message,actor_role,actor_id)
		VALUES ($1,$2,$3,$4,$5)`, ticketID, models.EventTypeComment, message, models.ActorRoleTechnician, actorID)
	return err
}

func (r *TicketRepository) getTicketForUpdate(ctx context.Context, tx pgx.Tx, ticketID string) (*models.Ticket, error) {
	var t models.Ticket
	query := `SELECT id, ticket_number, customer_id, device_serial, subject, description, category, status, assigned_technician_id, close_reason, created_at, updated_at, closed_at
		FROM tickets WHERE id=$1 FOR UPDATE`
	if err := tx.QueryRow(ctx, query, ticketID).Scan(&t.ID, &t.TicketNumber, &t.CustomerID, &t.DeviceSerial, &t.Subject, &t.Description, &t.Category, &t.Status, &t.AssignedTechnicianID, &t.CloseReason, &t.CreatedAt, &t.UpdatedAt, &t.ClosedAt); err != nil {
		return nil, err
	}
	return &t, nil
}

func scanTickets(rows pgx.Rows) ([]models.Ticket, error) {
	tickets := make([]models.Ticket, 0)
	for rows.Next() {
		var t models.Ticket
		if err := rows.Scan(&t.ID, &t.TicketNumber, &t.CustomerID, &t.DeviceSerial, &t.Subject, &t.Description, &t.Category, &t.Status, &t.AssignedTechnicianID, &t.CloseReason, &t.CreatedAt, &t.UpdatedAt, &t.ClosedAt); err != nil {
			return nil, err
		}
		tickets = append(tickets, t)
	}
	return tickets, rows.Err()
}
