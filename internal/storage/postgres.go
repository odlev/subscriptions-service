// Package storage is a nice package
package storage

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"
	"github.com/jackc/pgx/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/odlev/subscriptions/internal/config"
	"github.com/odlev/subscriptions/pkg/myerrors"
)

const DateLayout = "2006-01"

type Storage struct {
	db *pgxpool.Pool
}

func InitPostgres(log *slog.Logger, cfg config.Config) (*Storage, error) {
	const op = "storage.postgres.InitPostgres"

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.SSLMode)

	var db *pgxpool.Pool
	var err error

	for range 5 {
		db, err = pgxpool.New(context.Background(), dsn)
		if err != nil {
			return nil, fmt.Errorf("%s, %w", op, err)
		}
		time.Sleep(2 * time.Second)
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
	updated_at TIMESTAMPTZ DEFAULT NOW());
	CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`)
	if err != nil {
		return nil, fmt.Errorf("%s: error creating table: %w", op, err)
	}

	log.Info("Connect with PostgreSQL established successfully")
	return &Storage{db: db}, nil
}

func (s *Storage) CreateSubscription(sub *SubscriptionR) (uuid.UUID, error) {
	const op = "storage.postgres.NewSubscription"

	var endDate time.Time 
	startDate, err := time.Parse(DateLayout, sub.StartDate)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%s: time parse error: %w", op, err)
	}
	if sub.EndDate == "" {
		endDate = startDate.AddDate(1, 0, 0)
	} else {
		endDate, err = time.Parse(DateLayout, sub.EndDate)
		if err != nil {
			return uuid.Nil, fmt.Errorf("%s: time parse error: %w", op, err)
		}
	}
	if endDate.Before(startDate) {
		return uuid.Nil, fmt.Errorf("%s: %w", op, myerrors.ErrInvalidDateRange)
	}

	var id uuid.UUID
	// если не передан user_id - не передаем его в бд и бд создает его по дефолту
	if sub.UserID == uuid.Nil { 
		query := `INSERT INTO subscriptions 
		(service_name, price, start_date, end_date)
		values ($1, $2, $3, $4) RETURNING id`

		err := s.db.QueryRow(context.Background(), query, sub.ServiceName, sub.Price, startDate, endDate).Scan(&id)
		if err != nil {
			return uuid.Nil, fmt.Errorf("%s: %w", op, err)
		}
	} else {
		query := `INSERT INTO subscriptions 
		(service_name, price, user_id, start_date, end_date)
		values ($1, $2, $3, $4, $5) RETURNING id`

		err := s.db.QueryRow(context.Background(), query, sub.ServiceName, sub.Price, sub.UserID, startDate, endDate).Scan(&id)
		if err != nil {
			return uuid.Nil, fmt.Errorf("%s: %w", op, err)
		}
	}
	
	return id, nil
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

func (s *Storage) UpdateSubscription(id uuid.UUID, req UpdateSubscriptionRequest) error {
	const op = "storage.postgres.UpdateSubscription"

	startDate, endDate, err := parseDates(req)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	
	query := `UPDATE subscriptions SET 
		service_name = COALESCE($1, service_name),
		price = COALESCE($2, price),
		start_date = COALESCE($3, start_date),
		end_date = COALESCE($4, end_date),
		updated_at = NOW()
	WHERE id = $5;`	
	
	_, err = s.db.Exec(context.Background(), query, req.ServiceName, req.Price, startDate, endDate, id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func parseDates(req UpdateSubscriptionRequest) (*time.Time, *time.Time, error) {
    const op = "storage.postgres.parseDates"

    var startDate, endDate *time.Time

    if req.StartDate != "" {
        parsed, err := time.Parse(DateLayout, req.StartDate)
        if err != nil {
            return nil, nil, fmt.Errorf("%s: invalid start date: %w", op, err)
        }
        startDate = &parsed
    }

    if req.EndDate != "" {
        parsed, err := time.Parse(DateLayout, req.EndDate)
        if err != nil {
            return nil, nil, fmt.Errorf("%s: invalid end date: %w", op, err)
        }
        endDate = &parsed
    } else if req.StartDate != "" && req.EndDate == "" {
        // Если EndDate не указан, но есть StartDate - ставим +1 год
        parsed := startDate.AddDate(1, 0, 0)
        endDate = &parsed
    }

	if startDate != nil && endDate != nil && endDate.Before(*startDate) {
		return nil, nil, fmt.Errorf("%s: %w", op, myerrors.ErrInvalidDateRange)
	}

    return startDate, endDate, nil
}

func (s *Storage) GetListSubscriptions(userID, name string) ([]Subscription, error) {
	const op = "storage.postgres.GetAllSubscriptions"

	query := `SELECT id, service_name, price, user_id, start_date, end_date
	FROM subscriptions WHERE 1 = 1`

	args := []any{}

	if userID != "" {
		if _, err := uuid.Parse(userID); err != nil {
			return nil, fmt.Errorf("%s: failed parse id: %w", op, err)
		}
		args = append(args, userID)
		query = query + fmt.Sprintf(" AND user_id = $%d", len(args))

	}
	if name != "" {
		args = append(args, name)
		query = query + fmt.Sprintf(" AND service_name = $%d", len(args))
	}

	rows, err := s.db.Query(context.Background(), query, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	defer rows.Close()

	var subs []Subscription
	
	var sub Subscription
	for rows.Next() {
		if err := rows.Scan(&sub.ID, &sub.ServiceName, &sub.Price, &sub.UserID, &sub.StartDate, &sub.EndDate); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		subs = append(subs, Subscription{
			ID: sub.ID,
			ServiceName: sub.ServiceName,
			Price: sub.Price,
			UserID: sub.UserID,
			StartDate: sub.StartDate,
			EndDate: sub.EndDate,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: rows iteration error: %w", op, err)
	}

	return subs, nil

}
