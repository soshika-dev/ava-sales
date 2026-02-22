package service

import (
	"context"

	"ava-sales/internal/models"
	"ava-sales/internal/repository"
)

type AgencyService struct {
	tx repository.TxManager
}

func NewAgencyService(tx repository.TxManager) *AgencyService {
	return &AgencyService{tx: tx}
}

func (s *AgencyService) List(ctx context.Context, filter repository.AgencyFilter) ([]models.Agency, models.Pagination, error) {
	items, total, err := s.tx.Repo().ListAgencies(ctx, filter)
	if err != nil {
		return nil, models.Pagination{}, err
	}
	return items, models.Pagination{Page: filter.Page, PageSize: filter.PageSize, Total: total}, nil
}
