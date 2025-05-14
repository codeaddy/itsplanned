package routes

import (
	"itsplanned/common"
	"itsplanned/handlers"
	"itsplanned/middleware"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRouter(app *common.App) *gin.Engine {
	r := gin.Default()

	// Public routes (no middleware)
	r.POST("/register", func(c *gin.Context) { handlers.Register(c, app.DB) })
	r.POST("/login", func(c *gin.Context) { handlers.Login(c, app.DB) })
	r.POST("/password/reset-request", func(c *gin.Context) { handlers.RequestPasswordReset(c, app.DB) })
	r.POST("/password/reset", func(c *gin.Context) { handlers.ResetPassword(c, app.DB) })
	r.GET("/password/reset-redirect", func(c *gin.Context) { handlers.HandlePasswordResetRedirect(c, app.DB) })

	// Swagger documentation route (public)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Add public route for deeplink redirection - no authentication required
	r.GET("/events/redirect/:invite_code", func(c *gin.Context) { handlers.DeepLinkRedirect(c, app.DB) })

	// Protected routes
	protected := r.Group("/")
	protected.Use(middleware.AuthMiddleware())

	// User profile routes
	protected.GET("/profile", func(c *gin.Context) { handlers.GetProfile(c, app.DB) })
	protected.PUT("/profile", func(c *gin.Context) { handlers.UpdateProfile(c, app.DB) })
	protected.POST("/logout", func(c *gin.Context) { handlers.Logout(c, app.DB) })

	// Event routes
	protected.GET("/events", func(c *gin.Context) { handlers.GetEvents(c, app.DB) })
	protected.GET("/events/:id", func(c *gin.Context) { handlers.GetEvent(c, app.DB) })
	protected.POST("/events", func(c *gin.Context) { handlers.CreateEvent(c, app.DB) })
	protected.POST("/events/find_best_time_for_day", func(c *gin.Context) { handlers.FindBestTimeSlotsForDay(c, app.DB) })
	protected.PUT("/events/:id", func(c *gin.Context) { handlers.UpdateEvent(c, app.DB) })
	protected.DELETE("/events/:id", func(c *gin.Context) { handlers.DeleteEvent(c, app.DB) })
	protected.GET("/events/:id/leaderboard", func(c *gin.Context) { handlers.GetEventLeaderboard(c, app.DB) })
	protected.GET("/events/:id/participants", func(c *gin.Context) { handlers.GetEventParticipants(c, app.DB) })
	protected.GET("/events/:id/budget", func(c *gin.Context) { handlers.GetEventBudget(c, app.DB) })

	// Event invitation routes
	protected.POST("/events/invite", func(c *gin.Context) { handlers.GenerateInviteLink(c, app.DB) })
	protected.GET("/events/join/:invite_code", func(c *gin.Context) { handlers.JoinEvent(c, app.DB) })
	protected.DELETE("/events/:id/leave", func(c *gin.Context) { handlers.LeaveEvent(c, app.DB) })
	// Add route for joining an event via query parameter (for iOS app deeplink handling)
	protected.GET("/events/join", func(c *gin.Context) { handlers.JoinEvent(c, app.DB) })

	// Task routes
	protected.GET("/tasks", func(c *gin.Context) { handlers.GetTasks(c, app.DB) })
	protected.GET("/tasks/:id", func(c *gin.Context) { handlers.GetTask(c, app.DB) })
	protected.POST("/tasks", func(c *gin.Context) { handlers.CreateTask(c, app.DB) })
	protected.PUT("/tasks/:id", func(c *gin.Context) { handlers.UpdateTask(c, app.DB) })
	protected.DELETE("/tasks/:id", func(c *gin.Context) { handlers.DeleteTask(c, app.DB) })
	protected.PUT("/tasks/:id/assign", func(c *gin.Context) { handlers.AssignToTask(c, app.DB) })
	protected.PUT("/tasks/:id/complete", func(c *gin.Context) { handlers.CompleteTask(c, app.DB) })

	// Task status events routes
	protected.GET("/task-status-events/unread", func(c *gin.Context) { handlers.GetUnreadTaskStatusEvents(c, app.DB) })

	// AI Assistant routes
	protected.POST("/ai/message", func(c *gin.Context) { handlers.SendToYandexGPT(c, app.DB) })

	// Google Calendar integration
	protected.GET("/auth/google", handlers.GetGoogleOAuthURL)
	protected.GET("/auth/google/callback", handlers.GoogleOAuthCallback)
	protected.POST("/auth/oauth/save", func(c *gin.Context) { handlers.SaveOAuthToken(c, app.DB) })
	protected.DELETE("/auth/oauth/delete", func(c *gin.Context) { handlers.DeleteOAuthToken(c, app.DB) })
	protected.GET("/calendar/import", func(c *gin.Context) { handlers.ImportCalendarEvents(c, app.DB) })

	// OAuth web to app redirection - make this public
	r.GET("/auth/web-to-app", handlers.WebToAppRedirect)

	return r
}
