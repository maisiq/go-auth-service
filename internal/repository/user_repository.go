package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jmoiron/sqlx"
	"github.com/maisiq/go-auth-service/internal/domain"
)

type UserRepository struct {
	client *sqlx.DB
}

func NewUserRepository(c *sqlx.DB) *UserRepository {
	return &UserRepository{
		client: c,
	}
}

func (r *UserRepository) Add(ctx context.Context, user domain.User) error {
	stmt := `INSERT INTO users(id, email, role, hashed_password, created_at, social_account, social_id, social_provider) 
			 VALUES(:id, :email, :role, :hashed_password, :created_at, :social_account, :social_id, :social_provider)`

	_, err := r.client.NamedExecContext(ctx, stmt,
		map[string]interface{}{
			"id":              user.ID,
			"email":           user.Email,
			"role":            user.Role,
			"hashed_password": user.HashedPassword,
			"created_at":      user.CreatedAT.Unix(),
			"social_account":  user.SocialAccount,
			"social_id":       user.SocialID,
			"social_provider": user.SocialProvider,
		},
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.UniqueViolation {
				return ErrAlreadyExists
			}
			return err
		}
		return err
	}
	return nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	query := "SELECT id, email, hashed_password FROM users WHERE email=$1"
	var user domain.User

	if err := r.client.QueryRowxContext(ctx, query, email).StructScan(&user); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, ErrNotFound
		}
		return domain.User{}, err
	}
	return user, nil

}

func (r *UserRepository) Logs(ctx context.Context, email string) ([]domain.UserLog, error) {
	var logs = make([]domain.UserLog, 0)

	query := "SELECT id, user_email, user_agent, ip, logged_at FROM user_logs"

	rows, err := r.client.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var (
			log       domain.UserLog
			LogID     uuid.UUID
			UserEmail string
			UserAgent string
			IP        string
			LoggedAt  int64
		)

		if err := rows.Scan(&LogID, &UserEmail, &UserAgent, &IP, &LoggedAt); err != nil {
			return nil, err
		}

		log = domain.UserLog{
			ID:        LogID,
			UserEmail: UserEmail,
			UserAgent: UserAgent,
			IP:        IP,
			LoggedAt:  time.Unix(LoggedAt, 0),
		}

		logs = append(logs, log)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return logs, nil
}

func (r *UserRepository) AddLog(ctx context.Context, log domain.UserLog) error {
	stmt := "INSERT INTO user_logs(id, user_email, user_agent, ip, logged_at) VALUES ($1, $2, $3, $4, $5)"
	_, err := r.client.ExecContext(ctx, stmt, log.ID, log.UserEmail, log.UserAgent, log.IP, log.LoggedAt.Unix())
	if err != nil {
		return err
	}
	return nil
}

func (r *UserRepository) Update(ctx context.Context, user domain.User) error {
	stmt := `UPDATE users
			 SET email = $1, hashed_password = $2
			 WHERE id = $3`
	_, err := r.client.ExecContext(ctx, stmt, user.Email, user.HashedPassword, user.ID)
	if err != nil {
		return err
	}
	return nil
}
