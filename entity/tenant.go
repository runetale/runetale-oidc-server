package entity

import (
	"time"

	"github.com/google/uuid"
)

type Tenant struct {
	ID        uint      `db:"id"`
	TenantID  string    `db:"tenant_id"`
	Domain    string    `db:"domain"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func NewTenant(
	domain string,
) *Tenant {
	uid := uuid.New().String()
	return &Tenant{
		TenantID:  uid,
		Domain:    domain,
		CreatedAt: time.Now().In(time.UTC),
		UpdatedAt: time.Now().In(time.UTC),
	}
}
