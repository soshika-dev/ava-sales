package router

import (
	"ava-sales/internal/handlers"

	"github.com/gin-gonic/gin"
)

func New(agency *handlers.AgencyHandler, ticket *handlers.TicketHandler) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	api := r.Group("/api")
	{
		api.GET("/agencies", agency.List)
		api.GET("/agencies/:id", agency.GetByID)

		api.GET("/tickets", ticket.ListCustomerTickets)
		api.GET("/tickets/:id", ticket.GetTicket)
		api.POST("/tickets", ticket.CreateTicket)
		api.POST("/tickets/:id/attachments", ticket.AddAttachment)
		api.POST("/tickets/:id/feedback", ticket.AddFeedback)
		api.GET("/tickets/:id/events", ticket.ListEvents)

		tech := api.Group("/tech")
		tech.GET("/tickets", ticket.ListTechTickets)
		tech.PATCH("/tickets/:id", ticket.PatchTechTicket)
		tech.POST("/tickets/:id/comment", ticket.AddTechComment)
	}

	return r
}
