package models

import (
	"gorm.io/gorm"
	"time"
)

type AccountType string

const (
	TypeWA   AccountType = "wa"
	TypeTG   AccountType = "tg"
	TypeMAKS AccountType = "maks"
)

type Account struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `json:"name"`
	Type      AccountType    `json:"type"`
	SessionID string         `json:"session_id"` // ID сессии в адаптере
	Groups    []*Group       `gorm:"many2many:account_groups;" json:"groups"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type Group struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	Name          string     `json:"name"`
	N8nWebhookURL string     `json:"n8n_webhook_url"`
	Accounts      []*Account `gorm:"many2many:account_groups;" json:"-"`
}

// Структура нормализованного сообщения для очереди
type NormalizedMessage struct {
	AccountID uint   `json:"account_id"`
	Source    string `json:"source"`
	SenderID  string `json:"sender_id"`
	SenderName string `json:"sender_name"`
	Text      string `json:"text"`
	RawData   string `json:"raw_data"` // JSON string
}