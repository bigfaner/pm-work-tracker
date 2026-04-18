package handler

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"pm-work-tracker/backend/config"
	"pm-work-tracker/backend/internal/middleware"
	"pm-work-tracker/backend/internal/repository"
)

// Dependencies holds all services and configuration needed by the router.
// Handlers are wired here to avoid global state.
type Dependencies struct {
	Config     *config.LegacyConfig
	TeamRepo   repository.TeamRepo
	Auth       *AuthHandler
	Team       *TeamHandler
	MainItem   *MainItemHandler
	SubItem    *SubItemHandler
	Progress   *ProgressHandler
	ItemPool   *ItemPoolHandler
	View       *ViewHandler
	Report     *ReportHandler
	Admin      *AdminHandler
}

// SetupRouter creates a Gin engine with all route groups, middleware chains,
// and handler stubs registered.
func SetupRouter(deps *Dependencies) *gin.Engine {
	if deps.Config != nil && deps.Config.GinMode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// CORS middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     corsOrigins(deps),
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Health check — no auth required
	r.GET("/health", healthCheck)

	v1 := r.Group("/api/v1")

	// Auth routes (public login, authenticated logout)
	authGroup := v1.Group("/auth")
	{
		// Rate limit login: 10 req/min per IP
		authGroup.POST("/login", rateLimitMiddleware(10, time.Minute), deps.Auth.Login)
		authGroup.POST("/logout", middleware.AuthMiddleware(deps.Config.JWTSecret), deps.Auth.Logout)
	}

	// Team-scoped routes (require auth + team membership)
	teamsGroup := v1.Group("/teams/:teamId")
	teamsGroup.Use(
		middleware.AuthMiddleware(deps.Config.JWTSecret),
		middleware.TeamScopeMiddleware(deps.TeamRepo),
	)
	{
		// Team info
		teamsGroup.GET("", deps.Team.Get)
		teamsGroup.PUT("", middleware.RequireTeamRole("pm"), deps.Team.Update)
		teamsGroup.DELETE("", middleware.RequireTeamRole("pm"), deps.Team.Disband)

		// Members
		teamsGroup.GET("/members", deps.Team.ListMembers)
		teamsGroup.POST("/members", middleware.RequireTeamRole("pm"), deps.Team.InviteMember)
		teamsGroup.DELETE("/members/:userId", middleware.RequireTeamRole("pm"), deps.Team.RemoveMember)
		teamsGroup.PUT("/pm", middleware.RequireTeamRole("pm"), deps.Team.TransferPM)

		// Main items
		teamsGroup.POST("/main-items", deps.MainItem.Create)
		teamsGroup.GET("/main-items", deps.MainItem.List)
		teamsGroup.GET("/main-items/:itemId", deps.MainItem.Get)
		teamsGroup.PUT("/main-items/:itemId", deps.MainItem.Update)
		teamsGroup.POST("/main-items/:itemId/archive", deps.MainItem.Archive)

		// Sub items (under main items)
		teamsGroup.POST("/main-items/:itemId/sub-items", deps.SubItem.Create)
		teamsGroup.GET("/main-items/:itemId/sub-items", deps.SubItem.List)

		// Sub items (direct access)
		teamsGroup.GET("/sub-items/:subId", deps.SubItem.Get)
		teamsGroup.PUT("/sub-items/:subId", deps.SubItem.Update)
		teamsGroup.PUT("/sub-items/:subId/status", deps.SubItem.ChangeStatus)
		teamsGroup.PUT("/sub-items/:subId/assignee", deps.SubItem.Assign)

		// Progress records
		teamsGroup.POST("/sub-items/:subId/progress", deps.Progress.Append)
		teamsGroup.GET("/sub-items/:subId/progress", deps.Progress.List)
		teamsGroup.PATCH("/progress/:recordId/completion", deps.Progress.CorrectCompletion)

		// Item pool
		teamsGroup.POST("/item-pool", deps.ItemPool.Submit)
		teamsGroup.GET("/item-pool", deps.ItemPool.List)
		teamsGroup.GET("/item-pool/:poolId", deps.ItemPool.Get)
		teamsGroup.POST("/item-pool/:poolId/assign", middleware.RequireTeamRole("pm"), deps.ItemPool.Assign)
		teamsGroup.POST("/item-pool/:poolId/reject", middleware.RequireTeamRole("pm"), deps.ItemPool.Reject)

		// Views
		teamsGroup.GET("/views/weekly", deps.View.Weekly)
		teamsGroup.GET("/views/gantt", deps.View.Gantt)
		teamsGroup.GET("/views/table", deps.View.Table)
		teamsGroup.GET("/views/table/export", deps.View.ExportTable)

		// Reports
		teamsGroup.GET("/reports/weekly/preview", deps.Report.WeeklyPreview)
		teamsGroup.GET("/reports/weekly/export", deps.Report.WeeklyExport)
	}

	// Team list/create routes (outside :teamId group, auth only)
	v1.POST("/teams", middleware.AuthMiddleware(deps.Config.JWTSecret), deps.Team.Create)
	v1.GET("/teams", middleware.AuthMiddleware(deps.Config.JWTSecret), deps.Team.List)

	// Admin routes (superadmin only)
	adminGroup := v1.Group("/admin")
	adminGroup.Use(
		middleware.AuthMiddleware(deps.Config.JWTSecret),
		middleware.RequireRole("superadmin"),
	)
	{
		adminGroup.GET("/users", deps.Admin.ListUsers)
		adminGroup.PUT("/users/:userId/can-create-team", deps.Admin.UpdateCanCreateTeam)
		adminGroup.GET("/teams", deps.Admin.ListTeams)
	}

	return r
}

// healthCheck returns a simple status response with no auth.
func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// corsOrigins returns the list of allowed CORS origins from config.
// Falls back to "*" if none configured.
func corsOrigins(deps *Dependencies) []string {
	if deps.Config != nil && len(deps.Config.CORSOrigins) > 0 {
		return deps.Config.CORSOrigins
	}
	return []string{"*"}
}

// rateLimitMiddleware creates a per-IP rate limiter using golang.org/x/time/rate.
// limit is the number of requests allowed per interval.
func rateLimitMiddleware(limit int, interval time.Duration) gin.HandlerFunc {
	type visitor struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	var mu sync.Mutex
	visitors := make(map[string]*visitor)

	// Cleanup stale entries periodically
	go func() {
		for {
			time.Sleep(interval)
			mu.Lock()
			for ip, v := range visitors {
				if time.Since(v.lastSeen) > interval {
					delete(visitors, ip)
				}
			}
			mu.Unlock()
		}
	}()

	ratePerSecond := float64(limit) / interval.Seconds()

	return func(c *gin.Context) {
		ip := c.ClientIP()
		mu.Lock()
		v, exists := visitors[ip]
		if !exists {
			v = &visitor{limiter: rate.NewLimiter(rate.Limit(ratePerSecond), limit)}
			visitors[ip] = v
		}
		v.lastSeen = time.Now()
		mu.Unlock()

		if !v.limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code":    "RATE_LIMITED",
				"message": "too many requests",
			})
			return
		}
		c.Next()
	}
}
