package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"ava-sales/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type dbtx interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type postgresRepository struct {
	db dbtx
}

func (r *postgresRepository) NextTicketNumber(ctx context.Context) (string, error) {
	var ticketNumber string
	err := r.db.QueryRow(ctx, `SELECT 'TK-' || LPAD(nextval('ticket_number_seq')::text, 5, '0')`).Scan(&ticketNumber)
	if err != nil {
		return "", mapDBError(err)
	}
	return ticketNumber, nil
}

func (r *postgresRepository) CreateTicket(ctx context.Context, ticket *models.Ticket) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO tickets (
			id, ticket_number, customer_id, device_serial, subject, description, category, status,
			assigned_technician_id, close_reason, ticket_attachments_id, feedback_score, closed_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
	`,
		ticket.ID, ticket.TicketNumber, ticket.CustomerID, ticket.DeviceSerial, ticket.Subject, ticket.Description,
		ticket.Category, ticket.Status, ticket.AssignedTechnicianID, ticket.CloseReason, ticket.TicketAttachmentsID,
		ticket.FeedbackScore, ticket.ClosedAt,
	)
	return mapDBError(err)
}

func (r *postgresRepository) GetTicketByID(ctx context.Context, ticketID uuid.UUID) (*models.Ticket, error) {
	q := `
		SELECT id, ticket_number, customer_id, device_serial, subject, description, category, status,
			assigned_technician_id, close_reason, created_at, updated_at, ticket_attachments_id, feedback_score, closed_at
		FROM tickets WHERE id = $1
	`
	row := r.db.QueryRow(ctx, q, ticketID)
	ticket, err := scanTicket(row)
	if err != nil {
		return nil, mapDBError(err)
	}
	return ticket, nil
}

func (r *postgresRepository) GetTicketByIDForUpdate(ctx context.Context, ticketID uuid.UUID) (*models.Ticket, error) {
	q := `
		SELECT id, ticket_number, customer_id, device_serial, subject, description, category, status,
			assigned_technician_id, close_reason, created_at, updated_at, ticket_attachments_id, feedback_score, closed_at
		FROM tickets WHERE id = $1
		FOR UPDATE
	`
	row := r.db.QueryRow(ctx, q, ticketID)
	ticket, err := scanTicket(row)
	if err != nil {
		return nil, mapDBError(err)
	}
	return ticket, nil
}

func (r *postgresRepository) ListCustomerTickets(ctx context.Context, filter CustomerTicketFilter) ([]models.Ticket, int, error) {
	baseWhere := " WHERE customer_id = $1"
	args := []any{filter.CustomerID}
	argPos := 2
	if filter.Status != nil {
		baseWhere += fmt.Sprintf(" AND status = $%d", argPos)
		args = append(args, *filter.Status)
		argPos++
	}

	countSQL := "SELECT COUNT(1) FROM tickets" + baseWhere
	var total int
	if err := r.db.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, mapDBError(err)
	}

	page, pageSize := normalizePage(filter.Page, filter.PageSize)
	offset := (page - 1) * pageSize
	listSQL := `
		SELECT id, ticket_number, customer_id, device_serial, subject, description, category, status,
			assigned_technician_id, close_reason, created_at, updated_at, ticket_attachments_id, feedback_score, closed_at
		FROM tickets` + baseWhere + fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argPos, argPos+1)

	listArgs := append(args, pageSize, offset)
	rows, err := r.db.Query(ctx, listSQL, listArgs...)
	if err != nil {
		return nil, 0, mapDBError(err)
	}
	defer rows.Close()

	items := make([]models.Ticket, 0, pageSize)
	for rows.Next() {
		t, err := scanTicket(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, *t)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *postgresRepository) ListTechTickets(ctx context.Context, filter TechTicketFilter) ([]models.Ticket, int, error) {
	whereParts := []string{"1=1"}
	args := make([]any, 0, 4)
	argPos := 1

	if filter.Status != nil {
		whereParts = append(whereParts, fmt.Sprintf("status = $%d", argPos))
		args = append(args, *filter.Status)
		argPos++
	}
	if filter.AssignedOnly {
		whereParts = append(whereParts, fmt.Sprintf("assigned_technician_id = $%d", argPos))
		args = append(args, filter.TechnicianID)
		argPos++
	}

	whereSQL := " WHERE " + strings.Join(whereParts, " AND ")
	countSQL := "SELECT COUNT(1) FROM tickets" + whereSQL
	var total int
	if err := r.db.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, mapDBError(err)
	}

	page, pageSize := normalizePage(filter.Page, filter.PageSize)
	offset := (page - 1) * pageSize
	listSQL := `
		SELECT id, ticket_number, customer_id, device_serial, subject, description, category, status,
			assigned_technician_id, close_reason, created_at, updated_at, ticket_attachments_id, feedback_score, closed_at
		FROM tickets` + whereSQL + fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argPos, argPos+1)

	listArgs := append(args, pageSize, offset)
	rows, err := r.db.Query(ctx, listSQL, listArgs...)
	if err != nil {
		return nil, 0, mapDBError(err)
	}
	defer rows.Close()

	items := make([]models.Ticket, 0, pageSize)
	for rows.Next() {
		t, err := scanTicket(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, *t)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *postgresRepository) UpdateTicket(ctx context.Context, ticket *models.Ticket) error {
	_, err := r.db.Exec(ctx, `
		UPDATE tickets
		SET status = $1,
			assigned_technician_id = $2,
			close_reason = $3,
			ticket_attachments_id = $4,
			feedback_score = $5,
			closed_at = $6
		WHERE id = $7
	`, ticket.Status, ticket.AssignedTechnicianID, ticket.CloseReason, ticket.TicketAttachmentsID, ticket.FeedbackScore, ticket.ClosedAt, ticket.ID)
	return mapDBError(err)
}

func (r *postgresRepository) CreateTicketEvent(ctx context.Context, event *models.TicketEvent) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO ticket_events (id, ticket_id, event_type, from_status, to_status, message, actor_role, actor_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
	`, event.ID, event.TicketID, event.EventType, event.FromStatus, event.ToStatus, event.Message, event.ActorRole, event.ActorID)
	return mapDBError(err)
}

