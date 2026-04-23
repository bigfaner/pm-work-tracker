package handler

import (
	"io/fs"
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
	Config     *config.Config
	TeamRepo   repository.TeamRepo
	UserRepo   repository.UserRepo
	RoleRepo   repository.RoleRepo
	Auth       *AuthHandler
	Team       *TeamHandler
	MainItem   *MainItemHandler
	SubItem    *SubItemHandler
	Progress   *ProgressHandler
	ItemPool   *ItemPoolHandler
	View       *ViewHandler
	Report     *ReportHandler
	Admin      *AdminHandler
	Role       *RoleHandler
	Permission *PermissionHandler
}

// perm is a shorthand for creating a RequirePermission middleware with the deps' RoleRepo.
func (d *Dependencies) perm(code string) gin.HandlerFunc {
	return middleware.RequirePermission(code, d.RoleRepo)
}

// SetupRouter creates a Gin engine with all route groups, middleware chains,
// and handler stubs registered. fsys is the embedded frontend FS; pass nil
// to skip static/SPA routes (local dev / tests).
func SetupRouter(deps *Dependencies, fsys fs.FS) *gin.Engine {
	if deps.Config != nil && deps.Config.Server.GinMode == "release" {
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

	// Health check — no auth required, no prefix
	r.GET("/health", healthCheck)

	basePath := ""
	if deps.Config != nil {
		basePath = deps.Config.Server.BasePath
	}

	v1 := r.Group(basePath + "/v1")

	// Auth routes (public login, authenticated logout)
	authGroup := v1.Group("/auth")
	{
		// Rate limit login: 10 req/min per IP
		authGroup.POST("/login", rateLimitMiddleware(10, time.Minute), deps.Auth.Login)
		authGroup.POST("/logout", middleware.AuthMiddleware(deps.Config.Auth.JWTSecret, deps.UserRepo), deps.Auth.Logout)
	}

	// Team-scoped routes (require auth + team membership)
	teamsGroup := v1.Group("/teams/:teamId")
	teamsGroup.Use(
		middleware.AuthMiddleware(deps.Config.Auth.JWTSecret, deps.UserRepo),
		middleware.TeamScopeMiddleware(deps.TeamRepo, deps.RoleRepo),
	)
	{
		// Team info
		teamsGroup.GET("", deps.perm("team:read"), deps.Team.Get)
		teamsGroup.PUT("", deps.perm("team:update"), deps.Team.Update)
		teamsGroup.DELETE("", deps.perm("team:delete"), deps.Team.Disband)

		// Members
		teamsGroup.GET("/members", deps.Team.ListMembers)
		teamsGroup.GET("/search-users", deps.perm("team:invite"), deps.Team.SearchUsers)
		teamsGroup.POST("/members", deps.perm("team:invite"), deps.Team.InviteMember)
		teamsGroup.DELETE("/members/:userId", deps.perm("team:remove"), deps.Team.RemoveMember)
		teamsGroup.PUT("/members/:userId/role", deps.perm("team:invite"), deps.Team.UpdateMemberRole)
		teamsGroup.PUT("/pm", deps.perm("team:transfer"), deps.Team.TransferPM)

		// Main items
		teamsGroup.POST("/main-items", deps.perm("main_item:create"), deps.MainItem.Create)
		teamsGroup.GET("/main-items", deps.perm("main_item:read"), deps.MainItem.List)
		teamsGroup.GET("/main-items/:itemId", deps.perm("main_item:read"), deps.MainItem.Get)
		teamsGroup.PUT("/main-items/:itemId", deps.perm("main_item:update"), deps.MainItem.Update)
		teamsGroup.PUT("/main-items/:itemId/status", deps.perm("main_item:change_status"), deps.MainItem.ChangeStatus)
		teamsGroup.GET("/main-items/:itemId/available-transitions", deps.perm("main_item:read"), deps.MainItem.AvailableTransitions)
		teamsGroup.POST("/main-items/:itemId/archive", deps.perm("main_item:archive"), deps.MainItem.Archive)

		// Sub items (under main items)
		teamsGroup.POST("/main-items/:itemId/sub-items", deps.perm("sub_item:create"), deps.SubItem.Create)
		teamsGroup.GET("/main-items/:itemId/sub-items", deps.perm("sub_item:read"), deps.SubItem.List)

		// Sub items (direct access)
		teamsGroup.GET("/sub-items/:subId", deps.perm("sub_item:read"), deps.SubItem.Get)
		teamsGroup.PUT("/sub-items/:subId", deps.perm("sub_item:update"), deps.SubItem.Update)
		teamsGroup.PUT("/sub-items/:subId/status", deps.perm("sub_item:change_status"), deps.SubItem.ChangeStatus)
		teamsGroup.GET("/sub-items/:subId/available-transitions", deps.perm("sub_item:read"), deps.SubItem.AvailableTransitions)
		teamsGroup.PUT("/sub-items/:subId/assignee", deps.perm("sub_item:assign"), deps.SubItem.Assign)

		// Progress records
		teamsGroup.POST("/sub-items/:subId/progress", deps.perm("progress:create"), deps.Progress.Append)
		teamsGroup.GET("/sub-items/:subId/progress", deps.perm("progress:read"), deps.Progress.List)
		teamsGroup.PATCH("/progress/:recordId/completion", deps.perm("progress:update"), deps.Progress.CorrectCompletion)

		// Item pool
		teamsGroup.POST("/item-pool", deps.perm("item_pool:submit"), deps.ItemPool.Submit)
		teamsGroup.GET("/item-pool", deps.perm("sub_item:read"), deps.ItemPool.List)
		teamsGroup.GET("/item-pool/:poolId", deps.perm("sub_item:read"), deps.ItemPool.Get)
		teamsGroup.POST("/item-pool/:poolId/assign", deps.perm("item_pool:review"), deps.ItemPool.Assign)
		teamsGroup.POST("/item-pool/:poolId/convert-to-main", deps.perm("item_pool:review"), deps.ItemPool.ConvertToMain)
		teamsGroup.POST("/item-pool/:poolId/reject", deps.perm("item_pool:review"), deps.ItemPool.Reject)

		// Views
		teamsGroup.GET("/views/weekly", deps.perm("view:weekly"), deps.View.Weekly)
		teamsGroup.GET("/views/gantt", deps.perm("view:gantt"), deps.View.Gantt)
		teamsGroup.GET("/views/table", deps.perm("view:table"), deps.View.Table)
		teamsGroup.GET("/views/table/export", deps.perm("view:table"), deps.View.ExportTable)

		// Reports
		teamsGroup.GET("/reports/weekly/preview", deps.perm("report:export"), deps.Report.WeeklyPreview)
		teamsGroup.GET("/reports/weekly/export", deps.perm("report:export"), deps.Report.WeeklyExport)
	}

	// Team list/create routes (outside :teamId group, auth only)
	authMW := middleware.AuthMiddleware(deps.Config.Auth.JWTSecret, deps.UserRepo)
	v1.POST("/teams", authMW, deps.perm("team:create"), deps.Team.Create)
	v1.GET("/teams", authMW, deps.Team.List)

	// Admin routes (permission-gated)
	adminGroup := v1.Group("/admin")
	adminGroup.Use(authMW)
	{
		adminGroup.GET("/users", deps.perm("user:read"), deps.Admin.ListUsers)
		adminGroup.POST("/users", deps.perm("user:manage_role"), deps.Admin.CreateUser)
		adminGroup.GET("/users/:userId", deps.perm("user:read"), deps.Admin.GetUser)
		adminGroup.PUT("/users/:userId", deps.perm("user:update"), deps.Admin.UpdateUser)
		adminGroup.PUT("/users/:userId/status", deps.perm("user:update"), deps.Admin.ToggleUserStatus)
		adminGroup.GET("/teams", deps.perm("user:read"), deps.Admin.ListTeams)

		// Role management
		adminGroup.GET("/roles", deps.perm("user:manage_role"), deps.Role.ListRoles)
		adminGroup.POST("/roles", deps.perm("user:manage_role"), deps.Role.CreateRole)
		adminGroup.GET("/roles/:id", deps.perm("user:manage_role"), deps.Role.GetRole)
		adminGroup.PUT("/roles/:id", deps.perm("user:manage_role"), deps.Role.UpdateRole)
		adminGroup.DELETE("/roles/:id", deps.perm("user:manage_role"), deps.Role.DeleteRole)

		// Permission code registry
		adminGroup.GET("/permissions", deps.perm("user:manage_role"), deps.Permission.ListPermissionCodes)
	}

	// User-facing: current user's permissions (auth only, no permission code required)
	v1.GET("/me/permissions", authMW, deps.Permission.GetUserPermissions)

	// Static file routes — only registered when fsys is provided (production/embedded mode)
	if fsys != nil {
		bp := r.Group(basePath)
		bp.GET("/assets/*filepath", ServeStatic(fsys))
		bp.GET("/", ServeSPA(fsys))
		r.NoRoute(ServeSPA(fsys))
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
	if deps.Config != nil && len(deps.Config.CORS.Origins) > 0 {
		return deps.Config.CORS.Origins
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
