package workers

import (
	"context"
	"core-go/models"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/hibiken/asynq"
	"gorm.io/gorm"
)

const TypeSendToN8N = "message:send_n8n"

// Структура задачи для Redis
type N8NPayload struct {
	Message models.NormalizedMessage
}

// Хендлер задачи (выполняется асинхронно)
func HandleSendToN8N(db *gorm.DB) asynq.HandlerFunc {
	return func(ctx context.Context, t *asynq.Task) error {
		var p N8NPayload
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return fmt.Errorf("json unmarshal failed: %v", err)
		}

		// 1. Загружаем группы аккаунта
		var account models.Account
		if err := db.Preload("Groups").First(&account, p.Message.AccountID).Error; err != nil {
			return fmt.Errorf("account not found: %v", err)
		}

		if len(account.Groups) == 0 {
			log.Printf("Account %d has no groups, skipping", account.ID)
			return nil
		}

		client := resty.New()

		// 2. Рассылаем по всем группам
		for _, group := range account.Groups {
			// Формируем финальный JSON для n8n
			finalBody := map[string]interface{}{
				"source": p.Message.Source,
				"group":  group.Name,
				"sender": map[string]string{
					"id":   p.Message.SenderID,
					"name": p.Message.SenderName,
				},
				"message":   p.Message.Text,
				"timestamp": time.Now().Unix(),
			}

			// Шлем POST
			_, err := client.R().
				SetBody(finalBody).
				Post(group.N8nWebhookURL)

			if err != nil {
				log.Printf("Failed to send to n8n (Group %s): %v", group.Name, err)
				// Возврат ошибки заставит asynq повторить попытку (retry)
				return err
			}
		}

		return nil
	}
}
