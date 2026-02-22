package models

type Pagination struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
	Total    int `json:"total"`
}
