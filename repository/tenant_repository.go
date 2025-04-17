package repository

import (
	"database/sql"

	"github.com/runetale/runetale-oidc-server/database"
	"github.com/runetale/runetale-oidc-server/entity"
)

type TenantRepository struct {
	db *database.Postgres
}

type TenantRepositoryImpl interface {
	Create(tenant *entity.Tenant) error
	FindByDomain(domain string) (*entity.Tenant, error)
}

func NewTenantRepository(db *database.Postgres) TenantRepositoryImpl {
	return &TenantRepository{
		db: db,
	}
}

func (r *TenantRepository) Create(tenant *entity.Tenant) error {
	err := r.db.Exec(
		`INSERT INTO tenants (
			tenant_id,
			domain,
			created_at,
			updated_at
		) VALUES ($1, $2, $3, $4)
		`,
		tenant.TenantID,
		tenant.Domain,
		tenant.CreatedAt,
		tenant.UpdatedAt,
	)
	if err != nil {
		return err
	}
	return nil
}

func (r *TenantRepository) FindByDomain(domain string) (*entity.Tenant, error) {
	var (
		t entity.Tenant
	)

	row := r.db.QueryRow(
		`
			SELECT *
			FROM tenants
			WHERE
  				domain = $1
			LIMIT 1
		`, domain)

	err := row.Scan(
		&t.ID,
		&t.TenantID,
		&t.Domain,
		&t.CreatedAt,
		&t.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, database.ErrNoRows
		}
		return nil, err
	}

	return &t, nil
}
