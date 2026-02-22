package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"ava-sales/internal/middleware"
	"ava-sales/internal/models"
	"ava-sales/internal/repository"
	"ava-sales/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	ticketService *service.TicketService
	agencyService *service.AgencyService
}

type createTicketRequest struct {
	DeviceSerial string  `json:"device_serial" binding:"required"`
	Subject      string  `json:"subject" binding:"required"`
	Description  string  `json:"description" binding:"required"`
	Category     *string `json:"category"`
}

type addAttachmentRequest struct {
	FileType string `json:"file_type" binding:"required,oneof=IMAGE VIDEO"`
	FileURL  string `json:"file_url" binding:"required,url"`
}

type createFeedbackRequest struct {
	Score   int     `json:"score" binding:"required,min=1,max=5"`
	Solved  *bool   `json:"solved" binding:"required"`
	Comment *string `json:"comment"`
}

type patchTechTicketRequest struct {
	Status               *string `json:"status" binding:"omitempty,oneof=NEW IN_REVIEW IN_PROGRESS NEED_MORE_INFO RESOLVED REJECTED"`
	AssignedTechnicianID *string `json:"assigned_technician_id"`
	CloseReason          *string `json:"close_reason"`
}

type addCommentRequest struct {
	Message string `json:"message" binding:"required"`
}

func (h *Handler) ListAgencies(c *gin.Context) {
	page, pageSize, err := parsePagination(c)
	if err != nil {
		writeError(c, err)
		return
	}

	filter := repository.AgencyFilter{Page: page, PageSize: pageSize}
	if v := strings.TrimSpace(c.Query("city")); v != "" {
		filter.City = &v
	}
	if v := strings.TrimSpace(c.Query("province")); v != "" {
		filter.Province = &v
	}
	if v := strings.TrimSpace(c.Query("q")); v != "" {
		filter.Query = &v
	}

	items, meta, svcErr := h.agencyService.List(c.Request.Context(), filter)
	if svcErr != nil {
		writeError(c, svcErr)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items, "meta": meta})
}

func (h *Handler) ListCustomerTickets(c *gin.Context) {
	customerID, err := middleware.ParseCustomerID(c)
	if err != nil {
		writeError(c, err)
		return
	}
	page, pageSize, pageErr := parsePagination(c)
	if pageErr != nil {
		writeError(c, pageErr)
		return
	}

	var status *models.TicketStatus
	if raw := strings.TrimSpace(c.Query("status")); raw != "" {
		s := models.TicketStatus(strings.ToUpper(raw))
		if !s.Valid() {
			writeError(c, models.NewAPIError("VALIDATION_ERROR", "invalid status", http.StatusBadRequest, nil))
			return
		}
		status = &s
	}

	items, meta, svcErr := h.ticketService.ListCustomerTickets(c.Request.Context(), customerID, status, page, pageSize)
	if svcErr != nil {
		writeError(c, svcErr)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items, "meta": meta})
}

