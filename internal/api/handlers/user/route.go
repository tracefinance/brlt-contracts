package user

import (
	"vault0/internal/core/db"
	userService "vault0/internal/services/user"

	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all user-related routes and their dependencies
func SetupRoutes(router *gin.RouterGroup, db *db.DB) {
	// Create user repository
	userRepo := userService.NewSQLiteRepository(db)

	// Create user service
	userSvc := userService.NewService(userRepo)

	// Create user handler
	userHandler := NewHandler(userSvc)

	// Register user routes
	router.POST("/users", userHandler.CreateUser)
	router.GET("/users", userHandler.ListUsers)
	router.GET("/users/:id", userHandler.GetUser)
	router.PUT("/users/:id", userHandler.UpdateUser)
	router.DELETE("/users/:id", userHandler.DeleteUser)
}
