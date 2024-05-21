package pgdb

import (
	"bot_for_modeus/internal/model/pgmodel"
	"bot_for_modeus/internal/repo/pgerrs"
	"bot_for_modeus/pkg/postgres"
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	log "github.com/sirupsen/logrus"
)

const userPrefixLog = "/pgdb/user"

type UserRepo struct {
	*postgres.Postgres
}

func NewUserRepo(pg *postgres.Postgres) *UserRepo {
	return &UserRepo{pg}
}

func (r *UserRepo) CreateUser(ctx context.Context, u pgmodel.User) error {
	sql, args, _ := r.Builder.
		Insert("\"user\"").
		Columns("user_id", "full_name", "login", "password").
		Values(u.UserId, u.FullName, u.Login, u.Password).
		ToSql()
	if _, err := r.Pool.Exec(ctx, sql, args...); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return pgerrs.ErrAlreadyExists
			}
			log.Errorf("%s/CreateUser error exec stmt: %s", userPrefixLog, err)
			return err
		}
	}
	return nil
}

func (r *UserRepo) GetUserById(ctx context.Context, userId int64) (pgmodel.User, error) {
	sql, args, _ := r.Builder.
		Select("*").
		From("\"user\"").
		Where("user_id = ?", userId).
		ToSql()
	var u pgmodel.User

	err := r.Pool.QueryRow(ctx, sql, args...).Scan(
		&u.Id,
		&u.UserId,
		&u.FullName,
		&u.Login,
		&u.Password,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pgmodel.User{}, pgerrs.ErrNotFound
		}
		log.Errorf("%s/GetUserById error finding user: %s", userPrefixLog, err)
		return pgmodel.User{}, err
	}
	return u, nil
}

func (r *UserRepo) AddLoginPassword(ctx context.Context, userId int64, login, password string) error {
	sql, args, _ := r.Builder.
		Update("\"user\"").
		Set("login", login).
		Set("password", password).
		Where("user_id = ?", userId).
		ToSql()
	if _, err := r.Pool.Exec(ctx, sql, args...); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pgerrs.ErrNotFound
		}
		log.Errorf("%s/AddUserLoginPassword error exec stmt: %s", userPrefixLog, err)
		return err
	}
	return nil
}

func (r *UserRepo) DeleteLoginPassword(ctx context.Context, userId int64) error {
	sql, args, _ := r.Builder.
		Update("\"user\"").
		Set("login", "").
		Set("password", "").
		Where("user_id = ?", userId).
		ToSql()
	if _, err := r.Pool.Exec(ctx, sql, args...); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pgerrs.ErrNotFound
		}
		log.Errorf("%s/DeleteLoginPassword error exec stmt: %s", userPrefixLog, err)
		return err
	}
	return nil
}

func (r *UserRepo) UpdateFullName(ctx context.Context, userId int64, fullName string) error {
	sql, args, _ := r.Builder.
		Update("\"user\"").
		Set("full_name", fullName).
		Where("user_id = ?", userId).
		ToSql()
	if _, err := r.Pool.Exec(ctx, sql, args...); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pgerrs.ErrNotFound
		}
		log.Errorf("%s/UpdateUserFullName error exec stmt: %s", userPrefixLog, err)
		return err
	}
	return nil
}

func (r *UserRepo) DeleteUser(ctx context.Context, userId int64) error {
	sql, args, _ := r.Builder.
		Delete("\"user\"").
		Where("user_id = ?", userId).
		ToSql()
	if _, err := r.Pool.Exec(ctx, sql, args...); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pgerrs.ErrNotFound
		}
		log.Errorf("%s/DeleteUser error exec stmt: %s", userPrefixLog, err)
		return err
	}
	return nil
}
