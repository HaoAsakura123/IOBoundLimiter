package handlers

import (
	"ioboundlimiter/internal/auth"
	"ioboundlimiter/internal/storage"
	"ioboundlimiter/internal/util"
	"ioboundlimiter/internal/workers"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)


// Task represents a task structure
// @Description Модель задачи для создания
type Task struct {
    // Название задачи
    // @Example "Провести код-ревью"
    TaskName string `json:"taskname" binding:"required" example:"Какая то длинная io bound"`
}
// AddHandle godoc
//	@Summary		Добавить задачу
//	@Description	Добавляет новую задачу в систему обработки
//	@Tags			tasks
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			task	body		Task	true				"Данные задачи"
//	@Success		200		{object}	object	"{"status":"access","uuid":"string"}"
//	@Failure		400		{object}	object	"{"error":"should contain task"}"
//	@Failure		500		{object}	object	"{"error":"server is busy"}"
//	@Router			/api/add [post]
func AddHandle(c *gin.Context) {
	task := Task{}
	if err := c.ShouldBindJSON(&task); err != nil {
		log.Printf("ERROR: Validation error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "should contain task"})
		return
	}

	uuid, err := storage.AddToStorage(task.TaskName)
	if err != nil {
		log.Printf("ERROR: cannot create task: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot create task"})
		return
	}

	if err := workers.AddToChannel(uuid); err != nil {
		log.Printf("Server is busy: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server is busy"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "task is created",
		"uuid":   uuid,
	})

}

// TaskID represents task identifier
// @Description Идентификатор задачи в формате UUID
type TaskID struct {
    // Уникальный идентификатор задачи
    // @Example 6ba7b810-9dad-11d1-80b4-00c04fd430c8
    UUID string `json:"uuid" binding:"required" example:"6ba7b810-9dad-11d1-80b4-00c04fd430c8"`
}

// DeleteHandle godoc
//	@Summary		Удалить задачу
//	@Description	Удаляет задачу по UUID
//	@Tags			tasks
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			uuid	body		TaskID	true							"UUID задачи"
//	@Success		200		{object}	object	"{"status":"access","deleted	task":"string"}"
//	@Failure		400		{object}	object	"{"error":"string"}"
//	@Failure		404		{object}	object	"{"status":"Not	found	current	task"}"
//	@Router			/api/delete [delete]
func DeleteHandle(c *gin.Context) {
	uuid := TaskID{}

	if err := c.ShouldBindJSON(&uuid); err != nil {
		log.Printf("Bad request: should contain UUID: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request: should contain UUID"})
		return
	}

	if !storage.IsExists(uuid.UUID) {
		log.Printf("Task with this UUID: %s doesnt exists", uuid.UUID)
		c.JSON(http.StatusNoContent, gin.H{"status": "Not found current task"})
		return
	}

	if err := storage.DeleteTask(uuid.UUID); err != nil {
		log.Printf("Task with this UUID doesnt exists")
		c.JSON(http.StatusNoContent, gin.H{"error": "Task with this UUID doesnt exists"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":       "access",
		"deleted task": uuid.UUID,
	})

}

// GetHandle godoc
//	@Summary		Получить статус задачи
//	@Description	Возвращает текущий статус задачи
//	@Tags			tasks
//	@Accept			json
//	@Produce		json
//	@Param			uuid	body		TaskID	true	"UUID задачи"
//	@Success		200		{object}	object	"{"status":"access", "task name": "string", "createdAt": date, "current status": "string", "working time": "diff time" }"
//	@Success		204		{object}	object	"{"status":"not found task"}"
//	@Failure		400		{object}	object	"{"error":"Bad request: should contain UUID"}"
//	@Failure		404		{object}	object	"{"error":"Not	found	current	task"}"
//	@Router			/status [post]
func GetHandle(c *gin.Context) {
	uuid := TaskID{}

	if err := c.ShouldBindJSON(&uuid); err != nil {
		log.Printf("Bad request: should contain UUID: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request: should contain UUID"})
		return
	}

	if !storage.IsExists(uuid.UUID) {
		log.Printf("Task with this UUID: %s doesnt exists", uuid.UUID)
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found current task"})
		return
	}

	status, err := storage.GetResponse(uuid.UUID)

	if err != nil {
		log.Printf("Task with this UUID: %s doesnt exists: %v", uuid.UUID, err)
		c.JSON(http.StatusNoContent, gin.H{"status": "not found task"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":         "access",
		"task name":      status.Name,
		"created at":     status.DateOutput,
		"current status": status.CurStatus,
		"working time":   util.DifferenceTime(status.DateCreate),
	})
}

type RefreshRequest struct {
    // Refresh токен для обновления пары
    // @Example eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMTIzIn0.ABC123...
    Refresh string `json:"refresh" binding:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}
// RefreshHandler godoc
//	@Summary		Обновить токены
//	@Description	Обновляет пару access и refresh токенов
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			refresh	body		RefreshRequest	true	"Refresh токен"
//	@Success		200		{object}	object	"{"status":"string"}"
//	@Failure		400		{object}	object	"{"error":"string"}"
//	@Failure		401		{object}	object	"{"error":"invalid	authorization	format"}"
//	@Failure		500		{object}	object	"{"error":"string"}"
//	@Router			/api/refresh [post]
func RefreshHandler(c *gin.Context) {

	refresh := RefreshRequest{}
	if err := c.ShouldBindJSON(&refresh); err != nil {
		log.Printf("Should contain refresh token: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Should contain refresh token"})
		return
	}

	authHeader := c.GetHeader("Authorization")

	tokenParts := strings.Split(authHeader, " ")
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		log.Printf("invalid authorization format")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format"})
		return
	}
	accessToken := tokenParts[1]

	userID, err := auth.ValidateTokenPair(accessToken, refresh.Refresh)

	if err != nil {
		log.Printf("Cannot validate tokens: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot validate tokens"})
		return
	}

	// check in BD tokens
	if err := auth.CheckTokensExists(accessToken, refresh.Refresh); err != nil {
		log.Printf("Cannot find tokens in BD: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot find tokens in BD"})
		return
	}

	if err := auth.DeleteTokens(accessToken, refresh.Refresh); err != nil {
		log.Printf("Cannot delete tokens: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot delete tokens"})
		return
	}

	newAccess, newRefresh, err := auth.GenerateTokens(userID)

	if err != nil {
		log.Printf("Cannot generate tokens: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot generate tokens"})
		return
	}

	if err := auth.AddTokensToBd(newAccess, newRefresh); err != nil {
		log.Printf("Cannot add tokens to BD: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot add tokens to BD"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access":  newAccess,
		"refresh": newRefresh,
		"status":  "access",
	})
}

// RegisterHandler godoc
//	@Summary		Зарегистрировать пользователя
//	@Description	Создает нового пользователя и возвращает пару токенов
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	object	"{"status":"string"}"
//	@Failure		500	{object}	object	"{"error":"string"}"
//	@Router			/register [get]
func RegisterHandler(c *gin.Context) {
	UserID := uuid.New().String()

	access, refresh, err := auth.GenerateTokens(UserID)

	if err != nil {
		log.Printf("Cannot create jwt tokens: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot create tokens"})
		return
	}

	if err := auth.AddTokensToBd(access, refresh); err != nil {
		log.Printf("Cannot add tokens to BD: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot add tokens"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"userID":  UserID,
		"access":  access,
		"refresh": refresh,
		"status":  "access",
	})

}
