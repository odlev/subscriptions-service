// Package handlers is a nice package
package handlers

import (
	"log/slog"
	"net/http"

	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/odlev/subscriptions/internal/myerrors"
	"github.com/odlev/subscriptions/internal/sl"
	"github.com/odlev/subscriptions/internal/storage"
)

const DateLayout = "2006-01"

type DataWizard interface {
	NewSubscription(sub *storage.Subscription) (uuid.UUID, error)
	GetSubscription(id uuid.UUID) (*storage.Subscription, error)
	DeleteSubscription(id uuid.UUID) (string, error)
	UpdateSubscription(id uuid.UUID, req storage.UpdateSubscription) error
}

type SubscriptionResponse struct {
	ID          uuid.UUID `json:"id"`
	ServiceName string    `json:"service_name"`
	Price       int       `json:"price"`
	UserID      uuid.UUID `json:"user_id"`
	StartDate   string    `json:"start_date"`
	EndDate     string    `json:"end_date"`
}

func SubWithFormatTime(sub *storage.Subscription) SubscriptionResponse {
	return SubscriptionResponse{
		ID:          sub.ID,
		ServiceName: sub.ServiceName,
		Price:       sub.Price,
		UserID:      sub.UserID,
		StartDate:   sub.StartDate.Format(DateLayout),
		EndDate:     sub.EndDate.Format(DateLayout),
	}
}

func CreateSubscription(log *slog.Logger, dataWizard DataWizard) gin.HandlerFunc {
	return func(c *gin.Context) {
		const op = "handlers.subscriptions.CreateSubscription"

		log = log.With(slog.String("operation:", op))

		var req storage.Subscription

		
		if err := c.ShouldBindBodyWithJSON(&req); err != nil {
			log.Error("failed to decode request body", sl.Err(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to decode request body"})
			
			return
		}
		log.Info("request body was decoded", slog.Any("request", req))
		reqFormatTime := SubWithFormatTime(&req)

		id, err := dataWizard.NewSubscription(&reqFormatTime)
		if err != nil {
			log.Error("failed to create new subscription", sl.Err(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create new subscription"})

			return
		}
		log.Info("new subscription created!", slog.Any("susbcription id", id))
		c.JSON(http.StatusCreated, gin.H{"status": "Success", "ID": id})
	}
}

func GetSubscription(log *slog.Logger, dataWizard DataWizard) gin.HandlerFunc {
	return func(c *gin.Context) {
		const op = "handlers.subscriptions.GetSubscriptionByID"

		//log = log.With(slog.String("operation:", op))

		strID := c.Param("id")
		log.Info("parameter 'id' successfully received", "id", strID)

		id, err := uuid.Parse(strID)
		if err != nil {
			log.Error("failed to parse UUID", sl.Err(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse UUID"})

			return
		}
		log.Info("UUID succesfully parsed", "UUID", id)

		var req *storage.Subscription

		req, err = dataWizard.GetSubscription(id)
		if err != nil {
			log.Error("failed to get", sl.Err(err))

			if errors.Is(err, myerrors.ErrNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "subscription not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get subscription by id"})
			}
			return
		}

		log.Info("Subscription successfully got", slog.Any("subscription", SubWithFormatTime(req)))
		c.JSON(http.StatusOK, gin.H{"subscription": SubWithFormatTime(req)})
	}
}

func DeleteSubscription(log *slog.Logger, dataWizard DataWizard) gin.HandlerFunc {
	return func(c *gin.Context) {
		const op = "handlers.subscriptions.DeleteSubscription"

		//log = log.With(slog.String("operation:", op))

		strID := c.Param("id")
		log.Info("parameter 'id' successfully received", "id", strID)

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
