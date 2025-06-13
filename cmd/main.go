package main

import (
	"context"
	"ioboundlimiter/internal/handlers"
	"ioboundlimiter/internal/middleware"
	"ioboundlimiter/internal/workers"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/swaggo/gin-swagger" // gin-swagger middleware
	"github.com/swaggo/files" // swagger embed files
	_ "ioboundlimiter/docs" 

	"github.com/gin-gonic/gin"
)

//	@title			Tasks API
//	@version		1.0
//	@description	API для управления задачами с авторизацией и воркерами

//	@contact.name	API Support
//	@contact.email	support@tasksapi.com

//	@license.name	MIT
//	@license.url	https://opensource.org/licenses/MIT

//	@host						localhost:8080
//	@BasePath					/
//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
func main() {
	r := gin.Default()

	workers.InitWorkers()

	r.GET("/register", handlers.RegisterHandler)
	r.POST("/status", handlers.GetHandle)

	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware()) //Только авторизованные пользователи могут удалять и создавать таски
	{
		api.POST("/add", handlers.AddHandle)
		api.DELETE("/delete", handlers.DeleteHandle)

		api.POST("/refresh", handlers.RefreshHandler)
	}

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	// Graceful shutdown воркеров
	workers.Shutdown()
	log.Println("Server stopped gracefully")
}
