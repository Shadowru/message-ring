package main

import (
	"core-go/handlers"
	"core-go/models"
	"core-go/workers"
	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"os"
)

func main() {
	// 1. Подключение к БД
	dsn := os.Getenv("DATABASE_DSN")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Миграции
	db.AutoMigrate(&models.Account{}, &models.Group{})

	// 2. Настройка Redis (Asynq)
	redisAddr := os.Getenv("REDIS_ADDR")
	redisOpt := asynq.RedisClientOpt{Addr: redisAddr}
	
	// Клиент для отправки задач
	queueClient := asynq.NewClient(redisOpt)
	defer queueClient.Close()

	// 3. Запуск Воркера (Server) в отдельной горутине
	queueServer := asynq.NewServer(redisOpt, asynq.Config{
		Concurrency: 10, // 10 параллельных потоков обработки
	})

	mux := asynq.NewServeMux()
	mux.HandleFunc(workers.TypeSendToN8N, workers.HandleSendToN8N(db))

	go func() {
		if err := queueServer.Run(mux); err != nil {
			log.Fatalf("could not run server: %v", err)
		}
	}()

	// 4. Запуск HTTP сервера (Gin)
	r := gin.Default()
	
	whHandler := &handlers.WebhookHandler{DB: db, QueueClient: queueClient}

	api := r.Group("/api")
	{
		// Админка (тут добавить Middleware JWT auth)
		// api.Use(authMiddleware)
		
		webhooks := r.Group("/webhooks")
		{
			webhooks.POST("/wa", whHandler.HandleWA)
			webhooks.POST("/tg", whHandler.HandleTG) // Реализовать
			webhooks.POST("/maks", whHandler.HandleMAKS) // Реализовать
		}
	}

	r.Run(":8080") // Слушаем порт внутри контейнера
}