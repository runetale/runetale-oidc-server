package entity

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID         uint      `db:"id"`
	Username   string    `db:"username"`
	UserID     string    `db:"user_id"`
	TenantID   uint      `db:"tenant_id"`
	ProviderID string    `db:"provider_id"`
	Email      string    `db:"email"`
	Domain     string    `db:"domain"`
	Sub        string    `db:"sub"`
	Aud        string    `db:"aud"`
	Azp        string    `db:"azp"`
	Picture    string    `db:"picture"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
}

func NewUser(
	tenantid uint, username, provierid, email, domain, sub, aud, azp, picture string,
) *User {
	// memo: (shinta) is it safe?
	// what is the probability that uuid will be covered?
	uid := uuid.New().String()
	return &User{
		Username:   username,
		UserID:     uid,
		TenantID:   tenantid,
		ProviderID: provierid,
		Email:      email,
		Domain:     domain,
		Sub:        sub,
		Aud:        aud,
		Azp:        azp,
		Picture:    picture,
		CreatedAt:  time.Now().In(time.UTC),
		UpdatedAt:  time.Now().In(time.UTC),
	}
}