func (r *postgresRepository) ListTicketEvents(ctx context.Context, ticketID uuid.UUID) ([]models.TicketEvent, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, ticket_id, event_type, from_status, to_status, message, actor_role, actor_id, created_at
		FROM ticket_events
		WHERE ticket_id = $1
		ORDER BY created_at ASC
	`, ticketID)
	if err != nil {
		return nil, mapDBError(err)
	}
	defer rows.Close()

	items := make([]models.TicketEvent, 0)
	for rows.Next() {
		var e models.TicketEvent
		var fromStatus sql.NullString
		var toStatus sql.NullString
		var message sql.NullString
		if err := rows.Scan(&e.ID, &e.TicketID, &e.EventType, &fromStatus, &toStatus, &message, &e.ActorRole, &e.ActorID, &e.CreatedAt); err != nil {
			return nil, err
		}
		if fromStatus.Valid {
			s := models.TicketStatus(fromStatus.String)
			e.FromStatus = &s
		}
		if toStatus.Valid {
			s := models.TicketStatus(toStatus.String)
			e.ToStatus = &s
		}
		if message.Valid {
			e.Message = &message.String
		}
		items = append(items, e)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *postgresRepository) CreateAttachment(ctx context.Context, attachment *models.TicketAttachment) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO ticket_attachments (id, ticket_id, file_type, file_url, uploaded_by)
		VALUES ($1,$2,$3,$4,$5)
	`, attachment.ID, attachment.TicketID, attachment.FileType, attachment.FileURL, attachment.UploadedBy)
	return mapDBError(err)
}

func (r *postgresRepository) ListAttachments(ctx context.Context, ticketID uuid.UUID) ([]models.TicketAttachment, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, ticket_id, file_type, file_url, uploaded_by, created_at
		FROM ticket_attachments
		WHERE ticket_id = $1
		ORDER BY created_at ASC
	`, ticketID)
	if err != nil {
		return nil, mapDBError(err)
	}
	defer rows.Close()

	items := make([]models.TicketAttachment, 0)
	for rows.Next() {
		var a models.TicketAttachment
		if err := rows.Scan(&a.ID, &a.TicketID, &a.FileType, &a.FileURL, &a.UploadedBy, &a.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, a)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *postgresRepository) CreateFeedback(ctx context.Context, feedback *models.TicketFeedback) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO ticket_feedback (ticket_id, score, solved, comment)
		VALUES ($1,$2,$3,$4)
	`, feedback.TicketID, feedback.Score, feedback.Solved, feedback.Comment)
	return mapDBError(err)
}

func (r *postgresRepository) GetFeedback(ctx context.Context, ticketID uuid.UUID) (*models.TicketFeedback, error) {
	var f models.TicketFeedback
	var comment sql.NullString
	err := r.db.QueryRow(ctx, `
		SELECT ticket_id, score, solved, comment, created_at
		FROM ticket_feedback
		WHERE ticket_id = $1
	`, ticketID).Scan(&f.TicketID, &f.Score, &f.Solved, &comment, &f.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, mapDBError(err)
	}
	if comment.Valid {
		f.Comment = &comment.String
	}
	return &f, nil
}