func (h *Handler) GetCustomerTicket(c *gin.Context) {
	customerID, err := middleware.ParseCustomerID(c)
	if err != nil {
		writeError(c, err)
		return
	}
	ticketID, parseErr := uuid.Parse(c.Param("id"))
	if parseErr != nil {
		writeError(c, models.NewAPIError("VALIDATION_ERROR", "invalid ticket id", http.StatusBadRequest, nil))
		return
	}

	item, svcErr := h.ticketService.GetCustomerTicketDetail(c.Request.Context(), customerID, ticketID)
	if svcErr != nil {
		writeError(c, svcErr)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": item})
}

func (h *Handler) CreateCustomerTicket(c *gin.Context) {
	customerID, err := middleware.ParseCustomerID(c)
	if err != nil {
		writeError(c, err)
		return
	}

	var req createTicketRequest
	if bindErr := c.ShouldBindJSON(&req); bindErr != nil {
		writeError(c, models.NewAPIError("VALIDATION_ERROR", "invalid request body", http.StatusBadRequest, bindErr.Error()))
		return
	}

	item, svcErr := h.ticketService.CreateCustomerTicket(c.Request.Context(), service.CreateTicketInput{
		CustomerID:   customerID,
		DeviceSerial: req.DeviceSerial,
		Subject:      req.Subject,
		Description:  req.Description,
		Category:     req.Category,
	})
	if svcErr != nil {
		writeError(c, svcErr)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": item})
}

func (h *Handler) AddCustomerAttachment(c *gin.Context) {
	customerID, err := middleware.ParseCustomerID(c)
	if err != nil {
		writeError(c, err)
		return
	}
	ticketID, parseErr := uuid.Parse(c.Param("id"))
	if parseErr != nil {
		writeError(c, models.NewAPIError("VALIDATION_ERROR", "invalid ticket id", http.StatusBadRequest, nil))
		return
	}

	var req addAttachmentRequest
	if bindErr := c.ShouldBindJSON(&req); bindErr != nil {
		writeError(c, models.NewAPIError("VALIDATION_ERROR", "invalid request body", http.StatusBadRequest, bindErr.Error()))
		return
	}

	item, svcErr := h.ticketService.AddCustomerAttachment(c.Request.Context(), service.AddAttachmentInput{
		TicketID:   ticketID,
		CustomerID: customerID,
		FileType:   req.FileType,
		FileURL:    req.FileURL,
	})
	if svcErr != nil {
		writeError(c, svcErr)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": item})
}

func (h *Handler) CreateCustomerFeedback(c *gin.Context) {
	customerID, err := middleware.ParseCustomerID(c)
	if err != nil {
		writeError(c, err)
		return
	}
	ticketID, parseErr := uuid.Parse(c.Param("id"))
	if parseErr != nil {
		writeError(c, models.NewAPIError("VALIDATION_ERROR", "invalid ticket id", http.StatusBadRequest, nil))
		return
	}

	var req createFeedbackRequest
	if bindErr := c.ShouldBindJSON(&req); bindErr != nil {
		writeError(c, models.NewAPIError("VALIDATION_ERROR", "invalid request body", http.StatusBadRequest, bindErr.Error()))
		return
	}

	item, svcErr := h.ticketService.CreateCustomerFeedback(c.Request.Context(), service.CreateFeedbackInput{
		TicketID:   ticketID,
		CustomerID: customerID,
		Score:      req.Score,
		Solved:     *req.Solved,
		Comment:    req.Comment,
	})
	if svcErr != nil {
		writeError(c, svcErr)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": item})
}

func (h *Handler) ListTechTickets(c *gin.Context) {
	techID, err := middleware.ParseTechnicianID(c)
	if err != nil {
		writeError(c, err)
		return
	}
	page, pageSize, pageErr := parsePagination(c)
	if pageErr != nil {
		writeError(c, pageErr)
		return
	}

	var status *models.TicketStatus
	if raw := strings.TrimSpace(c.Query("status")); raw != "" {
		s := models.TicketStatus(strings.ToUpper(raw))
		if !s.Valid() {
			writeError(c, models.NewAPIError("VALIDATION_ERROR", "invalid status", http.StatusBadRequest, nil))
			return
		}
		status = &s
	}

	assigned := false
	if rawAssigned := strings.TrimSpace(c.Query("assigned")); rawAssigned != "" {
		parsed, parseBoolErr := strconv.ParseBool(rawAssigned)
		if parseBoolErr != nil {
			writeError(c, models.NewAPIError("VALIDATION_ERROR", "assigned must be true or false", http.StatusBadRequest, nil))
			return
		}
		assigned = parsed
	}

	items, meta, svcErr := h.ticketService.ListTechTickets(c.Request.Context(), techID, status, assigned, page, pageSize)
	if svcErr != nil {
		writeError(c, svcErr)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items, "meta": meta})
}

func (h *Handler) PatchTechTicket(c *gin.Context) {
	techID, err := middleware.ParseTechnicianID(c)
	if err != nil {
		writeError(c, err)
		return
	}
	ticketID, parseErr := uuid.Parse(c.Param("id"))
	if parseErr != nil {
		writeError(c, models.NewAPIError("VALIDATION_ERROR", "invalid ticket id", http.StatusBadRequest, nil))
		return
	}

	var req patchTechTicketRequest
	if bindErr := c.ShouldBindJSON(&req); bindErr != nil {
		writeError(c, models.NewAPIError("VALIDATION_ERROR", "invalid request body", http.StatusBadRequest, bindErr.Error()))
		return
	}

	var status *models.TicketStatus
	if req.Status != nil {
		s := models.TicketStatus(strings.ToUpper(*req.Status))
		status = &s
	}
	var assignedTechID *uuid.UUID
	if req.AssignedTechnicianID != nil {
		id, idErr := uuid.Parse(*req.AssignedTechnicianID)
		if idErr != nil {
			writeError(c, models.NewAPIError("VALIDATION_ERROR", "assigned_technician_id must be UUID", http.StatusBadRequest, nil))
			return
		}
		assignedTechID = &id
	}

	actor, actorErr := middleware.ParseActor(c, models.RoleTechnician, techID)
	if actorErr != nil {
		writeError(c, actorErr)
		return
	}
	if actor.Role != models.RoleTechnician && actor.Role != models.RoleAdmin {
		writeError(c, models.ErrForbidden)
		return
	}

	updated, svcErr := h.ticketService.PatchTechTicket(c.Request.Context(), ticketID, service.PatchTicketInput{
		Status:               status,
		AssignedTechnicianID: assignedTechID,
		CloseReason:          req.CloseReason,
		Actor:                actor,
	})
	if svcErr != nil {
		writeError(c, svcErr)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": updated})
}

func (h *Handler) AddTechComment(c *gin.Context) {
	techID, err := middleware.ParseTechnicianID(c)
	if err != nil {
		writeError(c, err)
		return
	}
	ticketID, parseErr := uuid.Parse(c.Param("id"))
	if parseErr != nil {
		writeError(c, models.NewAPIError("VALIDATION_ERROR", "invalid ticket id", http.StatusBadRequest, nil))
		return
	}

	var req addCommentRequest
	if bindErr := c.ShouldBindJSON(&req); bindErr != nil {
		writeError(c, models.NewAPIError("VALIDATION_ERROR", "invalid request body", http.StatusBadRequest, bindErr.Error()))
		return
	}

	event, svcErr := h.ticketService.AddTechComment(c.Request.Context(), service.AddCommentInput{
		TicketID: ticketID,
		TechID:   techID,
		Message:  req.Message,
	})
	if svcErr != nil {
		writeError(c, svcErr)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": event})
}
