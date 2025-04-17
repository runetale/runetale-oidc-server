package repository

import (
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/lib/pq"
	"github.com/runetale/runetale-oidc-server/database"
	"github.com/runetale/runetale-oidc-server/entity"
)

type UserRepository struct {
	db *database.Postgres
}

type UserRepositoryImpl interface {
	Create(user *entity.User) error
}

func NewUserRepository(db *database.Postgres) UserRepositoryImpl {
	return &UserRepository{
		db: db,
	}
}

// user
func (r *UserRepository) Create(user *entity.User) error {
	err := r.db.Exec(
		`INSERT INTO users (
			username,
			user_id,
			tenant_id,
			provider_id,
			email,
			domain,
			sub,
			aud,
			azp,
			picture,
			created_at,
			updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		`,
		user.Username,
		user.UserID,
		user.TenantID,
		user.ProviderID,
		user.Email,
		user.Domain,
		user.Sub,
		user.Aud,
		user.Azp,
		user.Picture,
		user.CreatedAt,
		user.UpdatedAt,
	)
	pgErr, ok := err.(*pq.Error)
	if ok {
		if pgErr.Code == UniqueViolation {
			return database.ErrAlreadyExist
		}
	}
	if err != nil {
		return err
	}
	return nil
}