func (r *postgresRepository) FeedbackExists(ctx context.Context, ticketID uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM ticket_feedback WHERE ticket_id = $1)`, ticketID).Scan(&exists)
	if err != nil {
		return false, mapDBError(err)
	}
	return exists, nil
}

func (r *postgresRepository) ListAgencies(ctx context.Context, filter AgencyFilter) ([]models.Agency, int, error) {
	whereParts := []string{"1=1"}
	args := make([]any, 0, 5)
	argPos := 1

	if filter.City != nil && strings.TrimSpace(*filter.City) != "" {
		whereParts = append(whereParts, fmt.Sprintf("city = $%d", argPos))
		args = append(args, strings.TrimSpace(*filter.City))
		argPos++
	}
	if filter.Province != nil && strings.TrimSpace(*filter.Province) != "" {
		whereParts = append(whereParts, fmt.Sprintf("province = $%d", argPos))
		args = append(args, strings.TrimSpace(*filter.Province))
		argPos++
	}
	if filter.Query != nil && strings.TrimSpace(*filter.Query) != "" {
		whereParts = append(whereParts, fmt.Sprintf("(name ILIKE $%d OR address ILIKE $%d)", argPos, argPos))
		args = append(args, "%"+strings.TrimSpace(*filter.Query)+"%")
		argPos++
	}

	whereSQL := " WHERE " + strings.Join(whereParts, " AND ")
	countSQL := "SELECT COUNT(1) FROM agencies" + whereSQL
	var total int
	if err := r.db.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, mapDBError(err)
	}

	page, pageSize := normalizePage(filter.Page, filter.PageSize)
	offset := (page - 1) * pageSize
	listSQL := `
		SELECT id, name, province, city, address, phone, latitude, longitude, version, created_at, updated_at
		FROM agencies` + whereSQL + fmt.Sprintf(" ORDER BY name ASC LIMIT $%d OFFSET $%d", argPos, argPos+1)
	listArgs := append(args, pageSize, offset)

	rows, err := r.db.Query(ctx, listSQL, listArgs...)
	if err != nil {
		return nil, 0, mapDBError(err)
	}
	defer rows.Close()

	items := make([]models.Agency, 0, pageSize)
	for rows.Next() {
		var a models.Agency
		if err := rows.Scan(
			&a.ID, &a.Name, &a.Province, &a.City, &a.Address, &a.Phone,
			&a.Latitude, &a.Longitude, &a.Version, &a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		items = append(items, a)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanTicket(s scanner) (*models.Ticket, error) {
	var t models.Ticket
	var category sql.NullString
	var assignedTech pgtype.UUID
	var closeReason sql.NullString
	var ticketAttachmentsID pgtype.UUID
	var feedbackScore sql.NullInt32
	var closedAt sql.NullTime

	err := s.Scan(
		&t.ID, &t.TicketNumber, &t.CustomerID, &t.DeviceSerial, &t.Subject, &t.Description, &category, &t.Status,
		&assignedTech, &closeReason, &t.CreatedAt, &t.UpdatedAt, &ticketAttachmentsID, &feedbackScore, &closedAt,
	)
	if err != nil {
		return nil, err
	}

	if category.Valid {
		t.Category = &category.String
	}
	if assignedTech.Valid {
		parsed, err := uuid.FromBytes(assignedTech.Bytes[:])
		if err == nil {
			t.AssignedTechnicianID = &parsed
		}
	}
	if closeReason.Valid {
		t.CloseReason = &closeReason.String
	}
	if ticketAttachmentsID.Valid {
		parsed, err := uuid.FromBytes(ticketAttachmentsID.Bytes[:])
		if err == nil {
			t.TicketAttachmentsID = &parsed
		}
	}
	if feedbackScore.Valid {
		score := int(feedbackScore.Int32)
		t.FeedbackScore = &score
	}
	if closedAt.Valid {
		timeValue := closedAt.Time.UTC()
		t.ClosedAt = &timeValue
	}

	return &t, nil
}

func normalizePage(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

func mapDBError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return models.ErrNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return models.NewAPIError("CONFLICT", "resource already exists", 409, pgErr.ConstraintName)
		case "23503":
			return models.NewAPIError("CONSTRAINT_ERROR", "invalid reference", 400, pgErr.ConstraintName)
		case "23514":
			return models.NewAPIError("CONSTRAINT_ERROR", "constraint validation failed", 400, pgErr.ConstraintName)
		}
	}
	return err
}
