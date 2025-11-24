package handlers

import (
	"core-go/models"
	"core-go/workers"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"gorm.io/gorm"
	"net/http"
)

type WebhookHandler struct {
	DB          *gorm.DB
	QueueClient *asynq.Client
}

// Прием вебхука от WAHA
func (h *WebhookHandler) HandleWA(c *gin.Context) {
	// Структура входящего JSON от WAHA (упрощенно)
	var wahaPayload struct {
		Event   string `json:"event"`
		Session string `json:"session"`
		Payload struct {
			From string `json:"from"`
			Body string `json:"body"`
			NotifyName string `json:"notifyName"`
		} `json:"payload"`
	}

	if err := c.ShouldBindJSON(&wahaPayload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if wahaPayload.Event != "message" {
		c.JSON(http.StatusOK, gin.H{"status": "ignored"})
		return
	}

	// Ищем аккаунт по сессии
	var account models.Account
	if err := h.DB.Where("session_id = ?", wahaPayload.Session).First(&account).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Account not found for session"})
		return
	}

	// Нормализация
	msg := models.NormalizedMessage{
		AccountID:  account.ID,
		Source:     "wa",
		SenderID:   wahaPayload.Payload.From,
		SenderName: wahaPayload.Payload.NotifyName,
		Text:       wahaPayload.Payload.Body,
	}

	// Отправка в очередь Asynq
	payload, _ := json.Marshal(workers.N8NPayload{Message: msg})
	task := asynq.NewTask(workers.TypeSendToN8N, payload)
	
	if _, err := h.QueueClient.Enqueue(task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Queue error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "queued"})
}

// HandleTG и HandleMAKS реализуются аналогично