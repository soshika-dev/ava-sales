package handlers

import (
	"net/http"

	"ava-sales/internal/models"
	"ava-sales/internal/repositories"
	"ava-sales/internal/response"
	"ava-sales/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type TicketHandler struct {
	repo    *repositories.TicketRepository
	service *services.TicketService
}

func NewTicketHandler(repo *repositories.TicketRepository, service *services.TicketService) *TicketHandler {
	return &TicketHandler{repo: repo, service: service}
}

type createTicketRequest struct {
	CustomerID   string  `json:"customer_id" binding:"required,uuid"`
	DeviceSerial string  `json:"device_serial" binding:"required"`
	Subject      string  `json:"subject" binding:"required"`
	Description  string  `json:"description" binding:"required"`
	Category     *string `json:"category"`
}

type attachmentRequest struct {
	FileType   string `json:"file_type" binding:"required,oneof=IMAGE VIDEO"`
	FileURL    string `json:"file_url" binding:"required,url"`
	UploadedBy string `json:"uploaded_by" binding:"required,uuid"`
}

type feedbackRequest struct {
	Score   int     `json:"score" binding:"required,min=1,max=5"`
	Solved  bool    `json:"solved" binding:"required"`
	Comment *string `json:"comment"`
}

type techPatchRequest struct {
	Status               *models.TicketStatus `json:"status" binding:"omitempty,oneof=NEW IN_REVIEW IN_PROGRESS NEED_MORE_INFO RESOLVED REJECTED"`
	AssignedTechnicianID *string              `json:"assigned_technician_id" binding:"omitempty,uuid"`
	CloseReason          *string              `json:"close_reason"`
}

type commentRequest struct {
	ActorID string `json:"actor_id" binding:"required,uuid"`
	Message string `json:"message" binding:"required"`
}

func (h *TicketHandler) ListCustomerTickets(c *gin.Context) {
	customerID := c.Query("customer_id")
	if _, err := uuid.Parse(customerID); err != nil {
		response.JSONError(c, http.StatusBadRequest, "validation_error", "customer_id query param must be valid UUID")
		return
	}
	items, err := h.repo.ListCustomerTickets(c.Request.Context(), customerID)
	if err != nil {
		response.JSONError(c, http.StatusInternalServerError, "internal_error", "failed to fetch tickets")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}

func (h *TicketHandler) GetTicket(c *gin.Context) {
	item, err := h.repo.GetTicket(c.Request.Context(), c.Param("id"))
	if err != nil {
		if err == pgx.ErrNoRows {
			response.JSONError(c, http.StatusNotFound, "not_found", "ticket not found")
			return
		}
		response.JSONError(c, http.StatusInternalServerError, "internal_error", "failed to fetch ticket")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": item})
}

func (h *TicketHandler) CreateTicket(c *gin.Context) {
	var req createTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.JSONError(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	item, err := h.service.CreateTicket(c.Request.Context(), &models.Ticket{
		CustomerID: req.CustomerID, DeviceSerial: req.DeviceSerial, Subject: req.Subject, Description: req.Description, Category: req.Category,
	})
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			response.JSONError(c, http.StatusConflict, "conflict", "ticket number collision, retry")
			return
		}
		response.JSONError(c, http.StatusInternalServerError, "internal_error", "failed to create ticket")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": item})
}

func (h *TicketHandler) AddAttachment(c *gin.Context) {
	var req attachmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.JSONError(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	item := &models.TicketAttachment{TicketID: c.Param("id"), FileType: req.FileType, FileURL: req.FileURL, UploadedBy: req.UploadedBy}
	if err := h.repo.AddAttachment(c.Request.Context(), item); err != nil {
		response.JSONError(c, http.StatusInternalServerError, "internal_error", "failed to add attachment")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": item})
}

func (h *TicketHandler) AddFeedback(c *gin.Context) {
	var req feedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.JSONError(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	item := &models.TicketFeedback{TicketID: c.Param("id"), Score: req.Score, Solved: req.Solved, Comment: req.Comment}
	if err := h.repo.AddFeedback(c.Request.Context(), item); err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			response.JSONError(c, http.StatusConflict, "conflict", "feedback already exists for this ticket")
			return
		}
		response.JSONError(c, http.StatusInternalServerError, "internal_error", "failed to add feedback")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": item})
}

func (h *TicketHandler) ListEvents(c *gin.Context) {
	items, err := h.repo.ListEvents(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.JSONError(c, http.StatusInternalServerError, "internal_error", "failed to fetch events")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}

func (h *TicketHandler) ListTechTickets(c *gin.Context) {
	items, err := h.repo.ListTechTickets(c.Request.Context(), repositories.TicketFilter{
		Status: c.Query("status"), AssignedTechnicianID: c.Query("assigned_technician_id"),
	})
	if err != nil {
		response.JSONError(c, http.StatusInternalServerError, "internal_error", "failed to fetch tech tickets")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}

func (h *TicketHandler) PatchTechTicket(c *gin.Context) {
	actorID := c.Query("actor_id")
	if _, err := uuid.Parse(actorID); err != nil {
		response.JSONError(c, http.StatusBadRequest, "validation_error", "actor_id query param must be valid UUID")
		return
	}
	current, err := h.repo.GetTicket(c.Request.Context(), c.Param("id"))
	if err != nil {
		if err == pgx.ErrNoRows {
			response.JSONError(c, http.StatusNotFound, "not_found", "ticket not found")
			return
		}
		response.JSONError(c, http.StatusInternalServerError, "internal_error", "failed to fetch ticket")
		return
	}

	var req techPatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.JSONError(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	patch := repositories.TicketPatch{Status: req.Status, AssignedTechnicianID: req.AssignedTechnicianID, CloseReason: req.CloseReason}
	if err := h.service.ValidatePatch(current, patch); err != nil {
		response.JSONError(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	updated, err := h.repo.UpdateTicketByTechnician(c.Request.Context(), c.Param("id"), actorID, patch)
	if err != nil {
		response.JSONError(c, http.StatusInternalServerError, "internal_error", "failed to patch ticket")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": updated})
}

func (h *TicketHandler) AddTechComment(c *gin.Context) {
	var req commentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.JSONError(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	if err := h.repo.AddComment(c.Request.Context(), c.Param("id"), req.ActorID, req.Message); err != nil {
		response.JSONError(c, http.StatusInternalServerError, "internal_error", "failed to add comment")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "comment added"})
}
