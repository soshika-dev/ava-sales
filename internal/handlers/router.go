package handlers

import (
	"ava-sales/internal/middleware"
	"ava-sales/internal/repository"
	"ava-sales/internal/service"
	"github.com/gin-gonic/gin"
)

func NewRouter(ticketSvc *service.TicketService, agencySvc *service.AgencyService, tx repository.TxManager, jwtSecret string) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestID(), middleware.RequestLogger())

	h := &Handler{ticketService: ticketSvc, agencyService: agencySvc}

	api := r.Group("/api")
	{
		api.GET("/agencies", h.ListAgencies)

		customer := api.Group("/tickets")
		customer.Use(middleware.JWTAuth(tx, jwtSecret), middleware.RequireCustomer())
		customer.GET("", h.ListCustomerTickets)
		customer.GET("/:id", h.GetCustomerTicket)
		customer.POST("", h.CreateCustomerTicket)
		customer.POST("/:id/attachments", h.AddCustomerAttachment)
		customer.POST("/:id/feedback", h.CreateCustomerFeedback)

		tech := api.Group("/tech/tickets")
		tech.Use(middleware.JWTAuth(tx, jwtSecret), middleware.RequireTechOrAdmin())
		tech.GET("", h.ListTechTickets)
		tech.PATCH("/:id", h.PatchTechTicket)
		tech.POST("/:id/comment", h.AddTechComment)
	}

	return r
}
