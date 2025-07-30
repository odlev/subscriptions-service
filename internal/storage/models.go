// Package storage is a nice package
package storage

import (
	"time"

	"github.com/google/uuid"
)

// Subscription - базовая модель БД
type Subscription struct {
	ID          uuid.UUID `json:"id,omitempty" example:"550e8400-e29b-41d4-a716-446655440200" format:"uuid"`
	ServiceName string    `json:"service_name" binding:"required" example:"Netflix"`
	Price       int       `json:"price" binding:"required,min=1" example:"500"`
	UserID      uuid.UUID `json:"user_id,omitempty" example:"550e8400-e29b-41d4-a716-446255440000" format:"uuid"`
	StartDate   time.Time `json:"start_date" binding:"required" example:"2025-07"`
	EndDate     time.Time `json:"end_date,omitempty" example:"2026-07"`
	//Description *string
}

// SubscriptionCreateRequest - структура для создания подписки (без ID)
type SubscriptionCreateRequest struct {
    ServiceName string    `json:"service_name" binding:"required" example:"Netflix" description:"Название сервиса (обязательное поле)"`
    Price       int       `json:"price" binding:"required,min=1" example:"500" description:"Стоимость подписки в рублях (обязательное поле)"`
    UserID      *uuid.UUID `json:"user_id,omitempty" example:"550e8400-e29b-41d4-a716-446655240000" format:"uuid" description:"ID пользователя (если не указан, будет сгенерирован автоматически)"`
    StartDate   string    `json:"start_date" binding:"required" example:"2025-07" description:"Дата начала в формате YYYY-MM (обязательное поле)"`
    EndDate     *string   `json:"end_date,omitempty" example:"2026-07" description:"Дата окончания в формате YYYY-MM (если не указана, будет start_date + 1 год)"`
}

// SubscriptionR - форматированная версия
type SubscriptionR struct {
	ID          uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440090" format:"uuid"`
	ServiceName string    `json:"service_name" example:"Netflix"`
	Price       int       `json:"price" example:"500"`
	UserID      uuid.UUID `json:"user_id" example:"550e8400-e29b-41d4-a716-446655240000" format:"uuid"`
	StartDate   string    `json:"start_date" example:"2025-07"`
	EndDate     string    `json:"end_date" example:"2026-07"`
}
// UpdateSubscriptionRequest - структура для обновления подписки
type UpdateSubscriptionRequest struct {
	ServiceName string `json:"service_name,omitempty" example:"Netflix"`
	Price       int    `json:"price,omitempty" example:"500"`
	StartDate   string `json:"start_date,omitempty" example:"2025-07"`
	EndDate     string `json:"end_date,omitempty" example:"2026-07"`
}
