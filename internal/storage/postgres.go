// Package storage is a nice package
package storage

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/odlev/subscriptions/internal/config"
	"github.com/odlev/subscriptions/internal/myerrors"
)


type Storage struct {
	db *pgxpool.Pool
}

type Subscription struct {
	ID          uuid.UUID `json:"id,omitempty"`
	ServiceName string    `json:"service_name" binding:"required"`
	Price       int       `json:"price" binding:"required,min=1"`
	UserID      uuid.UUID `json:"user_id,omitempty"`
	StartDate   time.Time `json:"start_date" binding:"required"`
	EndDate     time.Time `json:"end_date,omitempty"`
	//Description *string
}

type UpdateSubscription struct {
	ServiceName string `json:"service_name,omitempty"`
	Price int `json:"price,omitempty"`
	StartDate time.Time `json:"start_date,omitempty"`
	EndDate time.Time `json:"end_date,omitempty"`
}

func InitPostgres(log *slog.Logger, cfg config.Config) (*Storage, error) {
	const op = "storage.postgres.InitPostgres"

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.SSLMode)

	db, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, fmt.Errorf("%s, %w", op, err)
	}

	if err := db.Ping(context.Background()); err != nil {
		db.Close()
		return nil, fmt.Errorf("%s: ping failed: %w", op, err)
	}

	_, err = db.Exec(context.Background(), `CREATE TABLE IF NOT EXISTS subscriptions (
	id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
	service_name TEXT NOT NULL,
	price DECIMAL NOT NULL CHECK (price > 0),
	user_id UUID NOT NULL DEFAULT uuid_generate_v4(),
	start_date DATE NOT NULL,
	end_date DATE,
	updated_at TIMESTAMPTZ DEFAULT NOW());`)
	if err != nil {
		return nil, fmt.Errorf("%s: error creating table: %w", op, err)
	}

	log.Info("Connect with PostgreSQL established successfully")
	return &Storage{db: db}, nil
}

func (s *Storage) NewSubscription(sub *Subscription) (uuid.UUID, error) {
	const op = "storage.postgres.NewSubscription"

	if sub.EndDate.IsZero() {
		sub.EndDate = sub.StartDate.AddDate(1, 0, 0)
	}
	var id uuid.UUID
	// если не передан user_id - не передаем его в бд и бд создает его по дефолту
	if sub.UserID == uuid.Nil { 
		query := `INSERT INTO subscriptions 
		(service_name, price, start_date, end_date)
		values ($1, $2, $3, $4) RETURNING id`

		err := s.db.QueryRow(context.Background(), query, sub.ServiceName, sub.Price, sub.StartDate, sub.EndDate).Scan(&id)
		if err != nil {
			return uuid.Nil, fmt.Errorf("%s: %w", op, err)
		}
	} else {
		query := `INSERT INTO subscriptions 
		(service_name, price, user_id, start_date, end_date)
		values ($1, $2, $3, $4, $5) RETURNING id`

		err := s.db.QueryRow(context.Background(), query, sub.ServiceName, sub.Price, sub.UserID, sub.StartDate, sub.EndDate).Scan(&id)
		if err != nil {
			return uuid.Nil, fmt.Errorf("%s: %w", op, err)
		}
	}

	if err := sub.Validate(); err != nil {
		return uuid.Nil, fmt.Errorf("%s: invalid subscription: %w", op, err)
	}
	//slog.Info("user_id received", slog.Any("user_id", sub.UserID))
	
	return id, nil
}

func (s *Subscription) Validate() error {
	realStartTime := time.Time(s.StartDate)
	//realEndTime := time.Time(s.EndDate)

	if s.ServiceName == "" {
		return errors.New("service_name is required")
	}
	// пусть и такой сценарий проверяется в бд
	if s.Price <= 0 {
		return errors.New("price must be positive")
	}
	if realStartTime.IsZero() {
		return errors.New("start_date is required")
	}
	/* if realStartTime.After(realEndTime) {
		return errors.New("start_date must be before end_date")
	} */

	return nil
}

//GetSubscription позволяет получить все поля таблицы для одного uuid
func (s *Storage) GetSubscription(id uuid.UUID) (*Subscription, error){
	const op = "storage.postgres.GetSubscriptionByID"

	query := `SELECT id, service_name, price, user_id, start_date, end_date FROM subscriptions
	WHERE id = $1`

	var sub Subscription
	

	err := s.db.QueryRow(context.Background(), query, id).Scan(
		&sub.ID, &sub.ServiceName, &sub.Price, &sub.UserID, &sub.StartDate, &sub.EndDate,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, myerrors.ErrNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &sub, nil
}

func (s *Storage) DeleteSubscription(id uuid.UUID) (string, error) {
	const op = "storage.postgres.DeleteSusbcription"

	var serviceName string 

	queryGetName := `SELECT service_name FROM subscriptions WHERE id = $1;`

	err := s.db.QueryRow(context.Background(), queryGetName, id).Scan(&serviceName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", fmt.Errorf("%s: %w", op, myerrors.ErrNotFound)
		}
		return "", fmt.Errorf("%s: failed to delete: %w", op, err)
	}

	query := `DELETE FROM subscriptions WHERE id = $1;`

	_, err = s.db.Exec(context.Background(), query, id)
	if err != nil {
		return "", fmt.Errorf("%s, %w", op, err)
	}
	return serviceName, nil
}

func (s *Storage) UpdateSubscription(id uuid.UUID, req UpdateSubscription) error {
	const op = "storage.postgres.UpdateSubscription"

	query := `UPDATE subscriptions SET 
	`	
	_ = query
	return nil
}
