// Package handlers is a nice package
package handlers

import (
	"log/slog"
	"net/http"
	"strings"

	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/odlev/subscriptions/pkg/myerrors"
	"github.com/odlev/subscriptions/pkg/sl"
	"github.com/odlev/subscriptions/internal/storage"
)

const DateLayout = "2006-01"

type DataWizard interface {
	CreateSubscription(sub *storage.SubscriptionR) (uuid.UUID, error)
	GetSubscription(id uuid.UUID) (*storage.Subscription, error)
	DeleteSubscription(id uuid.UUID) (string, error)
	UpdateSubscription(id uuid.UUID, req storage.UpdateSubscriptionRequest) error
	GetListSubscriptions(userID, name string) ([]storage.Subscription, error)
}

// CreateSubscription godoc
// @Summary Создать подписку
// @Description Добавляет новую подписку для пользователя. Поля user_id и end_date опциональны, если не указать user_id - сгенерируется автоматически, если не указать end_date - прибавиться + 1 год от начала подписки.
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param input body storage.SubscriptionCreateRequest true "Данные подписки"
// @Success 201 {object} map[string]interface{} "Успешное создание"
// @Failure 400 {object} map[string]interface{} "Ошибка валидации"
// @Failure 500 {object} map[string]interface{} "Внутрення ошибка сервера"
// @Router /new [post]
func CreateSubscription(log *slog.Logger, dataWizard DataWizard) gin.HandlerFunc {
	return func(c *gin.Context) {
		// const op = "handlers.subscriptions.CreateSubscription"
		// log = log.With(slog.String("operation:", op))

		var req storage.SubscriptionR

		
		if err := c.ShouldBindBodyWithJSON(&req); err != nil {
			log.Error("failed to decode request body", sl.Err(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to decode request body"})
			
			return
		}
		log.Info("request body was decoded", slog.Any("request", req))

		id, err := dataWizard.CreateSubscription(&req)
		if err != nil {
			log.Error("failed to create new subscription", sl.Err(err))

			if errors.Is(err, myerrors.ErrInvalidDateRange) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "end_date can not be earlier than start_date"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create new subscription"/*, "details": err.Error()*/})
			}
			return
		}

		log.Info("new subscription created!", slog.Any("susbcription id", id))
		c.JSON(http.StatusCreated, gin.H{"status": "Success", "ID": id})
	}
}
// GetSubscription godoc
// @Summary Получить подписку по ID
// @Description Возвращает подписку в формате, готовом для API (с преобразованными датами в необходимый формат)
// @Tags subscriptions
// @Produce json
// @Param id path string true "ID подписки" format(uuid) example(c9fd9538-e38c-429c-981b-f3ed34aee585)
// @Success 200 {object} storage.SubscriptionR "Успешно получено"
// @Failure 400 {object} map[string]any "Неверный UUID" example({"error": "failed to parse UUID"})
// @Failure 404 {object} map[string]any "Подписка не найдена" example({"subscription": "not found"})
// @Failure 500 {object} map[string]any "Внутренняя ошибка сервера" example({"error": "failed to get subscription, internal error"})
// @Router /get/{id} [get]
func GetSubscription(log *slog.Logger, dataWizard DataWizard) gin.HandlerFunc {
	return func(c *gin.Context) {
		// const op = "handlers.subscriptions.GetSubscriptionByID"
		// log = log.With(slog.String("operation:", op))

		strID := c.Param("id")
		log.Info("parameter 'id' successfully received", "id", strID)

		id, err := uuid.Parse(strID)
		if err != nil {
			log.Error("failed to parse UUID", sl.Err(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse UUID"})

			return
		}
		log.Info("UUID succesfully parsed", "UUID", id)

		var sub *storage.Subscription

		sub, err = dataWizard.GetSubscription(id)
		if err != nil {
			log.Error("failed to get", sl.Err(err))

			if errors.Is(err, myerrors.ErrNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"subscription": "not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get subscription, internal error"})
			}
			return
		}

		log.Info("Subscription successfully got", slog.Any("subscription", SubToFormatTime(sub)))
		c.JSON(http.StatusOK, gin.H{"subscription": SubToFormatTime(sub)})
	}
}
// DeleteSubscription godoc
// @Summary Удалить подписку
// @Description Удаляет подписку по ID и возвращает название удаленного сервиса
// @Tags subscriptions
// @Produce json
// @Param id path string true "ID подписки" format(uuid) example(550e8400-e29b-41d4-a716-446655440000)
// @Success 200 {object} map[string]interface{} "Успешное удаление" example({"status":"Success","deleted service":"Netflix"})
// @Failure 400 {object} map[string]interface{} "Неверный ID" example({"error":"failed to parse id","details":"invalid UUID format"})
// @Failure 404 {object} map[string]interface{} "Подписка не найдена" example({"error":"subscription not found"})
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера" example({"error":"internal server error"})
// @Router /delete/{id} [delete]
func DeleteSubscription(log *slog.Logger, dataWizard DataWizard) gin.HandlerFunc {
	return func(c *gin.Context) {
		// const op = "handlers.subscriptions.DeleteSubscription"
		// log = log.With(slog.String("operation:", op))

		strID := c.Param("id")
		//log.Info("parameter 'id' successfully received", "id", strID)

		id, err := uuid.Parse(strID)
		if err != nil {
			log.Error("error parsing id", sl.Err(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse id", "details": err.Error()})

			return
		}
		
		serviceName, err := dataWizard.DeleteSubscription(id)
		if err != nil {
			log.Error("failed to delete", sl.Err(err))

			if errors.Is(err, myerrors.ErrNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "subscription not found"})
			} else {
				c.JSON(http.StatusNotFound, gin.H{"error": "internal server error"})
			}
			return
		}
		log.Info("subscription succesfully deleted", "name", serviceName)

		c.JSON(http.StatusOK, gin.H{"status": "Success", "deleted service": serviceName})

	}
}

//UpdateSubscription godoc
// @Summary Обновить подписку
// @Description Обновляет любые поля записи о подписке ID и User_ID, сохраняет время последнего обновления в поле updated_at базы данных
// @Tags subscriptions
// @Accept json
// @Produce json
// @Par
// @Param id path string true "ID подписки" format(uuid) example(550e8400-e29b-41d4-a716-446655440000)
// @Param request body storage.UpdateSubscriptionRequest true "Данные для обновления"
// @Success 200 {object} map[string]interface{} "Успешно обновлено" example({"status": "success"})
// @Failure 400 {object} map[string]interface{} "Неверный ID" example({"error":"failed to parse id"})
// @Failure 400 {object} map[string]interface{} "Некорректный запрос" example({"error": "failed to decode request body"})
// @Failure 400 {object} map[string]interface{} "Некорретный диапазон дат" example({"error": "invalid request"})
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера" example({"error": "internal error"})
// @Router /update/{id} [patch]
func UpdateSubscription(log *slog.Logger, dataWizard DataWizard) gin.HandlerFunc {
	return func(c *gin.Context) {
		const op = "handlers.subscriptions.UpdateSubscription"

		strID := c.Param("id")
		// log.Info("parameter 'id' successfully received", "id", strID)

		id, err := uuid.Parse(strID)
		if err != nil {
			log.Error("error parsing id", sl.Err(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse id"})

			return
		}
		
		var req storage.UpdateSubscriptionRequest

		err = c.ShouldBindJSON(&req)
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to decode request body"})
			
			return
		}
		log.Info("request body was decoded", "request", req)

		err = dataWizard.UpdateSubscription(id, req)
		if err != nil {
			log.Error("update error", sl.Err(err))

			if errors.Is(err, myerrors.ErrInvalidDateRange) || strings.Contains(err.Error(), "invalid") {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			}
			return
		}
		log.Info("update succesfully completed")

		c.JSON(http.StatusOK, gin.H{"status": "success"})
	}
}

// GetListSubscriptions godoc
// @Summary Получить список подписок
// @Description Возвращает список подписок с возможностью фильтрации по user_id и названию сервиса. Если подписки не найдены, возвращает "not found".
// @Tags subscriptions
// @Produce json
// @Param user_id query string false "ID пользователя для фильтрации" format(uuid) example(550e8400-e29b-41d4-a716-446655440000)
// @Param service_name query string false "Название сервиса для фильтрации" example(Netflix)
// @Success 200 {object} map[string]interface{} "Успешный запрос" example({"subscriptions": [{"id": "550e8400-e29b-41d4-a716-446655440000", "service_name": "Netflix", ...}]})
// @Success 200 {object} map[string]interface{} "Если подписок нет" example({"subscriptions": "not found"})
// @Failure 400 {object} map[string]interface{} "Неверный user_id" example({"error": "invalid user_id"})
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера" example({"error": "internal server error"})
// @Router /get/list [get]
func GetListSubscriptions(log *slog.Logger, dataWizard DataWizard) gin.HandlerFunc {
	return func(c *gin.Context) {
		// const op = "handlers.subscriptions.GetAllSubscriptions"

		userID := c.Query("user_id")
		serviceName := c.Query("service_name")
		log.Info("query parameters received", "userID", userID, "service_name", serviceName)
		

		subs, err := dataWizard.GetListSubscriptions(userID, serviceName)
		if err != nil {
			log.Error("error getting list subscriptions", sl.Err(err))
			if strings.Contains(err.Error(), "failed parse id") {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})

			return
		}
		if subs != nil {
			log.Info("subscriptions found", "subscriptions", SubsToFormatTime(subs))
			c.JSON(http.StatusOK, gin.H{"subscriptions": SubsToFormatTime(subs)})
			} else {
				log.Info("subscriptions not found")
				c.JSON(http.StatusNotFound, gin.H{"subscriptions": "not found"})
			}
	}
}

func SubToFormatTime(sub *storage.Subscription) storage.SubscriptionR {
	return storage.SubscriptionR{
		ID:          sub.ID,
		ServiceName: sub.ServiceName,
		Price:       sub.Price,
		UserID:      sub.UserID,
		StartDate:   sub.StartDate.Format(DateLayout),
		EndDate:     sub.EndDate.Format(DateLayout),
	}													
}

func SubsToFormatTime(subs []storage.Subscription) []storage.SubscriptionR {
	result := make([]storage.SubscriptionR, len(subs))
	for i, sub := range subs {
		result[i] = storage.SubscriptionR{
			ID:          sub.ID,
			ServiceName: sub.ServiceName,
			Price:       sub.Price,
			UserID:      sub.UserID,
			StartDate:   sub.StartDate.Format(DateLayout),
			EndDate:     sub.EndDate.Format(DateLayout),
		}
	}
	return result
}
