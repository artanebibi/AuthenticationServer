package handlers

import (
	"AuthServer/internal/domain/roles"
	"AuthServer/internal/middleware"
	"AuthServer/internal/repository"
	domain "AuthServer/internal/service"
	"net/http"

	db "AuthServer/internal/database"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var (
	database db.Service = db.New()

	userRepo     = repository.NewUserRepository(database)
	userRoleRepo = repository.NewUserRoleRepository(database)
	projectRepo  = repository.NewProjectRepository(database)
	jitRepo      = repository.NewJITRequestRepository(database)

	tokenService   domain.ITokenService   = domain.NewTokenService()
	hashService    domain.IHashService    = domain.NewHashService()
	userService    domain.IUserService    = domain.NewUserService(userRepo)
	projectService domain.IProjectService = domain.NewProjectService(projectRepo)
	rbacService    *domain.RBACService    = domain.NewRBACService(userRoleRepo)
	jitService     *domain.JITService     = domain.NewJITService(jitRepo, userRoleRepo)
)

type Server struct {
	Port     int
	Db       db.Service
	UserRepo repository.IUserRepository
}

func (s *Server) RegisterRoutes() http.Handler {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		AllowCredentials: true,
	}))

	// user interface
	r.LoadHTMLGlob("ui/templates/*")

	r.GET("/health", s.healthHandler)

	r.GET("/home", func(c *gin.Context) {
		c.HTML(200, "home.html", nil)
	})

	r.POST("api/login", s.Login)
	r.POST("api/register", s.Register)
	r.POST("/api/verification-code", s.VerificationCode)
	r.POST("/api/resend-verification", s.ResendVerificationCode)

	// user
	r.GET("/api/me", s.GetUserData)
	r.GET("/api/me/roles",
		middleware.RequireRole(rbacService, tokenService, roles.RoleUser, ""),
		s.GetMyRoles,
	)

	// project
	r.GET("/api/projects", s.GetAllProjects)
	r.GET("/api/projects/:id", s.GetProjectById)
	r.POST("/api/projects",
		middleware.RequireRole(rbacService, tokenService, roles.RoleProjectEditor, ""),
		s.CreateProject,
	)
	r.PUT("/api/projects/:id",
		middleware.RequireRole(rbacService, tokenService, roles.RoleProjectEditor, "id"),
		s.UpdateProject,
	)
	r.DELETE("/api/projects/:id",
		middleware.RequireRole(rbacService, tokenService, roles.RoleProjectEditor, "id"),
		s.DeleteProject,
	)

	// role assignment (Admin/Manager only)
	r.POST("/api/roles/global",
		middleware.RequireAnyRole(rbacService, tokenService, []roles.Role{roles.RoleAdmin, roles.RoleManager}, ""),
		s.AssignGlobalRole,
	)
	r.POST("/api/roles/resource",
		middleware.RequireAnyRole(rbacService, tokenService, []roles.Role{roles.RoleAdmin, roles.RoleManager}, ""),
		s.AssignResourceRole,
	)
	r.POST("/api/roles/temporary",
		middleware.RequireAnyRole(rbacService, tokenService, []roles.Role{roles.RoleAdmin, roles.RoleManager}, ""),
		s.AssignTemporaryRole,
	)
	r.GET("/api/users/:id/roles",
		middleware.RequireAnyRole(rbacService, tokenService, []roles.Role{roles.RoleAdmin, roles.RoleManager}, ""),
		s.GetUserRoles,
	)
	r.DELETE("/api/roles/:id",
		middleware.RequireAnyRole(rbacService, tokenService, []roles.Role{roles.RoleAdmin, roles.RoleManager}, ""),
		s.RevokeUserRole,
	)

	// JIT requests
	r.POST("/api/jit-requests",
		middleware.RequireRole(rbacService, tokenService, roles.RoleUser, ""),
		s.CreateJITRequest,
	)
	r.GET("/api/jit-requests/me",
		middleware.RequireRole(rbacService, tokenService, roles.RoleUser, ""),
		s.GetMyJITRequests,
	)
	r.GET("/api/jit-requests",
		middleware.RequireAnyRole(rbacService, tokenService, []roles.Role{roles.RoleAdmin, roles.RoleManager}, ""),
		s.GetJITRequests,
	)
	r.PATCH("/api/jit-requests/:id/approve",
		middleware.RequireAnyRole(rbacService, tokenService, []roles.Role{roles.RoleAdmin, roles.RoleManager}, ""),
		s.ApproveJITRequest,
	)
	r.PATCH("/api/jit-requests/:id/reject",
		middleware.RequireAnyRole(rbacService, tokenService, []roles.Role{roles.RoleAdmin, roles.RoleManager}, ""),
		s.RejectJITRequest,
	)

	return r
}
