package crypto

import (
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

type CustomcClaimsWithSub struct {
	Sub string `json:"sub"`
	*jwt.StandardClaims
}

type JwtIssuer interface {
	CreateJwtTokenWithSub(sub string) (string, error)
	GetCustomClamis(token, jwtSecret string) (*CustomcClaimsWithSub, string, error)
	GetJwtSecret() string
}

type Jwt struct {
	JwtSecret   string
	JwtAudience string
	JwtIss      string
}

func NewJwtIssuer(jwtSecret, jwtAud, jwtIss string) JwtIssuer {
	return &Jwt{
		JwtSecret:   jwtSecret,
		JwtAudience: jwtAud,
		JwtIss:      jwtIss,
	}
}

func (j *Jwt) CreateJwtTokenWithSub(sub string) (string, error) {
	var (
		now = time.Now().Unix()
		exp = time.Now().Add((time.Hour * 24) * 14).Unix() // 2weeks
		sb  strings.Builder
	)

	claims := CustomcClaimsWithSub{
		sub,
		&jwt.StandardClaims{
			Audience:  j.JwtAudience,
			ExpiresAt: exp,
			Id:        uuid.New().String(),
			Issuer:    j.JwtIss,
			IssuedAt:  now,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString([]byte(j.JwtSecret))
	if err != nil {
		return "", err
	}

	sb.WriteString(ss)

	return sb.String(), nil
}

func (j *Jwt) GetCustomClamis(token, jwtSecret string) (*CustomcClaimsWithSub, string, error) {
	t, err := jwt.ParseWithClaims(token, &CustomcClaimsWithSub{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})

	if claims, ok := t.Claims.(*CustomcClaimsWithSub); ok && t.Valid {
		return claims, token, nil
	}

	return nil, "", err
}

func (j *Jwt) GetJwtSecret() string {
	return j.JwtSecret
}
