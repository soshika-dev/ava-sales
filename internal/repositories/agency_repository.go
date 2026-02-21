package repositories

import (
	"context"
	"fmt"
	"strings"

	"ava-sales/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type AgencyFilter struct {
	City     string
	Province string
	Query    string
	Limit    int
	Offset   int
}

type AgencyRepository struct {
	db *pgxpool.Pool
}

func NewAgencyRepository(db *pgxpool.Pool) *AgencyRepository {
	return &AgencyRepository{db: db}
}

func (r *AgencyRepository) List(ctx context.Context, filter AgencyFilter) ([]models.Agency, error) {
	args := make([]any, 0)
	predicates := []string{"is_active = true"}
	idx := 1

	if filter.City != "" {
		predicates = append(predicates, fmt.Sprintf("city = $%d", idx))
		args = append(args, filter.City)
		idx++
	}
	if filter.Province != "" {
		predicates = append(predicates, fmt.Sprintf("province = $%d", idx))
		args = append(args, filter.Province)
		idx++
	}
	if filter.Query != "" {
		predicates = append(predicates, fmt.Sprintf("(name ILIKE $%d OR address ILIKE $%d)", idx, idx))
		args = append(args, "%"+filter.Query+"%")
		idx++
	}

	query := `SELECT id, name, province, city, address, phone, latitude, longitude, is_official, is_active, created_at, updated_at
		FROM agencies
		WHERE ` + strings.Join(predicates, " AND ") + fmt.Sprintf(" ORDER BY name ASC LIMIT $%d OFFSET $%d", idx, idx+1)
	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	agencies := make([]models.Agency, 0)
	for rows.Next() {
		var agency models.Agency
		if err := rows.Scan(
			&agency.ID, &agency.Name, &agency.Province, &agency.City, &agency.Address, &agency.Phone,
			&agency.Latitude, &agency.Longitude, &agency.IsOfficial, &agency.IsActive, &agency.CreatedAt, &agency.UpdatedAt,
		); err != nil {
			return nil, err
		}
		agencies = append(agencies, agency)
	}

	return agencies, rows.Err()
}

func (r *AgencyRepository) GetByID(ctx context.Context, id string) (*models.Agency, error) {
	var agency models.Agency
	query := `SELECT id, name, province, city, address, phone, latitude, longitude, is_official, is_active, created_at, updated_at
		FROM agencies WHERE id = $1`
	if err := r.db.QueryRow(ctx, query, id).Scan(
		&agency.ID, &agency.Name, &agency.Province, &agency.City, &agency.Address, &agency.Phone,
		&agency.Latitude, &agency.Longitude, &agency.IsOfficial, &agency.IsActive, &agency.CreatedAt, &agency.UpdatedAt,
	); err != nil {
		return nil, err
	}

	return &agency, nil
}
