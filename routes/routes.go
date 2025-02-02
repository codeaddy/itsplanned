package routes

import (
	"itsplanned/common"
	"itsplanned/handlers"

	"github.com/gin-gonic/gin"
)

func SetupRouter(app *common.App) *gin.Engine {
	r := gin.Default()

	r.POST("/register", func(c *gin.Context) { handlers.Register(c, app.DB) })
	r.POST("/login", func(c *gin.Context) { handlers.Login(c, app.DB) })
	r.GET("/profile", func(c *gin.Context) { handlers.GetProfile(c, app.DB) })

	// r.POST("/events", func(c *gin.Context) { handlers.CreateEvent(c, app.DB) })
	// r.PUT("/events/:id/budget", func(c *gin.Context) { handlers.UpdateEventBudget(c, app.DB) })
	// r.GET("/events/:id/budget", func(c *gin.Context) { handlers.GetEventBudget(c, app.DB) })

	// r.POST("/tasks", func(c *gin.Context) { handlers.CreateTask(c, app.DB) })
	// r.PUT("/tasks/:id/complete", func(c *gin.Context) { handlers.CompleteTask(c, app.DB) })
	// r.PUT("/tasks/:id/assign", func(c *gin.Context) { handlers.AssignToTask(c, app.DB) })

	r.GET("/auth/google", handlers.GetGoogleOAuthURL)
	r.GET("/auth/google/callback", handlers.GoogleOAuthCallback)
	r.POST("/calendar/google/events", handlers.FetchGoogleCalendarEventsHandler)

	return r
}
