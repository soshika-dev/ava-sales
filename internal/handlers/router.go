package handlers

import (
	"ava-sales/internal/middleware"
	"ava-sales/internal/service"
	"github.com/gin-gonic/gin"
)

func NewRouter(ticketSvc *service.TicketService, agencySvc *service.AgencyService) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestID(), middleware.RequestLogger())

	h := &Handler{ticketService: ticketSvc, agencyService: agencySvc}

	api := r.Group("/api")
	{
		api.GET("/agencies", h.ListAgencies)

		api.GET("/tickets", h.ListCustomerTickets)
		api.GET("/tickets/:id", h.GetCustomerTicket)
		api.POST("/tickets", h.CreateCustomerTicket)
		api.POST("/tickets/:id/attachments", h.AddCustomerAttachment)
		api.POST("/tickets/:id/feedback", h.CreateCustomerFeedback)

		api.GET("/tech/tickets", h.ListTechTickets)
		api.PATCH("/tech/tickets/:id", h.PatchTechTicket)
		api.POST("/tech/tickets/:id/comment", h.AddTechComment)
	}

	return r
}
