package handlers

import (
	"core-go/models"
	"core-go/workers"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"gorm.io/gorm"
)

type WebhookHandler struct {
	DB          *gorm.DB
	QueueClient *asynq.Client
}

// === WhatsApp (WAHA) ===
func (h *WebhookHandler) HandleWA(c *gin.Context) {
	var wahaPayload struct {
		Event   string `json:"event"`
		Session string `json:"session"`
		Payload struct {
			From       string `json:"from"`
			Body       string `json:"body"`
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

	// Ищем аккаунт
	var account models.Account
	if err := h.DB.Where("session_id = ?", wahaPayload.Session).First(&account).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Account not found for session"})
		return
	}

	// Формируем сообщение
	msg := models.NormalizedMessage{
		AccountID:  account.ID,
		Source:     "wa",
		SenderID:   wahaPayload.Payload.From,
		SenderName: wahaPayload.Payload.NotifyName,
		Text:       wahaPayload.Payload.Body,
	}

	h.enqueueMessage(c, msg)
}

// === Telegram ===
func (h *WebhookHandler) HandleTG(c *gin.Context) {
	var tgPayload struct {
		SessionID string `json:"sessionId"`
		Event     string `json:"event"`
		Data      struct {
			SenderID   string `json:"senderId"`
			SenderName string `json:"senderName"`
			Text       string `json:"text"`
		} `json:"data"`
	}

	if err := c.ShouldBindJSON(&tgPayload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var account models.Account
	// Предполагаем, что в БД session_id совпадает с тем, что шлет адаптер
	if err := h.DB.Where("session_id = ?", tgPayload.SessionID).First(&account).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		return
	}

	msg := models.NormalizedMessage{
		AccountID:  account.ID,
		Source:     "tg",
		SenderID:   tgPayload.Data.SenderID,
		SenderName: tgPayload.Data.SenderName,
		Text:       tgPayload.Data.Text,
	}

	h.enqueueMessage(c, msg)
}

// === MAKS ===
func (h *WebhookHandler) HandleMAKS(c *gin.Context) {
	// Примерная структура (зависит от реального API МАКС)
	var maksPayload struct {
		Token   string `json:"token"`
		Sender  string `json:"sender"`
		Message string `json:"message"`
	}

	if err := c.ShouldBindJSON(&maksPayload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Для МАКС можно искать аккаунт по типу, так как там часто один токен на систему
	var account models.Account
	if err := h.DB.Where("type = ?", "maks").First(&account).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "MAKS account not configured"})
		return
	}

	msg := models.NormalizedMessage{
		AccountID:  account.ID,
		Source:     "maks",
		SenderID:   maksPayload.Sender,
		SenderName: "Unknown", // Или достать из БД контактов
		Text:       maksPayload.Message,
	}

	h.enqueueMessage(c, msg)
}

// Вспомогательная функция отправки в очередь
func (h *WebhookHandler) enqueueMessage(c *gin.Context, msg models.NormalizedMessage) {
	payload, _ := json.Marshal(workers.N8NPayload{Message: msg})
	task := asynq.NewTask(workers.TypeSendToN8N, payload)

	if _, err := h.QueueClient.Enqueue(task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Queue error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "queued"})
}
