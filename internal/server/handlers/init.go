package handlers

import (
	"AuthServer/internal/repository"
	domain "AuthServer/internal/service"
	"net/http"

	db "AuthServer/internal/database"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var (
	database     db.Service           = db.New()
	userRepo                          = repository.NewUserRepository(database)
	userService  domain.IUserService  = domain.NewUserService(userRepo)
	tokenService domain.ITokenService = domain.NewTokenService()
	hashService  domain.IHashService  = domain.NewHashService()
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

	r.GET("/register", func(c *gin.Context) {
		c.HTML(200, "register.html", nil)
	})

	r.GET("/login", func(c *gin.Context) {
		c.HTML(200, "login.html", nil)
	})

	// register
	r.POST("/api/register", s.Register)

	// login
	r.POST("/api/login", s.Login)

	// get user data
	r.GET("/api/me", s.GetUserData)

	return r
}
