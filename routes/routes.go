package routes

import (
	"itsplanned/common"
	"itsplanned/handlers"
	"itsplanned/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter(app *common.App) *gin.Engine {
	r := gin.Default()

	r.POST("/register", func(c *gin.Context) { handlers.Register(c, app.DB) })
	r.POST("/login", func(c *gin.Context) { handlers.Login(c, app.DB) })

	r.Use(middleware.AuthMiddleware())

	// r.GET("/profile", func(c *gin.Context) { handlers.GetProfile(c, app.DB) })

	r.GET("/events", func(c *gin.Context) { handlers.GetEvents(c, app.DB) })
	r.POST("/events", func(c *gin.Context) { handlers.CreateEvent(c, app.DB) })
	r.POST("/events/find_best_time_for_day", func(c *gin.Context) { handlers.FindBestTimeSlotsForDay(c, app.DB) })
	r.PUT("/events/:id/budget", func(c *gin.Context) { handlers.UpdateEventBudget(c, app.DB) })
	r.PUT("/events/:id", func(c *gin.Context) { handlers.UpdateEvent(c, app.DB) })
	// r.GET("/events/:id/budget", func(c *gin.Context) { handlers.GetEventBudget(c, app.DB) })
	r.GET("/events/:id/leaderboard", func(c *gin.Context) { handlers.GetEventLeaderboard(c, app.DB) })

	r.POST("/events/invite", func(c *gin.Context) { handlers.GenerateInviteLink(c, app.DB) })
	r.GET("/events/join/:invite_code", func(c *gin.Context) { handlers.JoinEvent(c, app.DB) })

	r.POST("/tasks", func(c *gin.Context) { handlers.CreateTask(c, app.DB) })
	r.PUT("/tasks/:id/assign", func(c *gin.Context) { handlers.AssignToTask(c, app.DB) })
	r.PUT("/tasks/:id/complete", func(c *gin.Context) { handlers.CompleteTask(c, app.DB) })

	r.GET("/auth/google", handlers.GetGoogleOAuthURL)
	r.GET("/auth/google/callback", handlers.GoogleOAuthCallback)
	r.POST("/auth/oauth/save", func(c *gin.Context) { handlers.SaveOAuthToken(c, app.DB) })

	r.GET("/calendar/import", func(c *gin.Context) { handlers.ImportCalendarEvents(c, app.DB) })

	return r
}
